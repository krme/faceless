package api

import (
	"crypto/rand"
	"fmt"
	"ht/helper"
	"ht/model"
	"ht/server"
	"ht/web/handler"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Middleware struct {
	server  *server.Server
	csrfKey []byte
}

func NewMiddleware(server *server.Server) *Middleware {
	scrfKey := make([]byte, 32)
	n, err := rand.Read(scrfKey)
	if err != nil {
		panic(err)
	}
	if n != 32 {
		panic("unable to read 32 bytes for CSRF key")
	}

	return &Middleware{
		server:  server,
		csrfKey: scrfKey,
	}
}

func (r Middleware) getSession(c echo.Context) (*model.Session, error) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := r.server.SessionStore.Get(c.Request(), "auth")

	userId, userIdExists := session.Values["user_id"]
	authenticated, authenticatedExists := session.Values["authenticated"]
	emailVerified, emailVerifiedExists := session.Values["email_verified"]
	createdAt, createdAtExists := session.Values["created_at"]

	if !userIdExists || !authenticatedExists || !createdAtExists || !emailVerifiedExists {
		session.Values["authenticated"] = false
		session.Options.MaxAge = 0
		err := session.Save(c.Request(), c.Response().Writer)
		if err != nil {
			return nil, fmt.Errorf("error saving session: %v", err)
		}
	}

	currentSession := &model.Session{}
	if userIdString, ok := userId.(string); ok && len(userIdString) > 0 {
		userRid, err := uuid.Parse(userIdString)
		if err != nil {
			return nil, fmt.Errorf("invalid user uuid: %v", err)
		}
		currentSession.UserID = userRid
	}
	if authenticatedBool, ok := authenticated.(bool); ok {
		currentSession.Authenticated = authenticatedBool
	} else {
		return nil, fmt.Errorf("invalid type authenticated: %T", userId)
	}
	if emailVerifiedBool, ok := emailVerified.(bool); ok {
		currentSession.EmailVerified = emailVerifiedBool
	} else {
		return nil, fmt.Errorf("invalid type email_verified: %T", userId)
	}
	if createdAtTime, ok := createdAt.(int64); ok {
		currentSession.CreatedAt = time.Unix(createdAtTime, 0)
	} else {
		return nil, fmt.Errorf("invalid type created_at: %T", userId)
	}

	if time.Now().Add(time.Minute * -60).After(currentSession.CreatedAt) {
		session.Values["authenticated"] = false
		session.Options.MaxAge = 0
		err := session.Save(c.Request(), c.Response().Writer)
		if err != nil {
			return nil, fmt.Errorf("error saving session: %v", err)
		}
	}

	return currentSession, nil
}

func (r Middleware) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := r.getSession(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("error getting session: %v", err))
		}

		if !session.Authenticated {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("not logged in"))
		} else if !session.EmailVerified {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("email not verified"))
		} else {
			helper.SetContext(c, helper.UserRIDKey, session.UserID)
			return next(c)
		}
	}
}

func (r Middleware) AuthMiddlewareUnverified(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := r.getSession(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("error getting session: %v", err))
		}

		helper.SetContext(c, helper.UserRIDKey, session.UserID)
		return next(c)
	}
}

func (r Middleware) ViewAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := r.getSession(c)
		if err != nil {
			log.Printf("error getting session: %v", err)
			return handler.HandleLoginView(c)
		}

		if !session.Authenticated {
			return handler.HandleLoginView(c)
		} else if !session.EmailVerified {
			return handler.HandleVerifyEmailView(c)
		} else {
			helper.SetContext(c, helper.UserRIDKey, session.UserID)
			return next(c)
		}
	}
}

func (r Middleware) ThrottleMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		helper.Throttle()
		return next(c)
	}
}
