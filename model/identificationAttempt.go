package model

import (
	"time"

	"github.com/google/uuid"
)

type IdentificationAttempt struct {
	ID         int       `json:"id"`
	RID        uuid.UUID `json:"rid"`
	UserRID    uuid.UUID `json:"user_rid"`
	Recording  []byte    `json:"recording"`
	Identified bool      `json:"identified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
