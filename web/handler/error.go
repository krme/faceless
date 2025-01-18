package handler

import (
	"fmt"
	"ht/web/view/components"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HandleErrorView(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	var message interface{}
	message = err.Error()
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}
	c.Logger().Error(code, err)

	renderPopup(c, components.PopupError("Error", fmt.Sprint(message)))
}

func HandleCSRFErrorView(w http.ResponseWriter, r *http.Request) {
	renderPopupHTTP(w, components.PopupError("Error", "Invalid CSRF token, please reload the page."))
}
