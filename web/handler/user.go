package handler

import (
	"fmt"
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

	return render(c, screens.OnboardingRecording(&model.User{CreatedAt: time.Now()}, currentStep))
}

func (r *UserView) HandleOnboardingSuccess(c echo.Context) error {
	return render(c, screens.OnboardingSuccess())
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
