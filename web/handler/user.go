package handler

import (
	"encoding/json"
	"fmt"
	"ht/helper"
	"ht/model"
	"ht/server"
	"ht/web/view/screens"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type UserView struct {
	server *server.Server
}

func NewUserView(server *server.Server) *UserView {
	newUserView := &UserView{
		server: server,
	}
	return newUserView
}

func (r *UserView) HandleUser(c echo.Context) error {
	return render(c, screens.User(&model.User{CreatedAt: time.Now()}))
}

func (r *UserView) HandleOnboardingStart(c echo.Context) error {
	return render(c, screens.OnboardingStart())
}

func (r *UserView) HandleOnboardingRecording(c echo.Context) error {
	currentStepString := c.Param("step")
	currentStep, err := strconv.Atoi(currentStepString)
	if err != nil {
		return err
	}

	bytes, err := helper.StartJob(fmt.Sprintf("http://localhost:%v/jobs/createSentence", r.server.JobsPort), map[string]string{})
	if err != nil {
		return err
	}

	sentence := ""
	err = json.Unmarshal(bytes, &sentence)
	if err != nil {
		return err
	}

	return render(c, screens.OnboardingRecording(sentence, currentStep))
}

func (r *UserView) HandleOnboardingSuccess(c echo.Context) error {
	return render(c, screens.OnboardingSuccess())
}

func (r *UserView) Identify(c echo.Context) error {
	bytes, err := helper.StartJob(fmt.Sprintf("http://localhost:%v/jobs/createSentence", r.server.JobsPort), map[string]string{})
	if err != nil {
		return err
	}

	sentence := ""
	err = json.Unmarshal(bytes, &sentence)
	if err != nil {
		return err
	}

	return render(c, screens.Identify(sentence))
}

func (r *UserView) HandleShowResultPage(c echo.Context) error {
	return render(c, screens.ShowResultReady())
}

func (r *UserView) HandleShowResultSuccess(c echo.Context) error {
	return render(c, screens.ShowResultSuccess())
}

func (r *UserView) HandleShowResultFailure(c echo.Context) error {
	return render(c, screens.ShowResultFailure())
}

func (r *UserView) HandleAuthenticationWaiting(c echo.Context) error {
	return render(c, screens.WaitForAuthentication())
}

// api
func (r *UserView) HandleCreateReferenceRecording(c echo.Context) error {
	currentStepString := c.Param("step")
	currentStep, err := strconv.Atoi(currentStepString)
	if err != nil {
		return err
	}

	_, err = r.server.UserService.CreateReferenceRecording(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", fmt.Sprintf("/user/onboardingRecording/%v", currentStep+1))

	return c.NoContent(http.StatusCreated)
}
