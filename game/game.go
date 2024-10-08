package game

import (
	"GameService/consts/game_status"
	"GameService/repository/models"
	"fmt"
	"github.com/google/uuid"
	"github.com/ledongthuc/goterators"
	"sync"
	"time"
)

type Game struct {
	Name       string           `json:"name,omitempty"`
	Clients    map[*Client]bool `json:"-"`
	MaxSize    int              `json:"max_size,omitempty"`
	Status     string           `json:"status,omitempty"`
	Creator    uuid.UUID        `json:"creator_id,omitempty"`
	Topics     []Topic          `json:"topics,omitempty"`
	Round      *Round           `json:"round,omitempty"`
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	ID         uuid.UUID            `json:"id"`
	Users      []*User              `json:"users,omitempty"`
	Results    map[uuid.UUID]*Rates `json:"-"`
	mutex      sync.Mutex
}

// UserQuestion Генерируются в начале раунда.
type UserQuestion struct {
	Number   int                  `json:"number"`
	User     User                 `json:"user"`
	Question Question             `json:"question"`
	Rates    map[uuid.UUID]*Rates `json:"-"`
}
type Rates struct {
	Value int                `json:"value"`
	Tags  map[uuid.UUID]bool `json:"tags"`
}

type Round struct {
	Topic          uuid.UUID       `json:"topic"`
	UsersQuestions []*UserQuestion `json:"users-questions"`
}

type Topic struct {
	Id        uuid.UUID  `json:"id"`
	Used      bool       `json:"used"`
	Title     string     `json:"title,omitempty"`
	Questions []Question `json:"questions,omitempty"`
}

type Question struct {
	Id      uuid.UUID `json:"id"`
	TopicId uuid.UUID `json:"topic_id"`
	Content string    `json:"content"`
	Tags    []Tag     `json:"tags"`
}

