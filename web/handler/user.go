package handler

import (
	"ht/model"
	"ht/server"
	"ht/web/view/screens"
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
