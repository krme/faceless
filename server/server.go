package server

import (
	"fmt"
	"ht/helper"
	"ht/server/database"
	"ht/server/services/auth"
	"ht/server/services/user"
	"net/http"

	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
)

// Server represents the server including all services.
type Server struct {
	// session store
	SessionStore *pgstore.PGStore
	sessionDb    *database.DatabaseConfiguration
	// services
	AuthService *auth.AuthService
	UserService *user.UserService
	// jobs
	JobsPort string
}

func NewServer() (*Server, error) {
	sessionDb := &database.DatabaseConfiguration{
		Host:     helper.GetEnvVariable("DB_SESSION_HOST"),
		Port:     helper.GetEnvVariable("DB_SESSION_PORT"),
		Database: helper.GetEnvVariable("DB_SESSION_DATABASE"),
		Username: helper.GetEnvVariable("DB_SESSION_USERNAME"),
		Password: helper.GetEnvVariable("DB_SESSION_PASSWORD"),
		Schema:   helper.GetEnvVariable("DB_SESSION_SCHEMA"),
	}

	sessionStore, err := pgstore.NewPGStore(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", sessionDb.Username, sessionDb.Password, sessionDb.Host, sessionDb.Port, sessionDb.Database, sessionDb.Schema), []byte("secret-key"))
	if err != nil {
		return nil, err
	}
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60,
		SameSite: http.SameSiteLaxMode,
	}

	return &Server{
		SessionStore: sessionStore,
		sessionDb:    sessionDb,
		// services
		AuthService: auth.NewAuthService(sessionStore),
		UserService: user.NewUserService(),
		// jobs
		JobsPort: helper.GetEnvVariableWithoutDelete("JOBS_PORT"),
	}, nil
}