type Tag struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// NewGame creates a new game.
func NewGame(name string, id uuid.UUID, creator uuid.UUID, status string, maxSize int) *Game {
	return &Game{
		ID:         id,
		Name:       name,
		Topics:     make([]Topic, 0),
		Creator:    creator,
		Status:     status,
		MaxSize:    maxSize,
		Users:      make([]*User, 0),
		Clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

func (game *Game) GetId() uuid.UUID {
	return game.ID
}

func (game *Game) GetName() string {
	return game.Name
}

// RunGame run game, accepting various requests.
func (game *Game) RunGame() {
	for {
		select {
		case client := <-game.register:
			game.registerClientInGame(client)

		case client := <-game.unregister:
			game.unregisterClientInGame(client)

		case message := <-game.broadcast:
			game.broadcastToClientsInGame(message.encode())
		}
	}
}

func (game *Game) registerClientInGame(client *Client) {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	for i := range game.Users {
		if game.Users[i].Id == client.User.Id {
			message := NewMessage(UserJoinedAction, game, game.ID, client.User, time.Now())
			client.notifyClient(message)
			game.Clients[client] = true
			return
		}
	}

	if len(game.Users) == game.MaxSize {
		message := NewMessage(Error, ErrorMessage{
			Code:    1,
			Message: "max number of participants",
		}, game.ID, nil, time.Now())

		client.notifyClient(message)
		return
	}

	if game.Status == game_status.GameInProgress || game.Status == game_status.GameEnded {
		message := NewMessage(Error, ErrorMessage{
			Code:    2,
			Message: "game in progress",
		}, game.ID, client.User, time.Now())
		client.notifyClient(message)
		return
	}

	message := NewMessage(UserJoinedAction, game, game.ID, client.User, time.Now())

	game.Users = append(game.Users, client.User)
	client.notifyClientJoined(game)
	game.Clients[client] = true
	client.notifyClient(message)
	return
}

func (game *Game) endGame() {
	game.Status = game_status.GameEnded
}

func (game *Game) unregisterClientInGame(client *Client) {
	if _, ok := game.Clients[client]; ok {
		delete(game.Clients, client)
	}

	for i := range game.Users {
		if game.Users[i].Id == client.User.Id {
			game.Users = append(game.Users[:i], game.Users[i+1:]...)
			break
		}
	}

	if game.Round != nil && game.Round.UsersQuestions != nil {
		for i := range game.Round.UsersQuestions {
			if game.Round.UsersQuestions[i].User.Id == client.User.Id {
				delete(game.Round.UsersQuestions[i].Rates, game.Round.UsersQuestions[i].User.Id)
			}
		}
	}

}

func (game *Game) broadcastToClientsInGame(message []byte) {
	for client := range game.Clients {
		client.send <- message
	}
}

func (game *Game) getCreator() uuid.UUID {
	return game.Creator
}

func (game *Game) startRound(client *Client, topicId uuid.UUID) {
	if game.Status == game_status.GameEnded {
		return
	}
	if len(goterators.Filter(game.Topics, func(item Topic) bool {
		return item.Used == false
	})) == 0 {
		game.endGame()
		results := make([]models.Rates, 0)
		for i, _ := range game.Users {
			userTempId := game.Users[i].Id
			userId := uuid.Nil
			if game.Users[i].Authorized {
				userId = game.Users[i].Id
			}
			var tags []uuid.UUID
			for j := range game.Results[game.Users[i].Id].Tags {
				tags = append(tags, j)
			}
			results = append(results, models.Rates{
				Value:           game.Results[game.Users[i].Id].Value,
				Tags:            tags,
				UserId:          userId,
				UserTemporaryId: userTempId,
				Name:            game.Users[i].Name,
			})
		}
		_ = client.wsServer.service.SaveResults(game.ID, results)
		_ = client.wsServer.service.EndGame(game.ID)
		payload, _ := client.wsServer.service.GetResults(game.ID)
		game.broadcast <- &Message{
			Action:  GameEndedAction,
			Payload: payload,
			Target:  game.ID,
		}
		return
	}

	game.Round = &Round{
		Topic:          uuid.Nil,
		UsersQuestions: make([]*UserQuestion, 0),
	}

	var topic *Topic

	for i := range game.Topics {
		if game.Topics[i].Id == topicId {
			topic = &game.Topics[i]
			break
		}
	}
	if topic == nil || topic.Used == true {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    9,
			Message: "topic is already used",
		}, game.ID, nil, time.Now()))
		return
	}
	if len(game.Users) != len(topic.Questions) {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    8,
			Message: "number of users is not equal to number of questions",
		}, game.ID, nil, time.Now()))
		return
	}

	cnt := 1
	for i := range game.Users {
		game.Round.UsersQuestions = append(game.Round.UsersQuestions, &UserQuestion{
			User:     *game.Users[i],
			Question: topic.Questions[i],
			Number:   cnt,
			Rates:    make(map[uuid.UUID]*Rates),
		})
		cnt++
	}

	topic.Used = true
	game.Round.Topic = topic.Id
	game.broadcast <- &Message{
		Action:  StartRoundAction,
		Target:  game.ID,
		Payload: game.Round.UsersQuestions,
		Time:    time.Now(),
	}
}

func (game *Game) setTopics(topics []models.Topic) {
	game.Topics = make([]Topic, 0)
	for i := range topics {

		game.Topics = append(game.Topics, Topic{
			Id:        topics[i].Id,
			Used:      false,
			Title:     topics[i].Title,
			Questions: nil,
		})
	}
}

func (game *Game) isCreator(client *Client) bool {
	if game.getCreator() != client.User.Id {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    8,
			Message: "permission denied",
		}, game.ID, nil, time.Now()))
		return false
	}
	return true
}

func (game *Game) startGame(client *Client) {
	if !game.isCreator(client) {
		return
	}
	if len(game.Topics) == 0 {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    6,
			Message: "number of topics is 0",
		},
			game.ID, nil, time.Now()))
		return
	}
	if game.Status == game_status.GameInProgress || game.Status == game_status.GameEnded {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    5,
			Message: "game is in progress or ended",
		}, game.ID, nil, time.Now()))
		return
	}

	meetingNumber, passcode, err := client.wsServer.service.Meeting.CreateMeeting()
	if err != nil {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    4,
			Message: fmt.Sprintf("error to start game: %s", err.Error()),
		}, game.ID, nil, time.Now()))
		return
	}
	meetingJWT, _ := client.wsServer.generator.GenerateJWTForMeeting(meetingNumber, 0)

	questions := make(map[uuid.UUID][]models.Question)

	for i, _ := range game.Topics {
		var err error
		questions[game.Topics[i].Id], err = client.wsServer.service.GetRandQuestionsWithLimit(game.Topics[i].Id, len(game.Users))
		if err != nil {
			continue
		}
		if len(questions[game.Topics[i].Id]) != len(game.Users) {
			client.notifyClient(NewMessage(Error, ErrorMessage{
				Code:    4,
				Message: fmt.Sprintf("not enough question to start game"),
			}, game.ID, nil, time.Now()))
			return
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

	err = client.wsServer.service.StartGame(game.ID)
	if err != nil {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    4,
			Message: fmt.Sprintf("error to start game: %s", err.Error()),
		}, game.ID, nil, time.Now()))
		return
	}

	var payload = &startGameMessage{
		Game:          *game,
		MeetingNumber: meetingNumber,
		Passcode:      passcode,
		Token:         meetingJWT,
	}

	hostMeetingJWT, _ := client.wsServer.generator.GenerateJWTForMeeting(meetingNumber, 1)

	game.Status = "in_progress"
	for client := range game.Clients {
		if client.User.Id == game.Creator {
			payload.Token = hostMeetingJWT
		}
		client.notifyClient(NewMessage(StartGameAction, &payload, game.ID, client.User, time.Now()))
	}
}

