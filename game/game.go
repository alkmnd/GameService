package game

import (
	"encoding/json"
	"log"
)

type Game struct {
	Name       string           `json:"name,omitempty"`
	Clients    map[*Client]bool `json:"-"`
	MaxSize    int              `json:"max_size,omitempty"`
	Status     string           `json:"status,omitempty"`
	Creator    int              `json:"creator_id,omitempty"`
	Topics     []Topic          `json:"topics,omitempty"`
	RoundsLeft []*Round         `json:"-"`
	Round      *Round           `json:"round,omitempty"`
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	ID         int         `json:"id"`
	Users      []*User     `json:"users,omitempty"`
	Results    map[int]int `json:"-"`
}

// UsersQuestions Генерируются в начале раунда.
type UsersQuestions struct {
	Number   int         `json:"number"`
	User     *User       `json:"user"`
	Question string      `json:"question"`
	Rates    map[int]int `json:"-"`
}

type Round struct {
	Topic              *Topic            `json:"topic"`
	UsersQuestions     []*UsersQuestions `json:"users-questions"`
	UsersQuestionsLeft []*UsersQuestions `json:"users-questions-left"`
}

type Topic struct {
	Id        int      `json:"id"`
	Used      bool     `json:"used"`
	Title     string   `json:"title,omitempty"`
	Questions []string `json:"questions,omitempty"`
}

func NewGame(name string, id int, creator int, status string) *Game {
	return &Game{
		ID:      id,
		Name:    name,
		Topics:  make([]Topic, 0),
		Creator: creator,
		Status:  status,
		MaxSize: 3,
		Users:   make([]*User, 0),
		// state
		Clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

func (game *Game) GetId() int {
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

const welcomeMessage = "%s joined the room"

type UserList struct {
	Users []User `json:"users"`
}

func (game *Game) listUsersInGame(client *Client) {
	for existingClient := range game.Clients {
		message := &Message{
			Action: UserJoinedAction,
			Target: game,
			Sender: existingClient.User,
		}
		game.broadcastToClientsInGame(message.encode())
	}
}

func (game *Game) notifyClientJoined(client *Client) {

	message := &Message{
		Action:  JoinGameAction,
		Target:  game,
		Payload: []byte{},
		Sender:  client.User,
	}
	log.Println("notifyClientJoined")

	game.broadcastToClientsInGame(message.encode())
}

func (game *Game) notifyClient(client *Client, message *Message) {
	client.send <- message.encode()
}

func (game *Game) registerClientInGame(client *Client) {
	log.Println("client try to join")

	for i := range game.Users {
		if game.Users[i].Id == client.User.Id {
			message := &Message{
				Action:  UserJoinedAction,
				Target:  game,
				Payload: []byte{},
				Sender:  client.User,
			}

			game.notifyClient(client, message)
			game.Clients[client] = true

			log.Println("client already register")
			return
		}
	}

	if game.Status == "in_progress" {
		payload, _ := json.Marshal("game in progress")
		message := &Message{
			Action: Error,
			Target: &Game{
				ID: game.ID,
			},
			Payload: payload,
			Sender:  client.User,
		}
		game.notifyClient(client, message)
		return
	}

	message := &Message{
		Action:  UserJoinedAction,
		Target:  game,
		Payload: []byte{},
		Sender:  client.User,
	}

	game.Users = append(game.Users, client.User)
	log.Println("client joined")
	game.notifyClientJoined(client)
	game.Clients[client] = true
	game.notifyClient(client, message)
	//game.listUsersInGame(client)
	//}

	//message := &Message{
	//	Action:  Error,
	//	Target:  game,
	//	Payload: []byte{},
	//	Sender:  &client.User,
	//}

	return
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
			if game.Round.UsersQuestions[i].User.Id == client.User.Id {
				delete(game.Round.UsersQuestions[i].Rates, game.Round.UsersQuestions[i].User.Id)
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
