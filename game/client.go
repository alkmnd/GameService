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
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Client struct {
	// The actual websocket connection.
	ID       uuid.UUID
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
	User     *User
}

func newClient(conn *websocket.Conn, wsServer *WsServer, user User) *Client {
	return &Client{
		ID:       uuid.New(),
		User:     &user,
		conn:     conn,
		wsServer: wsServer,
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

	id, access, err := wsServer.service.ParseToken(token[0])
	if err != nil {
		return
	}

	if access != "user" {
		return
	}

	user, err := wsServer.service.GetUserById(id)

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
		client.handleJoinGameMessage(message)
	case StartGameAction:
		client.handleStartGameMessage(message)
	case LeaveGameAction:
		client.handleLeaveGameMessage(message)
	case SelectTopicAction:
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
	Value  int       `json:"value"`
	UserId uuid.UUID `json:"user_id"`
	Tags   []Tag     `json:"tags"`
}

func (client *Client) handleEndGameMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	game.Status = "ended"

	game.endGame()
	results := make(map[uuid.UUID]models.Rates)
	for i, v := range game.Results {
		tags := make([]uuid.UUID, len(v.Tags))
		for k := range v.Tags {
			tags[k] = v.Tags[k]
		}
		results[i] = models.Rates{
			Value: v.Value,
			Tags:  tags,
		}
	}
	_ = client.wsServer.service.SaveResults(gameId, results)
	game.broadcast <- &Message{
		Action: EndGameAction,
		Target: game,
	}

	return

}

func (client *Client) handleStartStageMessage(message Message) {

	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if game.getCreator() != client.User.Id {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 8
		client.send <- messageError.encode()
		return
	}

	if len(goterators.Filter(game.Topics, func(item Topic) bool {
		return item.Used == false
	})) == 0 && len(game.Round.UsersQuestions) == 0 {
		game.endGame()
		results := make(map[uuid.UUID]models.Rates)
		for i, v := range game.Results {
			tags := make([]uuid.UUID, len(v.Tags))
			for k := range v.Tags {
				tags[k] = v.Tags[k]
			}
			results[i] = models.Rates{
				Value: v.Value,
				Tags:  tags,
			}
		}
		_ = client.wsServer.service.SaveResults(gameId, results)
		_ = client.wsServer.service.EndGame(gameId)
		game.broadcast <- &Message{
			Action: EndGameAction,
			Target: game,
		}

		return
	}
	if len(game.Round.UsersQuestions) == 0 {
		game.broadcast <- &Message{
			Action: RoundEndAction,
			Target: game,
		}
		return
	}

	var respondent *UserQuestion
	if len(game.Round.UsersQuestions) > 0 {
		respondent = game.Round.UsersQuestions[0]
	}
	payload := respondent

	game.broadcast <- &Message{
		Action:  StartStageAction,
		Target:  message.Target,
		Payload: payload,
		Sender:  message.Sender,
	}

}
func (client *Client) handleRateMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if game.Results == nil {
		game.Results = make(map[uuid.UUID]*Rates)
	}

	var rate ratePayload
	if payload, ok := message.Payload.(ratePayload); !ok {
		rate = payload
	} else {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 11
		client.send <- messageError.encode()
		return
	}

	usersQuestions := goterators.Filter(game.Round.UsersQuestions, func(item *UserQuestion) bool {
		return item.User == rate.UserId
	})[0]

	if client.User.Id == rate.UserId {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 11
		client.send <- messageError.encode()
		return
	}

	message.Target = game

	usersQuestions.Rates[client.User.Id].Value = rate.Value
	_, ok := game.Results[rate.UserId]
	if !ok {
		game.Results[rate.UserId].Value = rate.Value
	} else {
		game.Results[rate.UserId].Value += rate.Value
	}

	if len(usersQuestions.Rates) == len(game.Users)-1 {
		game.Round.UsersQuestions = goterators.Filter(game.Round.UsersQuestions, func(item *UserQuestion) bool {
			return item.User != usersQuestions.User
		})
		game.broadcast <- &Message{
			Action: RateEndAction,
			Target: message.Target,
		}
		return
	}

	game.broadcast <- &message
}

