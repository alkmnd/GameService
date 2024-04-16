package game

import (
	"GameService/connectteam_service/models"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ledongthuc/goterators"
	"log"
	"net/http"
	//"encoding/binary"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	// The actual websocket connection.
	ID       uuid.UUID
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
	games    map[*Game]bool
	User     *User
}

func newClient(conn *websocket.Conn, wsServer *WsServer, user User) *Client {
	return &Client{
		ID:       uuid.New(),
		User:     &user,
		conn:     conn,
		wsServer: wsServer,
		games:    make(map[*Game]bool),
		send:     make(chan []byte, 256),
	}

}

func (client *Client) GetName() string {
	return client.User.Name
}

func (client *Client) GetId() uuid.UUID {
	return client.ID
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for game := range client.games {
		game.unregister <- client
	}
}

// ServeWs handles websocket requests from Clients requests.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	//name, ok := r.URL.Query()["name"]
	//
	//if !ok || len(name[0]) < 1 {
	//	log.Println("Url Param 'name' is missing")
	//	return
	//}

	token, ok := r.URL.Query()["token"]

	if !ok || len(token[0]) < 1 {
		log.Println("Url Param 'token' is missing")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// если токен пустой, используем client id

	id, _, err := wsServer.service.ParseToken(token[0])
	if err != nil {
		log.Println(err)
		return
	}

	user, err := wsServer.service.GetUserById(id)

	// TODO: проверить есть ли уже клиент на сервере

	client := newClient(conn, wsServer, User{
		Id:   id,
		Name: fmt.Sprintf("%s %s", user.FirstName, user.SecondName),
	})

	go client.writePump()
	go client.readPump()

	wsServer.register <- client
}

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less than pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		client.handleNewMessage(jsonMessage)
	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func (client *Client) handleNewMessage(jsonMessage []byte) {

	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("handleNewMessage error on unmarshal JSON message %s", err)
		return
	}

	// Attach the client object as the sender of the message.
	message.Sender = client.User

	switch message.Action {

	case SendMessageAction:
		if game := client.wsServer.findGame(message.Target.ID); game != nil {
			game.broadcast <- &message
		}
	case JoinGameAction:
		gameId := message.Target.ID

		game := client.wsServer.findGame(gameId)
		if game == nil {
			return
		}
		log.Println("joinGameAction")
		client.handleJoinGameMessage(message)
	case StartGameAction:
		log.Println("startGameAction")
		client.handleStartGameMessage(message)
	case LeaveGameAction:
		client.handleLeaveGameMessage(message)

	case SelectTopicAction:
		log.Println("selectTopicAction")
		client.handleSelectTopicGameMessage(message)

	case StartRoundAction:
		client.handleStartRoundMessage(message)

	case UserStartAnswerAction:
		client.handleUserStartAnswerMessage(message)
	case UserEndAnswerAction:
		client.handleUserEndAnswerMessage(message)
	case RateAction:
		client.handleRateMessage(message)
	case StartStageAction:
		client.handleStartStageMessage(message)
	case EndGameAction:
		client.handleEndGameMessage(message)
	}

}

type ratePayload struct {
	Value  int `json:"value"`
	UserId int `json:"user_id"`
}

func (client *Client) handleEndGameMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	game.Status = "ended"

	_ = client.wsServer.service.SaveResults(gameId, game.Results)
	game.broadcast <- &Message{
		Action: EndGameAction,
		Target: game,
	}

	return

}

