package requests

import (
	"GameService/repository/models"
	"github.com/google/uuid"
)

type Repository struct {
	Game
	User
	Topic
	Meeting
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

type Meeting interface {
	CreateMeeting() (meetingNumber string, passcode string, err error)
}

func NewHTTPService(apiKey string,
	accessToken string,
	refreshToken string,
	clientId string,
	clientSecret string) *Repository {
	return &Repository{
		Game:    NewGameRepo(apiKey),
		User:    NewUserService(apiKey),
		Topic:   NewTopicRepo(apiKey),
		Meeting: NewMeetingRepo(accessToken, refreshToken, clientId, clientSecret),
	}
}
