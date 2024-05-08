package requests

import (
	"GameService/connectteam_service/models"
	"github.com/google/uuid"
)

type HTTPService struct {
	Game
	User
	Topic
}

type Game interface {
	GetGame(id uuid.UUID) (models.Game, error)
	SaveResults(id uuid.UUID, results map[uuid.UUID]models.Rates) error
	EndGame(id uuid.UUID) error
	StartGame(id uuid.UUID) error
}
type Topic interface {
	GetTopic(id uuid.UUID) (models.Topic, error)
	GetRandQuestionsWithLimit(topicId uuid.UUID, limit int) ([]models.Question, error)
	GetRandTopicsWithLimit(limit int) (questions []models.Topic, err error)
}

type User interface {
	ParseToken(token string) (id uuid.UUID, access string, err error)
	GetUserById(id uuid.UUID) (user models.User, err error)
	GetCreatorPlan(id uuid.UUID) (plan models.UserPlan, err error)
}

func NewHTTPService(apiKey string) *HTTPService {
	return &HTTPService{
		Game:  NewGameService(apiKey),
		User:  NewUserService(apiKey),
		Topic: NewTopicService(apiKey),
	}
}