func (game *Game) startStage(client *Client) {
	if game.Status == game_status.GameEnded {
		return
	}
	if len(goterators.Filter(game.Topics, func(item Topic) bool {
		return item.Used == false
	})) == 0 && len(game.Round.UsersQuestions) == 0 {
		game.endGame()
		results := make([]models.Rates, 0)
		for i, _ := range game.Users {
			userTempId := game.Users[i].Id
			userId := uuid.Nil
			if game.Users[i].Authorized {
				userId = game.Users[i].Id
			}
			var tags []uuid.UUID
			for j := range game.Results[game.Users[i].Id].Tags {
				tags = append(tags, j)
			}
			results = append(results, models.Rates{
				Value:           game.Results[game.Users[i].Id].Value,
				Tags:            tags,
				UserId:          userId,
				UserTemporaryId: userTempId,
				Name:            game.Users[i].Name,
			})
		}
		_ = client.wsServer.service.SaveResults(game.ID, results)
		_ = client.wsServer.service.EndGame(game.ID)
		payload, _ := client.wsServer.service.GetResults(game.ID)
		game.broadcast <- &Message{
			Action:  GameEndedAction,
			Payload: payload,
			Target:  game.ID,
		}

		return
	}
	if len(game.Round.UsersQuestions) == 0 {
		game.broadcast <- &Message{
			Action:  RoundEndAction,
			Target:  game.ID,
			Payload: game.Topics,
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
		Target:  game.ID,
		Payload: payload,
		Sender:  client.User,
		Time:    time.Now(),
	}
}

func (game *Game) initRates(client *Client) {
	if game.Round == nil || game.Round.UsersQuestions == nil {
		client.notifyClient(NewMessage(Error, 12, game.ID, nil, time.Now()))
		return
	}

	stage, _, err := goterators.Find(game.Round.UsersQuestions, func(item *UserQuestion) bool {
		return item.User.Id == client.User.Id
	})

	if err != nil {
		client.notifyClient(NewMessage(Error, ErrorMessage{
			Code:    12,
			Message: fmt.Sprintf("error to find stage: %s", err.Error()),
		}, game.ID, nil, time.Now()))
		return
	}

	for i := range game.Users {
		if game.Users[i].Id != client.User.Id {
			stage.Rates[game.Users[i].Id] = &Rates{
				Value: 0,
				Tags:  make(map[uuid.UUID]bool),
			}
		}
	}
}

func (game *Game) updateResults(client *Client, user uuid.UUID, value int, tags []uuid.UUID) {
	usersQuestions := goterators.Filter(game.Round.UsersQuestions, func(item *UserQuestion) bool {
		return item.User.Id == user
	})[0]
	_, rateOk := usersQuestions.Rates[client.User.Id]
	if !rateOk {
		usersQuestions.Rates[client.User.Id] = &Rates{
			Value: 0,
			Tags:  make(map[uuid.UUID]bool),
		}
		usersQuestions.Rates[client.User.Id].Value = value
		for i := range tags {
			usersQuestions.Rates[client.User.Id].Tags[tags[i]] = true
		}
		_, ok := game.Results[user]
		if !ok {
			game.Results[user] = &Rates{
				Value: 0,
				Tags:  make(map[uuid.UUID]bool),
			}
			game.Results[user].Value = value
		} else {
			game.Results[user].Value += value
		}
		if tags != nil {
			for i := range tags {
				game.Results[user].Tags[tags[i]] = true
			}
		}
	}
}
