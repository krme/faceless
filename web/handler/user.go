package handler

import (
	"fmt"
	"ht/model"
	"ht/server"
	"ht/web/view/screens"
	"net/http"
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
	// TODO user from db and step from url
	return render(c, screens.OnboardingRecording(&model.User{CreatedAt: time.Now()}, 1))
}

func (r *UserView) HandleOnboardingSuccess(c echo.Context) error {
	return render(c, screens.OnboardingSuccess())
}

func (r *UserView) Test(c echo.Context) error {
	return render(c, screens.Test())
}

func (r *UserView) HandleShowResultPage(c echo.Context) error {
	return render(c, screens.ShowResultReady())
}

func (r *UserView) HandleShowResult(c echo.Context) error {
	return render(c, screens.ShowResult())
}

func (r *UserView) HandleAuthenticationWaiting(c echo.Context) error {
	return render(c, screens.WaitForAuthentication())
}

// api
func (r *UserView) HandleCreateReferenceRecording(c echo.Context) error {
	// TODO get current step from url
	currentStep := 1

	_, err := r.server.UserService.CreateReferenceRecording(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", fmt.Sprintf("/user/onboardingRecording/%v", currentStep+1))

	return c.NoContent(http.StatusCreated)
}
