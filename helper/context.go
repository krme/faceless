package helper

import (
	"context"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ContextKey string

const (
	UrlKey ContextKey = "url"
	// url params
	UserRIDKey       ContextKey = "userRid"
	UserEmailKey     ContextKey = "userEmail"
	ProjectRidKey    ContextKey = "projectRid"
	DatamodelRidKey  ContextKey = "datamodelRid"
	DatamodelKeyKey  ContextKey = "datamodelKey"
	DatamodelTypeKey ContextKey = "datamodelType"
	EnumRidKey       ContextKey = "enumRid"
	DataRidsKey      ContextKey = "dataRid"
	// query params
	ViewTypeKey ContextKey = "viewType"
	ValueKey    ContextKey = "value"
	// request context
	RequestContextKey ContextKey = "requestContextKey"
)

func SetContext(c echo.Context, key ContextKey, value any) {
	ctx := context.WithValue(c.Request().Context(), key, value)
	c.SetRequest(c.Request().WithContext(ctx))
}

func GetCurrentUserRID(c context.Context) uuid.UUID {
	userRid, ok := c.Value(UserRIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}
	return userRid
}

func GetCurrentDatamodelType(c context.Context) string {
	datamodelType, ok := c.Value(DatamodelTypeKey).(string)
	if !ok {
		return ""
	}
	return datamodelType
}

type RequestContext struct {
	Url          *url.URL
	HxRequest    bool
	ViewType     string
	Value        string
	ProjectRID   uuid.UUID
	DatamodelRID uuid.UUID
	DatamodelKey string
	DataRIDs     uuid.UUIDs
}

func GetRequestContext(c context.Context) RequestContext {
	value, ok := c.Value(RequestContextKey).(RequestContext)
	if !ok {
		return RequestContext{}
	}
	return value
}

func RIDIsEmpty(in uuid.UUID) bool {
	return in.String() == uuid.UUID{}.String()
}