func (client *Client) handleStartStageMessage(message Message) {

	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if len(goterators.Filter(game.Topics, func(item Topic) bool {
		return item.Used == false
	})) == 0 && len(game.Round.UsersQuestions) == 0 {
		game.Status = "ended"
		_ = client.wsServer.service.SaveResults(gameId, game.Results)
		game.broadcast <- &Message{
			Action: EndGameAction,
			Target: game,
		}

		return
	}
	if len(game.Round.UsersQuestions) == 0 {
		game.RoundsLeft = append(game.RoundsLeft, game.Round)
		game.broadcast <- &Message{
			Action: RoundEndAction,
			Target: game,
		}
		return
	}

	//if len(game.Users) == len(game.Round.UsersQuestionsLeft) {
	//	game.RoundsLeft = append(game.RoundsLeft, game.Round)
	//	game.broadcast <- &Message{
	//		Action: RoundEndAction,
	//		Target: game,
	//	}
	//	return
	//}
	var respondent *UsersQuestions
	var err error
	if len(game.Round.UsersQuestionsLeft) == 0 {
		respondent, _, err = goterators.Find(game.Round.UsersQuestions, func(item *UsersQuestions) bool {
			return item.User.Id == game.Creator
		})

		if err != nil {
			return
		}
	} else {
		respondent = game.Round.UsersQuestions[0]
	}

	payload, _ := json.Marshal(respondent)

	game.broadcast <- &Message{
		Action:  StartStageAction,
		Target:  game,
		Payload: payload,
		Sender:  message.Sender,
	}

}
func (client *Client) handleRateMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if game.Results == nil {
		game.Results = make(map[int]int)
	}

	var ratePayload ratePayload
	err := json.Unmarshal(message.Payload, &ratePayload)
	if err != nil {
		return
	}

	usersQuestions := goterators.Filter(game.Round.UsersQuestions, func(item *UsersQuestions) bool {
		return item.User.Id == ratePayload.UserId
	})[0]

	if client.User.Id == ratePayload.UserId {
		return
	}

	message.Target = game

	usersQuestions.Rates[client.User.Id] = ratePayload.Value
	_, ok := game.Results[ratePayload.UserId]
	if !ok {
		game.Results[ratePayload.UserId] = ratePayload.Value
	} else {
		game.Results[ratePayload.UserId] += ratePayload.Value
	}

	if len(usersQuestions.Rates) == len(game.Users)-1 {
		game.Round.UsersQuestionsLeft = append(game.Round.UsersQuestionsLeft, usersQuestions)
		game.Round.UsersQuestions = goterators.Filter(game.Round.UsersQuestions, func(item *UsersQuestions) bool {
			return item.User.Id != usersQuestions.User.Id
		})
		game.broadcast <- &Message{
			Action: EndRateAction,
			Target: game,
		}
		return
	}

	game.broadcast <- &message
}

func (client *Client) handleUserEndAnswerMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	stage, _, err := goterators.Find(game.Round.UsersQuestions, func(item *UsersQuestions) bool {
		return item.User.Id == client.User.Id
	})

	if err != nil {
		return
	}

	for i := range game.Users {
		if game.Users[i].Id != client.User.Id {
			stage.Rates[game.Users[i].Id] = 0
		}
	}

	//if len(game.Round.UsersQuestions) == 1 {
	//	game.broadcast <- &Message{
	//		Action: RoundEndAction,
	//		Target: game,
	//	}
	//	return
	//}
	game.broadcast <- &message
}

func (client *Client) handleUserStartAnswerMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)
	game.broadcast <- &message
}

func (client *Client) handleStartRoundMessage(message Message) {
	var messageSend *Message
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if len(goterators.Filter(game.Topics, func(item Topic) bool {
		return item.Used == false
	})) == 0 {
		game.broadcast <- &Message{
			Action: EndGameAction,
			Target: game,
		}
		return
	}

	game.Round = &Round{
		Topic:              &Topic{},
		UsersQuestions:     make([]*UsersQuestions, 0),
		UsersQuestionsLeft: make([]*UsersQuestions, 0),
	}
	// TODO: check if user is creator
	var topic *Topic
	if err := json.Unmarshal(message.Payload, &topic); err != nil {
		// error
		return
	}
	var topicFound *Topic

	for i := range game.Topics {
		if game.Topics[i].Id == topic.Id {
			topicFound = &game.Topics[i]
			break
		}
	}
	if topicFound == nil {
		// error
		return
	}
	if len(game.Users) != len(topicFound.Questions) {
		// error
		return
	}

	cnt := 1
	for i := range game.Users {
		game.Round.UsersQuestions = append(game.Round.UsersQuestions, &UsersQuestions{
			User: game.Users[i],
			// TODO: choose random elem
			Question: topicFound.Questions[i],
			Number:   cnt,
			Rates:    make(map[int]int),
		})
		cnt++

	}

	topicFound.Used = true
	game.Round.Topic = topicFound

	messageSend = &Message{
		Action: StartRoundAction,
		Target: game,
		//Payload: payload,
	}

	game.broadcast <- messageSend
}

