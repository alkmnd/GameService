package requests

import (
	"GameService/connectteam_service/endpoints"
	"GameService/connectteam_service/models"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"strconv"
)

type TopicService struct {
	apiKey string
}

func NewTopicService(apiKey string) Topic {
	return &TopicService{apiKey: apiKey}
}

func (s *TopicService) GetRandQuestionsWithLimit(topicId uuid.UUID, limit int) (questions []models.Question, _ error) {
	client := resty.New()
	println(topicId.String())
	var resp, err = client.R().
		SetHeader("X-API-Key", s.apiKey).
		SetQueryParams(map[string]string{
			"topic_id": topicId.String(),
			"limit":    strconv.Itoa(limit),
		}).
		Get(endpoints.GetRandQuestionsURL)
	if err != nil {
		return questions, err
	}

	err = json.Unmarshal(resp.Body(), &questions)
	println(string(resp.Body()))
	if err != nil {
		return questions, err
	}

	return questions, err
}

func (s *TopicService) GetRandTopicsWithLimit(limit int) (questions []models.Topic, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("limit", strconv.Itoa(limit)).Get(endpoints.GetRandTopicsURL)
	if err != nil {
		return questions, err
	}

	err = json.Unmarshal(resp.Body(), &questions)
	if err != nil {
		return questions, err
	}

	return questions, err
}
func (s *TopicService) GetTopic(id uuid.UUID) (topic models.Topic, err error) {
	topic.Id = id
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("id", id.String()).Get(endpoints.GetTopicWithIdURL)
	if err != nil {
		return topic, err
	}

	err = json.Unmarshal(resp.Body(), &topic)
	if err != nil {
		return topic, err
	}

	return topic, err
}
