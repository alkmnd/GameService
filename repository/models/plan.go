package models

import (
	"github.com/google/uuid"
)

type UserPlan struct {
	Id       uuid.UUID `json:"id"`
	PlanType string    `json:"plan_type"`
}
