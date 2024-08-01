package game

import (
	"GameService/consts/game_status"
	"GameService/consts/plan_types"
	"GameService/repository/models"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ledongthuc/goterators"
	"log"
	"net/http"
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
	ID uuid.UUID
	// The actual websocket connection.
	conn       *websocket.Conn
	wsServer   *WsServer
	send       chan []byte
	Authorized bool
	User       *User
}

// newClient creates a new client.
func newClient(conn *websocket.Conn, wsServer *WsServer, user User, authorized bool) *Client {
	return &Client{
		ID:         uuid.New(),
		User:       &user,
		conn:       conn,
		wsServer:   wsServer,
		send:       make(chan []byte, 256),
		Authorized: authorized,
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

	token, ok := r.URL.Query()["token"]

	var userId uuid.UUID
	var userName string
	var client *Client
	if !ok {
		name, ok := r.URL.Query()["name"]
		if !ok {
			logrus.Println("wrong URL query")
			return
		}

		if len(name[0]) < 0 {
			logrus.Println("wrong param 'name'")
			return
		}

		userName = name[0]
		userId = uuid.New()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Println(fmt.Sprintf("error when upgade: %s", err))
			return
		}

		client = newClient(conn, wsServer, User{
			Id:   userId,
			Name: userName,
		}, false)

	} else if len(token[0]) > 0 {
		id, access, err := wsServer.service.ParseToken(token[0])
		if err != nil {
			logrus.Println(fmt.Sprintf("cannot parse token: %s", err.Error()))
			return
		}

		if access != "user" {
			logrus.Println("cannot connect to the server: permission denied")
			return
		}

		user, err := wsServer.service.GetUserById(id)
		userId = user.Id
		userName = user.FirstName + user.SecondName
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Println(err)
			return
		}

		client = newClient(conn, wsServer, User{
			Id:   userId,
			Name: userName,
		}, true)

	}

	logrus.Println(fmt.Sprintf("user %s successfully connected", userId))

	go client.writePump()
	go client.readPump()

	wsServer.register <- client
}

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Hour

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
		if game := client.wsServer.findGame(message.Target); game != nil {
			game.broadcast <- &message
		}
	case JoinGameAction:
		client.handleJoinGameMessage(message)
	case StartGameAction:
		client.handleStartGameMessage(message)
	case LeaveGameAction:
		client.handleLeaveGameMessage(message) // TODO
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
		client.handleEndGameMessage(message) // TODO
	case DeleteUserAction:
		client.handleDeleteUserAction(message) // TODO
	}

}

func (client *Client) handleDeleteUserAction(message Message) {
	game := client.wsServer.findGame(message.Target)
	if game.getCreator() != client.User.Id {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = ErrorMessage{
			Code:    8,
			Message: "permission denied",
		}
		client.send <- messageError.encode()
		return
	}
	var userId uuid.UUID
	jsonPayload, err := json.Marshal(message.Payload)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonPayload, &userId)
	for i, _ := range game.Clients {
		if i.User.Id == userId {
			game.unregister <- i
			break
		}
	}
	var messageSend Message
	messageSend.Action = UserDeletedAction
	messageSend.Sender = client.User
	messageSend.Target = game.ID
	messageSend.Payload = game.Users
	game.broadcast <- &messageSend
}

type ratePayload struct {
	Value  int         `json:"value"`
	UserId uuid.UUID   `json:"user_id"`
	Tags   []uuid.UUID `json:"tags"`
}

func (client *Client) handleEndGameMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)

	if game == nil {
		return
	}

	if game.getCreator() != client.User.Id {
		var messageError Message
		messageError.Action = Error
		messageError.Target = message.Target
		messageError.Payload = ErrorMessage{
			Code:    8,
			Message: "permission denied",
		}
		client.send <- messageError.encode()
		return
	}

	game.endGame()
	_ = client.wsServer.service.EndGame(game.ID)
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

	game.broadcast <- &Message{
		Action: GameEndedAction,
		Target: game.ID,
		Time:   time.Now(),
	}

	return

}

func (client *Client) handleStartStageMessage(message Message) {

	gameId := message.Target
	game := client.wsServer.findGame(gameId)

	if !game.isCreator(client) {
		return
	}

	game.startStage(client)
}

func (client *Client) handleRateMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)

	if game == nil || game.Round == nil {
		return
	}
	if game.Results == nil {
		game.Results = make(map[uuid.UUID]*Rates)
	}

	var rate ratePayload
	jsonPayload, err := json.Marshal(message.Payload)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonPayload, &rate)
	if err != nil {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    11,
			Message: fmt.Sprintf("cannot ummarshal payload: %s", err.Error())}, game.ID, nil, time.Now()))
		return
	}
	if client.User.Id == rate.UserId {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    11,
			Message: "player cannot rate themselves"}, game.ID, nil, time.Now()))
		return
	}

	game.updateResults(client, rate.UserId, rate.Value, rate.Tags)

	game.broadcast <- &message

	usersQuestions := goterators.Filter(game.Round.UsersQuestions, func(item *UserQuestion) bool {
		return item.User.Id == rate.UserId
	})[0]

	if len(usersQuestions.Rates) == len(game.Users)-1 {
		game.Round.UsersQuestions = goterators.Filter(game.Round.UsersQuestions, func(item *UserQuestion) bool {
			return item.User != usersQuestions.User
		})
		game.broadcast <- NewMessage(
			RateEndAction, nil,
			game.ID,
			nil, time.Now())
		return
	}

}

