package handler

import (
	"ht/helper"
	"ht/server"
	"ht/web/view/screens"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthView struct {
	server *server.Server
}

func NewAuthView(server *server.Server) *AuthView {
	newAuthView := &AuthView{
		server: server,
	}
	return newAuthView
}

func HandleRegisterView(c echo.Context) error {
	c.Response().Header().Add("HX-Push-Url", "/register")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return render(c, screens.Register())
}

func HandleVerifyEmailView(c echo.Context) error {
	c.Response().Header().Add("HX-Push-Url", "/verifyEmail")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return render(c, screens.VerifyEmail())
}

func HandleLoginView(c echo.Context) error {
	c.Response().Header().Add("HX-Push-Url", "/login")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return render(c, screens.Login())
}

func HandleForgotPasswordView(c echo.Context) error {
	c.Response().Header().Add("HX-Push-Url", "/forgotPassword")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return render(c, screens.ForgotPassword())
}

func HandleResetPasswordView(c echo.Context) error {
	c.Response().Header().Add("HX-Push-Url", "/resetPassword")
	c.Response().Header().Add("HX-Reswap", "innerHTML")
	return render(c, screens.ResetPassword())
}

func (r *AuthView) HandleAccount(c echo.Context) error {
	auth, err := r.server.AuthService.GetAuth(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	return render(c, screens.Account(auth))
}

// auth handler
func (r *AuthView) HandleRegisterWithEmail(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.RegisterWithEmail(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", "/verifyEmail")

	return c.NoContent(http.StatusCreated)
}

func (r *AuthView) HandleRequestNewEmailVerificationCode(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.RequestNewEmailVerificationCode(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return HandleInfoView(c, "Success", "Verification email sent successfully.")
}

func (r *AuthView) HandleVerifyEmail(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.VerifyEmail(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", "/projects")

	return c.NoContent(http.StatusOK)
}

func (r *AuthView) HandleLoginWithEmail(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.LoginWithEmail(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err)
	}

	c.Response().Header().Add("HX-Redirect", "/projects")

	return c.NoContent(http.StatusOK)
}

func (r *AuthView) HandleRequestPasswordReset(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.RequestPasswordReset(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", "/resetPassword")

	return HandleInfoView(c, "Success", "Reset email sent successfully.")
}

func (r *AuthView) HandleResetPassword(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.ResetPassword(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", "/projects")

	return c.NoContent(http.StatusOK)
}

func (r *AuthView) HandleLogout(c echo.Context) error {
	helper.SetContext(c, helper.ProjectRidKey, uuid.UUID{})
	err := r.server.AuthService.Logout(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	c.Response().Header().Add("HX-Redirect", "/login")

	return c.NoContent(http.StatusCreated)
}
