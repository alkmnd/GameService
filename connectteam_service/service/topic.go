package service

import "GameService/connectteam_service/models"

type TopicService struct{}

func NewTopicService() Topic {
	return &TopicService{}
}

func (s *TopicService) GetRandWithLimit(topicId int, limit int) ([]models.Question, error) {
	//TODO implement me
	panic("implement me")
}

func (s *TopicService) GetTopic(id int) (models.Topic, error) {
	//TODO implement me
	panic("implement me")
}
