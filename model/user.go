package model

import "time"

type User struct {
	ID         int       `json:"id"`
	RID        string    `json:"rid"`
	Recording1 []byte    `json:"recording_1"`
	Recording2 []byte    `json:"recording_2"`
	Recording3 []byte    `json:"recording_3"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
