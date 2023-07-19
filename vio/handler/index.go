package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (v *VioHandler) Index(c echo.Context) error {
	code, codeErr := v.checkSecretCode(c)
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"isAuthorized": codeErr == nil,
		"secretCode":   code,
	})
}
