package game

import (
	service "GameService/repository/requests"
	"github.com/google/uuid"
)

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	games      map[*Game]bool
	service    *service.Repository
	generator  *JWTGenerator
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer(service *service.Repository, generator *JWTGenerator) *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		games:      make(map[*Game]bool),
		service:    service,
		generator:  generator,
	}
}

// Run our websocket server, accepting various requests
func (server *WsServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		}

	}
}

func (server *WsServer) findGame(id uuid.UUID) *Game {
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
	var maxSize int
	creatorPlan, _ := server.service.GetCreatorPlan(dbGame.CreatorId)
	switch creatorPlan.PlanType {
	case "basic":
		maxSize = 3
	case "advanced":
		maxSize = 5
	case "premium":
		maxSize = 10
	}
	foundGame = NewGame(dbGame.Name, dbGame.Id, dbGame.CreatorId, dbGame.Status, maxSize)
	go foundGame.RunGame()
	server.games[foundGame] = true
	return foundGame
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
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

func (server *WsServer) findGameByID(ID uuid.UUID) *Game {
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
