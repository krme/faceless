package user

import (
	"ht/helper"
	"ht/model"
	"ht/server/database"
	"log"
	"os"

	"github.com/labstack/echo/v4"
)

type UserService struct {
	logger *log.Logger
	userDb UserDBHandlerFunctions
}

func NewUserService() *UserService {
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
	err := userDb.CreateTable()
	if err != nil {
		log.Fatal(err.Error())
	}

	newUserService := &UserService{
		logger: logger,
		userDb: userDb,
	}

	return newUserService
}

func (r *UserService) CreateReferenceRecording(c echo.Context) (*model.User, error) {
	r.logger.Println("creating recording")

	userRid := helper.GetCurrentUserRID(c.Request().Context())

	user, err := r.userDb.SelectUser(userRid)
	if err != nil {
		return nil, err
	}

	// TODO get recording from request body
	// if step == "1" {
	// 	user.Recording1 = recording
	//}

	data, err := r.userDb.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return data, nil
}
