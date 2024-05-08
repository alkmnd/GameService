package requests

import (
	"GameService/connectteam_service/endpoints"
	"GameService/connectteam_service/models"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type GameService struct {
	apiKey string
}

func NewGameService(apiKey string) Game {
	return &GameService{apiKey: apiKey}
}

func (s *GameService) SaveResults(id uuid.UUID, results map[uuid.UUID]models.Rates) error {
	client := resty.New()
	var _, err = client.R().
		SetHeader("X-API-Key", s.apiKey).
		SetBody(map[string]interface{}{"results": results}).
		SetPathParam("id", id.String()).Post(endpoints.SaveResultsURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *GameService) EndGame(id uuid.UUID) error {
	client := resty.New()
	var _, err = client.R().
		SetHeader("X-API-Key", s.apiKey).
		SetPathParam("id", id.String()).Patch(endpoints.EndGameURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *GameService) StartGame(id uuid.UUID) error {
	client := resty.New()
	var _, err = client.R().
		SetHeader("X-API-Key", s.apiKey).
		SetPathParam("id", id.String()).Patch(endpoints.StartGameURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *GameService) GetGame(id uuid.UUID) (game models.Game, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("id", id.String()).Get(endpoints.GetGameURL)
	if err != nil {
		return game, err
	}

	err = json.Unmarshal(resp.Body(), &game)
	if err != nil {
		return game, err
	}

	return game, err
}
