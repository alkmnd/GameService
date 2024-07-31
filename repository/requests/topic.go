package requests

import (
	"GameService/repository/endpoints"
	"GameService/repository/models"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"strconv"
)

type TopicRepo struct {
	apiKey string
}

func NewTopicRepo(apiKey string) Topic {
	return &TopicRepo{apiKey: apiKey}
}

func (s *TopicRepo) GetRandQuestionsWithLimit(topicId uuid.UUID, limit int) (questions []models.Question, _ error) {
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

func (s *TopicRepo) GetRandTopicsWithLimit(limit int) (topics []models.Topic, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("limit", strconv.Itoa(limit)).Get(endpoints.GetRandTopicsURL)
	if err != nil {
		return topics, err
	}

	err = json.Unmarshal(resp.Body(), &topics)
	if err != nil {
		return topics, err
	}

	return topics, err
}
func (s *TopicRepo) GetTopic(id uuid.UUID) (topic models.Topic, err error) {
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
