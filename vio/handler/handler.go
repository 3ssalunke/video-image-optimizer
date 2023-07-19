package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/3ssalunke/vio/vio/config"
	"github.com/3ssalunke/vio/vio/processing"
	"github.com/3ssalunke/vio/vio/storage"
	"github.com/labstack/echo/v4"
)

type VioHandler struct {
	Processing processing.Processing
	Storge     storage.Storage
}

const (
	SecretCodeKey       = "code"
	SecretCodeCookieTTL = 30 * 24 * time.Hour
)

func (v *VioHandler) checkSecretCode(c echo.Context) (string, error) {
	var code string
	if config.App.Global.SecretCode != "" {
		cookie, err := c.Cookie(SecretCodeKey)
		if err != nil || cookie.Value == "" {
			code = c.QueryParam(SecretCodeKey)
			if code == "" {
				code = c.FormValue(SecretCodeKey)
			}
		} else {
			code = cookie.Value
		}

		if code != config.App.Global.SecretCode {
			return code, errors.New("secret code is invalid")
		}

		newCookie := new(http.Cookie)
		newCookie.Name = SecretCodeKey
		newCookie.Value = code
		newCookie.Expires = time.Now().Add(SecretCodeCookieTTL)
		newCookie.HttpOnly = true
		c.SetCookie(newCookie)
	}

	return code, nil
}
