package models

import "github.com/google/uuid"

type Rates struct {
	Value int         `json:"value"`
	Tags  []uuid.UUID `json:"tags"`
}
