package models

import "github.com/google/uuid"

type Game struct {
	Id        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	Name      string    `json:"name"`
	CreatorId uuid.UUID `json:"creator_id"`
	MaxSize   int       `json:"max_size"`
}
