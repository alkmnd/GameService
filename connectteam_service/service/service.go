package service

import "GameService/connectteam_service/models"

type HTTPService struct {
	Game
	User
	Topic
}

type Game interface {
	GetGame(id int) (models.Game, error)
	SaveResults(id int, results map[int]int) error
	EndGame(id int) error
	StartGame(id int) error
}
type Topic interface {
	GetTopic(id int) (models.Topic, error)
	GetRandWithLimit(topicId int, limit int) ([]models.Question, error)
}

type User interface {
	ParseToken(token string) (id int, access string, err error)
	GetUserById(id int) (user models.User, err error)
}

func NewService() *HTTPService {
	return &HTTPService{
		Game:  NewGameService(),
		User:  NewUserService(),
		Topic: NewTopicService(),
	}
}