func (client *Client) handleUserEndAnswerMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)

	game.initRates(client)
	message.Time = time.Now()

	game.broadcast <- &message
}

func (client *Client) handleUserStartAnswerMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)
	message.Time = time.Now()
	game.broadcast <- &message
}

func (client *Client) handleStartRoundMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)
	if !game.isCreator(client) {
		return
	}

	var topicId uuid.UUID
	if payload, ok := message.Payload.(string); ok {
		_uuid, err := uuid.Parse(payload)
		if err != nil {
			client.notifyClient(NewMessage(Error, ErrorMessage{
				Code:    9,
				Message: "incorrect payload",
			}, game.ID, nil, time.Now()))
			return
		}
		topicId = _uuid
	} else {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    9,
			Message: "incorrect payload",
		}, game.ID, nil, time.Now()))
		return
	}

	game.startRound(client, topicId)

}

type startGameMessage struct {
	Game          Game   `json:"game"`
	MeetingNumber string `json:"meeting_number"`
	Passcode      string `json:"passcode"`
	Token         string `json:"token"`
}

func (client *Client) handleStartGameMessage(message Message) {
	gameId := message.Target

	game := client.wsServer.findGame(gameId)
	game.startGame(client)
}

func (client *Client) handleSelectTopicGameMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)

	if !game.isCreator(client) {
		return
	}

	userPlan, err := client.wsServer.service.GetCreatorPlan(game.getCreator())
	if err != nil {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    7,
			Message: fmt.Sprintf("error to get topics from database: %s", err.Error()),
		}, game.ID, nil, time.Now()))
		return
	}

	switch userPlan.PlanType {
	case plan_types.Basic:
		topics, err := client.wsServer.service.GetRandTopicsWithLimit(3)
		if err != nil {
			client.notifyClient(NewMessage(Error, ErrorMessage{
				Code:    7,
				Message: fmt.Sprintf("cannot get topics: %s", err.Error()),
			}, game.ID, nil, time.Now()))
		}
		game.setTopics(topics)
		client.notifyClient(NewMessage(
			message.Action,
			game.Topics,
			game.ID,
			message.Sender,
			time.Now()))
		return
	case plan_types.Advanced, plan_types.Premium:
		var topics []models.Topic
		if payloadData, ok := message.Payload.([]interface{}); ok {
			for i := range payloadData {
				uuidString, ok := payloadData[i].(string)
				if !ok {
					continue
				}
				topicUuid, err := uuid.Parse(uuidString)
				if err != nil {
					continue
				}
				topic, _ := client.wsServer.service.GetTopic(topicUuid)
				topics = append(topics, topic)
			}
			game.setTopics(topics)
			client.notifyClient(NewMessage(
				message.Action,
				game.Topics,
				game.ID,
				message.Sender,
				time.Now()))
			return
		} else {
			client.notifyClient(NewMessage(Error, ErrorMessage{
				Code:    7,
				Message: "incorrect payload",
			}, game.ID, nil, time.Now()))
			return
		}

	default:
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    7,
			Message: "cannot get creator plan",
		}, game.ID, nil, time.Now()))
		return
	}

}

func (client *Client) handleJoinGameMessage(message Message) {
	gameId := message.Target
	game := client.wsServer.findGame(gameId)
	if game == nil {
		message := &Message{
			Action: Error,
			Target: message.Target,
			Payload: ErrorMessage{
				Code:    2,
				Message: fmt.Sprintf("user %s cannot join the game %s", client.User.Id, game.ID),
			},
			Time: time.Now(),
		}
		client.notifyClient(message)
		return
	}
	game.register <- client

}

func (client *Client) notifyClient(message *Message) {
	client.send <- message.encode()
}

func (client *Client) notifyClientJoined(game *Game) {
	message := &Message{
		Action:  JoinGameAction,
		Target:  game.ID,
		Payload: game,
		Sender:  client.User,
	}
	game.broadcastToClientsInGame(message.encode())
}

func (client *Client) handleLeaveGameMessage(message Message) {
	game := client.wsServer.findGame(message.Target)
	if client.User.Id == game.Creator {
		_ = client.wsServer.service.EndGame(game.ID)
		game.Status = game_status.GameEnded
		game.unregister <- client
		game.broadcast <- NewMessage(UserLeftAction, nil, game.ID, nil, time.Now())
		game.broadcast <- NewMessage(GameEndedAction, nil, game.ID, nil, time.Now())
		return

	}

	game.unregister <- client
	game.broadcast <- NewMessage(UserLeftAction, nil, game.ID, nil, time.Now())
	if len(game.Users) < 2 {
		game.broadcast <- NewMessage(GameEndedAction, nil, game.ID, nil, time.Now())
	}
}