func (client *Client) handleUserEndAnswerMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	stage, _, err := goterators.Find(game.Round.UsersQuestions, func(item *UserQuestion) bool {
		return item.User == client.User.Id
	})

	if err != nil {
		return
	}

	for i := range game.Users {
		if game.Users[i].Id != client.User.Id {
			stage.Rates[game.Users[i].Id].Value = 0
		}
	}

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
	if client.User.Id != game.getCreator() {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 8
		client.send <- messageError.encode()
		return
	}

	if len(goterators.Filter(game.Topics, func(item Topic) bool {
		return item.Used == false
	})) == 0 {
		game.endGame()
		game.endGame()
		results := make(map[uuid.UUID]models.Rates)
		for i, v := range game.Results {
			tags := make([]uuid.UUID, len(v.Tags))
			for k := range v.Tags {
				tags[k] = v.Tags[k]
			}
			results[i] = models.Rates{
				Value: v.Value,
				Tags:  tags,
			}
		}
		_ = client.wsServer.service.SaveResults(gameId, results)
		_ = client.wsServer.service.EndGame(game.ID)
		game.broadcast <- &Message{
			Action: EndGameAction,
			Target: game,
		}
		return
	}

	game.Round = &Round{
		Topic:          uuid.Nil,
		UsersQuestions: make([]*UserQuestion, 0),
	}

	var topicId uuid.UUID
	if payload, ok := message.Payload.(uuid.UUID); ok {
		topicId = payload
	} else {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 9
		client.send <- messageError.encode()
		return
	}
	var topic *Topic

	for i := range game.Topics {
		if game.Topics[i].Id == topicId {
			topic = &game.Topics[i]
			break
		}
	}
	if topic == nil {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 9
		client.send <- messageError.encode()
		return
	}
	if len(game.Users) != len(topic.Questions) {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 9
		client.send <- messageError.encode()
		return
	}

	cnt := 1
	for i := range game.Users {
		game.Round.UsersQuestions = append(game.Round.UsersQuestions, &UserQuestion{
			User:     game.Users[i].Id,
			Question: topic.Questions[i].Id,
			Number:   cnt,
			Rates:    make(map[uuid.UUID]*Rates),
		})
		cnt++

	}

	topic.Used = true
	game.Round.Topic = topic.Id

	messageSend = &Message{
		Action:  StartRoundAction,
		Target:  message.Target,
		Payload: game.Round.UsersQuestions,
	}
	game.broadcast <- messageSend
}

func (client *Client) handleStartGameMessage(message Message) {
	var messageSend Message
	gameId := message.Target.ID

	game := client.wsServer.findGame(gameId)
	if game.getCreator() != client.User.Id {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 8
		client.send <- messageError.encode()
		return
	}
	if game.Status == "in_progress" || game.Status == "ended" {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 4
		client.send <- messageError.encode()
		return
	}
	game.Status = "in_progress"

	err := client.wsServer.service.StartGame(gameId)
	if err != nil {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 4
		client.send <- messageError.encode()
		return
	}
	if len(game.Topics) == 0 {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 6
		client.send <- messageError.encode()
		return
	}
	questions := make(map[uuid.UUID][]models.Question)

	for i, _ := range game.Topics {
		questions[game.Topics[i].Id], err = client.wsServer.service.GetRandQuestionsWithLimit(game.Topics[i].Id, len(game.Users))
		if err != nil {
			continue
		}
		game.Topics[i].Questions = make([]Question, len(game.Users))
		for j := 0; j < len(game.Users); j++ {
			if questions[game.Topics[i].Id] != nil {
				tags := make([]Tag, len(questions[game.Topics[i].Id][j].Tags))
				for k := range questions[game.Topics[i].Id][j].Tags {
					tags[k] = Tag{
						Id:   questions[game.Topics[i].Id][j].Tags[k].Id,
						Name: questions[game.Topics[i].Id][j].Tags[k].Name,
					}
				}
				game.Topics[i].Questions[j] = Question{
					Id:      questions[game.Topics[i].Id][j].Id,
					TopicId: questions[game.Topics[i].Id][j].TopicId,
					Content: questions[game.Topics[i].Id][j].Content,
					Tags:    tags,
				}
			}
		}
	}

	messageSend.Action = StartGameAction
	messageSend.Target = game
	game.broadcast <- &messageSend
}

func (client *Client) handleSelectTopicGameMessage(message Message) {
	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)

	if message.Sender.Id != game.getCreator() {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 7
		client.send <- messageError.encode()
		return
	}

	userPlan, err := client.wsServer.service.GetCreatorPlan(game.getCreator())
	if err != nil {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 7
		client.send <- messageError.encode()
		return
	}

	switch userPlan.PlanType {
	case "basic":
		topics, _ := client.wsServer.service.GetRandTopicsWithLimit(3)
		game.setTopics(topics)
		game.notifyClient(client, &message)
		return
	case "advanced", "premium":
		if topics, ok := message.Payload.([]models.Topic); ok {
			for i := range topics {
				topic, _ := client.wsServer.service.GetTopic(topics[i].Id)
				topics[i].Title = topic.Title
			}
			game.setTopics(topics)
			game.notifyClient(client, &message)
			return
		} else {
			var messageError Message
			messageError.Action = Error
			messageError.Target = message.Target
			messageError.Payload = 7
			client.send <- messageError.encode()
			return
		}

	default:
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = 7
		client.send <- messageError.encode()
		return
	}

}

func (client *Client) handleJoinGameMessage(message Message) {

	gameId := message.Target.ID
	game := client.wsServer.findGame(gameId)
	if game == nil {
		return
	}

	game.register <- client
}

func (client *Client) handleLeaveGameMessage(message Message) {
	game := client.wsServer.findGame(message.Target.ID)
	game.unregister <- client
	var messageSend Message
	messageSend.Action = UserLeftAction
	messageSend.Sender = client.User
	messageSend.Target = game
	game.broadcast <- &messageSend
}
