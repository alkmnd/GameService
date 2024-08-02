package models

import "github.com/google/uuid"

type Rates struct {
	Value  int         `json:"value"`
	Tags   []uuid.UUID `json:"tags"`
	UserId uuid.UUID   `json:"user_id"`
	Name   string      `json:"name"`
}

type GetResultsResponse struct {
	Value  int       `json:"value"`
	Tags   []Tag     `json:"tags"`
	UserId uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}
