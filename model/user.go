package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                   int       `json:"id"`
	RID                  uuid.UUID `json:"rid"`
	Recording1           []byte    `json:"recording_1"`
	Recording2           []byte    `json:"recording_2"`
	Recording3           []byte    `json:"recording_3"`
	Recording1Normalised []byte    `json:"recording_1_normalised"`
	Recording2Normalised []byte    `json:"recording_2_normalised"`
	Recording3Normalised []byte    `json:"recording_3_normalised"`
	// Recording1Mfcc       []float32 `json:"recording_1_mfcc"`
	// Recording2Mfcc       []float32 `json:"recording_2_mfcc"`
	// Recording3Mfcc       []float32 `json:"recording_3_mfcc"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
