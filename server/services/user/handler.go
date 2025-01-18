package user

import (
	"ht/helper"
	"ht/server/database"
	"log"
	"os"

	"github.com/antonlindstrom/pgstore"
	"github.com/google/uuid"
)

type UserService struct {
	logger *log.Logger
	userDb UserDBHandlerFunctions
}

func NewUserService(sessionStore *pgstore.PGStore) *UserService {
	logger := log.New(os.Stdout, "user: ", log.LstdFlags)
	dbConnection := database.NewDatabase(
		"user",
		&database.DatabaseConfiguration{
			Host:     helper.GetEnvVariable("DB_USER_HOST"),
			Port:     helper.GetEnvVariable("DB_USER_PORT"),
			Database: helper.GetEnvVariable("DB_USER_DATABASE"),
			Username: helper.GetEnvVariable("DB_USER_USERNAME"),
			Password: helper.GetEnvVariable("DB_USER_PASSWORD"),
			Schema:   helper.GetEnvVariable("DB_USER_SCHEMA"),
		},
	)
	var userDb UserDBHandlerFunctions = newUserDBHandler(dbConnection)

	// creates main user table
	err := userDb.CreateTable(uuid.UUID{})
	if err != nil {
		log.Fatal(err.Error())
	}

	newUserService := &UserService{
		logger: logger,
		userDb: userDb,
	}

	return newUserService
}
