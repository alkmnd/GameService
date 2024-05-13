package game

import (
	"GameService/connectteam_service/models"
	"github.com/google/uuid"
	"log"
	"time"
)

type Game struct {
	Name         string           `json:"name,omitempty"`
	Clients      map[*Client]bool `json:"-"`
	MaxSize      int              `json:"max_size,omitempty"`
	Status       string           `json:"status,omitempty"`
	Creator      uuid.UUID        `json:"creator_id,omitempty"`
	Topics       []Topic          `json:"topics,omitempty"`
	Round        *Round           `json:"round,omitempty"`
	register     chan *Client
	CurrentState interface{} `json:"current_state,omitempty"`
	unregister   chan *Client
	broadcast    chan *Message
	ID           uuid.UUID            `json:"id"`
	Users        []*User              `json:"users,omitempty"`
	Results      map[uuid.UUID]*Rates `json:"-"`
}

// UserQuestion Генерируются в начале раунда.
type UserQuestion struct {
	Number   int                  `json:"number"`
	User     uuid.UUID            `json:"user"`
	Question Question             `json:"question"`
	Rates    map[uuid.UUID]*Rates `json:"-"`
}
type Rates struct {
	Value int         `json:"value"`
	Tags  []uuid.UUID `json:"tags"`
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

func NewGame(name string, id uuid.UUID, creator uuid.UUID, status string, maxSize int) *Game {
	return &Game{
		ID:      id,
		Name:    name,
		Topics:  make([]Topic, 0),
		Creator: creator,
		Status:  status,
		MaxSize: maxSize,
		Users:   make([]*User, 0),
		// state
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

func (game *Game) RunGame() {
	log.Println("RunGame()")
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

func (game *Game) notifyClientJoined(client *Client) {

	message := &Message{
		Action:  JoinGameAction,
		Target:  game.ID,
		Payload: game,
		Sender:  client.User,
	}

	game.broadcastToClientsInGame(message.encode())
}

func (game *Game) notifyClient(client *Client, message *Message) {
	client.send <- message.encode()
}

func (game *Game) registerClientInGame(client *Client) {
	if len(game.Users) == game.MaxSize {
		message := Message{
			Action: Error,
			Payload: ErrorMessage{
				Code:    1,
				Message: "max number of participants",
			},
			Target: game.ID,
			Time:   time.Now(),
		}
		client.send <- message.encode()
		return
	}

	for i := range game.Users {
		if game.Users[i].Id == client.User.Id {
			message := &Message{
				Action:  UserJoinedAction,
				Target:  game.ID,
				Payload: game,
				Sender:  client.User,
			}

			game.notifyClient(client, message)
			game.Clients[client] = true
			return
		}
	}

	if game.Status == "in_progress" {
		message := &Message{
			Action: Error,
			Target: game.ID,
			Payload: ErrorMessage{
				Code:    2,
				Message: "game in progress",
			},
			Sender: client.User,
		}
		game.notifyClient(client, message)
		return
	}

	message := &Message{
		Action:  UserJoinedAction,
		Target:  game.ID,
		Payload: game,
		Sender:  client.User,
		Time:    time.Now(),
	}

	game.Users = append(game.Users, client.User)
	game.notifyClientJoined(client)
	game.Clients[client] = true
	game.notifyClient(client, message)
	return
}

func (game *Game) endGame() {
	game.Status = "ended"
}

func (game *Game) unregisterClientInGame(client *Client) {
	if _, ok := game.Clients[client]; ok {
		delete(game.Clients, client)
	}

	for i := range game.Users {
		if game.Users[i].Id == client.User.Id {
			game.Users = append(game.Users[:i], game.Users[i+1:]...)
		}
	}

	if game.Round != nil && game.Round.UsersQuestions != nil {
		for i := range game.Round.UsersQuestions {
			if game.Round.UsersQuestions[i].User == client.User.Id {
				delete(game.Round.UsersQuestions[i].Rates, game.Round.UsersQuestions[i].User)
			}
		}
	}

}

func (game *Game) broadcastToClientsInGame(message []byte) {
	for client := range game.Clients {
		log.Printf("broadcast message to client %s", client.GetName())
		client.send <- message
	}
}

func (game *Game) getCreator() uuid.UUID {
	return game.Creator
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
