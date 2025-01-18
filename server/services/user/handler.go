package user

import (
	"bytes"
	"database/sql"
	"ht/helper"
	"ht/model"
	"ht/server/database"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const MAX_SIZE_MB = 5

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

func (r *UserService) selectOrInsertUser(userRid uuid.UUID) (*model.User, error) {
	user, err := r.userDb.SelectUser(userRid)
	if err != nil && err == sql.ErrNoRows {
		user, err := r.userDb.InsertUser(&model.User{RID: userRid})
		if err != nil {
			return nil, err
		}
		return user, nil
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserService) CreateReferenceRecording(c echo.Context) (*model.User, error) {
	r.logger.Println("creating recording")

	currentStepString := c.Param("step")
	currentStep, err := strconv.Atoi(currentStepString)
	if err != nil {
		return nil, err
	}

	if err := c.Request().ParseMultipartForm(MAX_SIZE_MB << 20); err != nil {
		return nil, err
	}

	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, MAX_SIZE_MB<<20)
	file, _, err := c.Request().FormFile("recording")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		return nil, err
	}

	userRid := helper.GetCurrentUserRID(c.Request().Context())
	user, err := r.selectOrInsertUser(userRid)
	if err != nil {
		return nil, err
	}

	if currentStep == 1 {
		user.Recording1 = buf.Bytes()
	} else if currentStep == 2 {
		user.Recording2 = buf.Bytes()
	} else if currentStep == 3 {
		user.Recording3 = buf.Bytes()
	} else {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid step value")
	}

	data, err := r.userDb.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return data, nil
}
