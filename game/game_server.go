package game

import (
	service "GameService/connectteam_service/service"
	"log"
)

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	games      map[*Game]bool
	service    *service.HTTPService
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer(service *service.HTTPService) *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		games:      make(map[*Game]bool),
		service:    service,
	}
}

// Run our websocket server, accepting various requests
func (server *WsServer) Run() {
	for {
		select {

		case client := <-server.register:
			log.Println("game_server Run() register")
			server.registerClient(client)

		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			log.Println("game_server Run() broadcast")
			server.broadcastToClients(message)
		}

	}
}

func (server *WsServer) findGame(id int) *Game {
	var foundGame *Game
	for game := range server.games {
		if game.GetId() == id {
			foundGame = game
			return foundGame
		}
	}

	dbGame, err := server.service.GetGame(id)

	if err != nil || dbGame.Status == "ended" {
		return foundGame
	}

	log.Println("game found in data base")

	foundGame = NewGame(dbGame.Name, dbGame.Id, dbGame.CreatorId, "not_started")
	go foundGame.RunGame()
	server.games[foundGame] = true
	return foundGame
}

//func (server *WsServer) createGame(name string, id int) *Game {
//	game := NewGame(name, id)
//	go game.RunGame()
//	server.games[game] = true
//
//	return game
//}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		log.Println("broadcastToClients")
		client.send <- message
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
	}
}

func (server *WsServer) findGameByID(ID int) *Game {
	var foundGame *Game
	for game := range server.games {
		if game.GetId() == ID {
			foundGame = game
			break
		}
	}

	return foundGame
}

func (server *WsServer) findClientByID(ID string) *Client {
	var foundClient *Client
	for client := range server.clients {
		if client.ID.String() == ID {
			foundClient = client
			break
		}
	}

	return foundClient
}
