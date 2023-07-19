package handler

import (
	"net/http"
	"strings"

	"github.com/3ssalunke/vio/vio/config"
	"github.com/3ssalunke/vio/vio/entity"
	"github.com/labstack/echo/v4"
)

func (h *VioHandler) GetMeta(c echo.Context) error {
	names := strings.Split(c.Param("name"), ",")
	var files []*entity.ProcessingFile

	for _, name := range names {
		file, err := h.Storge.GetFile("orig", name)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		files = append(files, file)
	}

	return c.Render(http.StatusOK, "meta.html", map[string]interface{}{
		"files":  files,
		"host":   c.Request().URL,
		"blocks": config.App.Meta.Blocks,
	})
}
