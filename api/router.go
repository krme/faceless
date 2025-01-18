package api

import (
	"context"
	"fmt"
	"ht/helper"
	"ht/server"
	"ht/web/handler"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/time/rate"
)

type Router struct {
	echo   *echo.Echo
	server *server.Server
}

func StartServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	server, err := server.NewServer()
	if err != nil {
		log.Fatal(err.Error())
	}

	echo := echo.New()
	router := &Router{
		echo:   echo,
		server: server,
	}
	router.RegisterRoutes()

	echo.HTTPErrorHandler = handler.HandleErrorView
	echo.Logger.SetLevel(log.DEBUG)
	echo.Logger.Fatal(
		echo.Start(fmt.Sprintf(":%v", helper.GetEnvVariable("PORT"))),
	)

	<-ctx.Done()
	router.server.SessionStore.Close()
	router.server.SessionStore.StopCleanup(router.server.SessionStore.Cleanup(time.Minute * 5))
	fmt.Println("Server stopped")
}

func (r *Router) RegisterRoutes() {
	m := NewMiddleware(r.server)

	authView := handler.NewAuthView(r.server)
	userView := handler.NewUserView(r.server)

	r.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(20),
	)))
	// TODO remove csrf.Secure(false) in production
	csrfMiddleware := csrf.Protect(m.csrfKey, csrf.Path("/"), csrf.Secure(false), csrf.ErrorHandler(http.HandlerFunc(handler.HandleCSRFErrorView)))
	r.echo.Use(echo.WrapMiddleware(csrfMiddleware))
	r.echo.Use(middleware.Recover())
	// r.echo.Use(m.ThrottleMiddleware)
	// r.echo.Use(middleware.Logger())

	// view
	r.echo.GET("/", handler.HandleRegisterView)
	r.echo.GET("/register", handler.HandleRegisterView)
	r.echo.GET("/verifyEmail", handler.HandleVerifyEmailView)
	r.echo.GET("/login", handler.HandleLoginView)
	r.echo.GET("/forgotPassword", handler.HandleForgotPasswordView)
	r.echo.GET("/resetPassword", handler.HandleResetPasswordView)

	// api
	r.echo.POST("/auth/registerWithEmail", authView.HandleRegisterWithEmail)
	r.echo.POST("/auth/requestNewEmailVerificationCode", m.AuthMiddlewareUnverified(authView.HandleRequestNewEmailVerificationCode))
	r.echo.POST("/auth/verifyEmail", m.AuthMiddlewareUnverified(authView.HandleVerifyEmail))
	r.echo.POST("/auth/loginWithEmail", authView.HandleLoginWithEmail)
	r.echo.POST("/auth/requestPasswordReset", authView.HandleRequestPasswordReset)
	r.echo.POST("/auth/resetPassword", m.AuthMiddlewareUnverified(authView.HandleResetPassword))
	r.echo.POST("/auth/logout", authView.HandleLogout)

	// view
	r.echo.GET("/account", m.ViewAuthMiddleware(userView.HandleUser))

	r.echo.RouteNotFound("/*", handler.HandleNotFound)

	r.echo.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	r.echo.Static("/static/", "./web/static")
}
