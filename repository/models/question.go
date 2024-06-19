package models

import "github.com/google/uuid"

type Question struct {
	Id      uuid.UUID `json:"id"`
	TopicId uuid.UUID `json:"topic_id"`
	Content string    `json:"content"`
	Tags    []Tag     `json:"tags"`
}

type Tag struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
