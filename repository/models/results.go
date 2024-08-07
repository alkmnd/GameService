package models

import "github.com/google/uuid"

type Rates struct {
	Value           int         `json:"value"`
	Tags            []uuid.UUID `json:"tags"`
	UserId          uuid.UUID   `json:"user_id"`
	UserTemporaryId uuid.UUID   `json:"user_temp_id"`
	Name            string      `json:"name"`
}

type Results struct {
	Value           int       `json:"value"`
	Tags            []Tag     `json:"tags"`
	UserId          uuid.UUID `json:"user_id"`
	UserTemporaryId uuid.UUID `json:"user_temp_id"`
	Name            string    `json:"name"`
}

type GetResultsResponse struct {
	Results []Results `json:"results"`
}
