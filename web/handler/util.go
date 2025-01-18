package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func render(ctx echo.Context, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(http.StatusOK, buf.String())
}

func renderPopup(c echo.Context, component templ.Component) error {
	c.Response().Header().Add("HX-Retarget", "#global-popup")
	c.Response().Header().Add("HX-Reswap", "afterend")
	return render(c, component)
}

func renderHTTP(writer http.ResponseWriter, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(context.Background(), buf); err != nil {
		return err
	}

	writer.WriteHeader(http.StatusOK)
	fmt.Fprint(writer, buf.String())
	return nil
}

func renderPopupHTTP(writer http.ResponseWriter, component templ.Component) error {
	writer.Header().Add("HX-Retarget", "#global-popup")
	writer.Header().Add("HX-Reswap", "afterend")
	return renderHTTP(writer, component)
}
