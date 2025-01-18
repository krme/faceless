package handler

import (
	"ht/server"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	server *server.Server
}

func NewAuthHandler(server *server.Server) *AuthHandler {
	newAuthHandler := &AuthHandler{
		server: server,
	}
	return newAuthHandler
}

func (r *AuthHandler) HandleRegisterWithEmail(c echo.Context) error {
	err := r.server.AuthService.RegisterWithEmail(c)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusCreated)
}

func (r *AuthHandler) HandleRequestNewEmailVerificationCode(c echo.Context) error {
	err := r.server.AuthService.RequestNewEmailVerificationCode(c)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "Verification email sent successfully.")
}

func (r *AuthHandler) HandleVerifyEmail(c echo.Context) error {
	err := r.server.AuthService.VerifyEmail(c)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (r *AuthHandler) HandleLoginWithEmail(c echo.Context) error {
	err := r.server.AuthService.LoginWithEmail(c)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (r *AuthHandler) HandleRequestPasswordReset(c echo.Context) error {
	err := r.server.AuthService.RequestPasswordReset(c)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "Reset email sent successfully.")
}

func (r *AuthHandler) HandleResetPassword(c echo.Context) error {
	err := r.server.AuthService.ResetPassword(c)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (r *AuthHandler) HandleLogout(c echo.Context) error {
	err := r.server.AuthService.Logout(c)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
