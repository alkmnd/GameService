package requests

import (
	"GameService/repository/endpoints"
	"GameService/repository/models"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type GameRepo struct {
	apiKey string
}

func NewGameRepo(apiKey string) Game {
	return &GameRepo{apiKey: apiKey}
}

func (s GameRepo) GetResults(gameId uuid.UUID) (results []models.Rates, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("id", gameId.String()).Get(endpoints.GetResultsURL)
	if err != nil {
		return results, err
	}
	err = json.Unmarshal(resp.Body(), &results)
	if err != nil {
		return results, err
	}

	return results, err
}

func (s *GameRepo) SaveResults(id uuid.UUID, results []models.Rates) error {
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

func (s *GameRepo) EndGame(id uuid.UUID) error {
	client := resty.New()
	var _, err = client.R().
		SetHeader("X-API-Key", s.apiKey).
		SetPathParam("id", id.String()).Patch(endpoints.EndGameURL)
	if err != nil {
		return err
	}

	return nil
}

func (s *GameRepo) StartGame(id uuid.UUID) error {
	client := resty.New()
	var _, err = client.R().
		SetHeader("X-API-Key", s.apiKey).
		SetPathParam("id", id.String()).Patch(endpoints.StartGameURL)
	if err != nil {
		return err
	}
	return nil
}

func (s *GameRepo) GetGame(id uuid.UUID) (game models.Game, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("id", id.String()).Get(endpoints.GetGameURL)
	if err != nil {
		return game, err
	}

	println(string(resp.Body()))
	err = json.Unmarshal(resp.Body(), &game)
	if err != nil {
		return game, err
	}

	return game, err
}
