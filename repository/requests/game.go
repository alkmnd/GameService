package requests

import (
	"GameService/repository/endpoints"
	"GameService/repository/models"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type GameRepo struct {
	apiKey string
}

func NewGameRepo(apiKey string) Game {
	return &GameRepo{apiKey: apiKey}
}

func (s *GameRepo) SaveResults(id uuid.UUID, results map[uuid.UUID]models.Rates) error {
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

	// Вывод тела ответа в виде строки
	fmt.Printf("API Response Body: %s\n", string(resp.Body()))

	// Вывод статуса ответа
	fmt.Printf("API Response Status: %s\n", resp.Status())

	// Вывод заголовков ответа
	fmt.Printf("API Response Headers: %v\n", resp.Header())

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
