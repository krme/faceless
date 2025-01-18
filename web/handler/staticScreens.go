package handler

import (
	"ht/web/view/components"
	"ht/web/view/screens"

	"github.com/labstack/echo/v4"
)

func HandleNotFound(c echo.Context) error {
	return render(c, screens.NotFound())
}

func HandleInfoView(c echo.Context, title string, description string) error {
	return renderPopup(c, components.PopupInfo(title, description))
}
