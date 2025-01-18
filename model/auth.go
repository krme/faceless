package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Authenticated bool
	UserID        uuid.UUID
	EmailVerified bool
	CreatedAt     time.Time
}

type Auth struct {
	ID                           int       `json:"id"`
	RID                          uuid.UUID `json:"rid"`
	Email                        string    `json:"email"`
	EmailVerified                bool      `json:"email_verified"`
	PasswordTemp                 string    `json:"password_temp"`
	PasswordTempRequestDate      time.Time `json:"-"`
	PasswordTempValid            bool      `json:"password_temp_valid"`
	PasswordSet                  bool      `json:"password_set"`
	PasswordHash                 string    `json:"-"`
	PasswordResetCodeHash        string    `json:"-"`
	PasswordResetRequestDate     time.Time `json:"-"`
	EmailVerificationCodeHash    string    `json:"-"`
	EmailVerificationRequestDate time.Time `json:"-"`
	EmailToChangeTo              string    `json:"-"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
	// input fields not saved
	Password string `json:"password" vld:"min8 max30 rex'^(.*[A-Z])+(.*)$' rex'^(.*[a-z])+(.*)$' rex'^(.*\\d)+(.*)$' rex'^(.*[\x60!@#$%^&*()_+={};/':\"|\\,.<>/?~-])+(.*)$'"`
}

func (r *Auth) IsEmpty() bool {
	return r.ID == 0
}

func (r *Auth) ToKey() string {
	return "not supported"
}

func (r *Auth) ToName() string {
	return "not supported"
}

func (r *Auth) GetRidString() string {
	return r.RID.String()
}
