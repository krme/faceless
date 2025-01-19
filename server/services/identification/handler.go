package identification

import (
	"bytes"
	"fmt"
	"ht/helper"
	"ht/model"
	"ht/server/database"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

const MAX_SIZE_MB = 5

type IdentificationAttemptService struct {
	logger                  *log.Logger
	identificationAttemptDb IdentificationAttemptDBHandlerFunctions
	jobsPort                string
}

func NewIdentificationAttemptService() *IdentificationAttemptService {
	logger := log.New(os.Stdout, "identificationAttempt: ", log.LstdFlags)
	dbConnection := database.NewDatabase(
		"identificationAttempt",
		&database.DatabaseConfiguration{
			Host:     helper.GetEnvVariable("DB_IDENTIFICATION_HOST"),
			Port:     helper.GetEnvVariable("DB_IDENTIFICATION_PORT"),
			Database: helper.GetEnvVariable("DB_IDENTIFICATION_DATABASE"),
			Username: helper.GetEnvVariable("DB_IDENTIFICATION_USERNAME"),
			Password: helper.GetEnvVariable("DB_IDENTIFICATION_PASSWORD"),
			Schema:   helper.GetEnvVariable("DB_IDENTIFICATION_SCHEMA"),
		},
	)
	var identificationAttemptDb IdentificationAttemptDBHandlerFunctions = newIdentificationAttemptDBHandler(dbConnection)

	// creates main identificationAttempt table
	err := identificationAttemptDb.CreateTable()
	if err != nil {
		log.Fatal(err.Error())
	}

	newIdentificationAttemptService := &IdentificationAttemptService{
		logger:                  logger,
		identificationAttemptDb: identificationAttemptDb,
		jobsPort:                helper.GetEnvVariableWithoutDelete("JOBS_PORT"),
	}

	return newIdentificationAttemptService
}

func (r *IdentificationAttemptService) CreateIdentificationAttempt(c echo.Context) (*model.IdentificationAttempt, error) {
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

	identificationAttempt := &model.IdentificationAttempt{
		UserRID:   helper.GetCurrentUserRID(c.Request().Context()),
		Recording: buf.Bytes(),
	}

	data, err := r.identificationAttemptDb.InsertIdentificationAttempt(identificationAttempt)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *IdentificationAttemptService) GetLatestIdentificationAttempt(c echo.Context) (*model.IdentificationAttempt, error) {
	userId := helper.GetCurrentUserRID(c.Request().Context())
	identificationAttempt, err := r.identificationAttemptDb.SelectLatestIdentificationAttemptByUserRID(userId)
	if err != nil {
		return nil, err
	}

	return identificationAttempt, nil
}

// jobs
func (r *IdentificationAttemptService) ProcessIdentificationAttempt(c echo.Context) (bool, error) {
	userId := helper.GetCurrentUserRID(c.Request().Context())
	body, err := helper.StartJob(fmt.Sprintf("http://localhost:%v/jobs/identify", r.jobsPort), map[string]string{"user_rid": userId.String()})
	if err != nil {
		return false, err
	}

	if string(body) == "true" {
		return true, nil
	}
	return false, nil
}
