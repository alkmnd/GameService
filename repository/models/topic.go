package models

import "github.com/google/uuid"

type Topic struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}
