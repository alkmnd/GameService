package models

type Question struct {
	Id      int    `json:"id"`
	TopicId int    `json:"topic_id"`
	Content string `json:"content"`
}
