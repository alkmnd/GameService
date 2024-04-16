package service

import "GameService/connectteam_service/models"

type GameService struct{}

func NewGameService() Game {
	return &GameService{}
}

func (s *GameService) SaveResults(id int, results map[int]int) error {
	//TODO implement me
	panic("implement me")
}

func (s *GameService) EndGame(id int) error {
	//TODO implement me
	panic("implement me")
}

func (s *GameService) StartGame(id int) error {
	//TODO implement me
	panic("implement me")
}

func (s *GameService) GetGame(id int) (models.Game, error) {
	//TODO implement me
	panic("implement me")
}
