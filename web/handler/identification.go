package handler

import (
	"encoding/json"
	"fmt"
	"ht/helper"
	"ht/model"
	"ht/server"
	"ht/web/view/screens"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type IdentificationView struct {
	server *server.Server
}

func NewIdentificationView(server *server.Server) *IdentificationView {
	newIdentificationView := &IdentificationView{
		server: server,
	}
	return newIdentificationView
}

func (r *IdentificationView) HandleIdentification(c echo.Context) error {
	bytes, err := helper.StartJob(fmt.Sprintf("http://localhost:%v/jobs/createSentence", r.server.JobsPort), map[string]string{})
	if err != nil {
		return err
	}

	sentence := ""
	err = json.Unmarshal(bytes, &sentence)
	if err != nil {
		return err
	}

	return render(c, screens.Identification(sentence))
}

func (r *IdentificationView) HandleAuthenticationWaiting(c echo.Context) error {
	userId := helper.GetCurrentUserRID(c.Request().Context())
	log.Printf("starting identification for userId: %v", userId)
	identificationAttempt, err := helper.StartJob(fmt.Sprintf("http://localhost:%v/jobs/identify", r.server.JobsPort), map[string]string{"user_rid": userId.String()})
	if err != nil {
		return err
	}

	log.Printf("identificationAttempt result: %v", string(identificationAttempt))

	return render(c, screens.WaitForAuthentication())
}

func (r *IdentificationView) HandleResult(c echo.Context) error {
	identificationAttempt := &model.IdentificationAttempt{}
	var err error
	for identificationAttempt.Used {
		identificationAttempt, err = r.server.IdentificationService.GetLatestIdentificationAttempt(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	identificationAttempt, err = r.server.IdentificationService.UpdateIdentificationAttemptUsed(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	log.Printf("identificationAttempt: %v, %v", identificationAttempt.ID, identificationAttempt.Identified)

	return render(c, screens.Result(identificationAttempt.Identified))
}

// api
func (r *IdentificationView) HandleCreateIdentificationAttempt(c echo.Context) error {
	log.Println("identificationAttempt")

	_, err := r.server.IdentificationService.CreateIdentificationAttempt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", "/identification/identificationPending")

	return c.NoContent(http.StatusCreated)
}