func (client *Client) handleStartGameMessage(message Message) {
	//  меняем статус,
	var messageSend Message
	gameId := message.Target.ID

	game := client.wsServer.findGame(gameId)
	game.Status = "in_progress"

	err := client.wsServer.service.StartGame(gameId)
	if err != nil {
		log.Println("handleStartGameMessage unknown game")
		return
	}
	if len(game.Topics) == 0 {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload, err = json.Marshal("number of rounds is 0")
		client.send <- messageError.encode()
		log.Printf("handleStartGameMessage %s", messageError.Payload)
		return
	}
	questions := make(map[int][]models.Question)

	for i, _ := range game.Topics {
		questions[game.Topics[i].Id], err = client.wsServer.service.GetRandWithLimit(game.Topics[i].Id, len(game.Users))
		if err != nil {
			continue
		}
		game.Topics[i].Questions = make([]string, len(game.Users))
		for j := 0; j < len(game.Users); j++ {
			if questions[game.Topics[i].Id] != nil {
				game.Topics[i].Questions[j] = questions[game.Topics[i].Id][j].Content
			}
		}
	}
	//game.RoundsLeft = len(game.Clients)
	messageSend.Action = StartGameAction
	messageSend.Target = game

	//bytes, err := json.Marshal(game)
	//log.Println(json.Unmarshal(bytes, &game.Topics))
	if err != nil {
		return
	}
	game.broadcast <- &messageSend
}

func (client *Client) handleSelectTopicGameMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if message.Sender.Id != game.Creator {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload, _ = json.Marshal("message.Sender.Id != game.Creator")
		client.send <- messageError.encode()
		log.Printf("handleSelectTopicMessage %s", messageError.Payload)
		return
	}

	if err := json.Unmarshal(message.Payload, &game.Topics); err != nil {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload, _ = json.Marshal("incorrect payload")
		log.Printf("handleSelectTopicMessage %s", err.Error())
		client.send <- messageError.encode()
		log.Printf("handleSelectTopicMessage %s", messageError.Payload)
		return
	}

	for i := range game.Topics {
		log.Println("get topics from db")
		topic, _ := client.wsServer.service.GetTopic(game.Topics[i].Id)
		game.Topics[i].Title = topic.Title
	}

	println("handleSelectTopicGameMessage")
	message.Target = game

	// TODO: добавить title у топиков
	game.broadcast <- &message
}

func (client *Client) handleJoinGameMessage(message Message) {
	// TODO:  проверить есть ли этот пользователь в игре по id клиента

	gameId := message.Target.ID

	game := client.wsServer.findGame(gameId)

	//var messageSuccess Message
	//for i := range game.Users {
	//	if game.Users[i].Id == client.User.Id {
	//		return
	//	}
	//}

	if len(game.Users) == game.MaxSize {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload, _ = json.Marshal("maximum number of participants")
		client.send <- messageError.encode()
		return
	}

	//client.send
	client.games[game] = true

	game.register <- client
}

func (client *Client) handleLeaveGameMessage(message Message) {
	game := client.wsServer.findGame(message.Target.ID)
	if _, ok := client.games[game]; ok {
		delete(client.games, game)
	}

	game.unregister <- client
	var messageSend Message
	messageSend.Action = UserLeftAction
	messageSend.Sender = client.User
	messageSend.Target = game
	game.broadcast <- &messageSend
}
