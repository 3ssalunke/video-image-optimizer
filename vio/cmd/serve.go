package cmd

import (
	"fmt"

	"github.com/3ssalunke/vio/vio/config"
	"github.com/3ssalunke/vio/vio/handler"
	"github.com/3ssalunke/vio/vio/processing"
	"github.com/3ssalunke/vio/vio/storage"
	"github.com/3ssalunke/vio/vio/template"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "`serve` starts server on configured port",
	RunE: func(cmd *cobra.Command, args []string) error {
		e := echo.New()

		h := &handler.VioHandler{
			Processing: processing.Processing{
				Image: processing.NewImageBackend(),
				Video: processing.NewVideoBackend(),
			},
			Storge: storage.NewStorage(storage.NewFileSystemBackend(config.App.Storage.Dir)),
		}

		e.Use(middleware.Recover())
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "${remote_ip} - [${time_rfc3339}] \"${method} ${uri}\" ${status} ${bytes_out} \"-\" \"${user_agent}\" \n",
		}))

		if config.App.Global.MaxUploadSize != "" {
			e.Use(middleware.BodyLimit(config.App.Global.MaxUploadSize))
		}

		e.HTTPErrorHandler = h.ErrorHandler

		e.Renderer = template.NewTemplateRenderer("html")

		e.Static("/static", "static/")
		e.Static("/favicon", "static/favicon")
		e.File("/favicon.ico", "static/favicon/favicon.ico")

		e.GET("/", h.Index)
		e.POST("/upload/multipart/", h.UploadMultipart)
		e.POST("/upload/bytes/", h.UploadBodyBytes)
		e.GET("/meta/:name", h.GetMeta)
		e.GET("/:length/:name", h.GetResizedFile)
		e.GET("/:name", h.GetOriginalFile)

		address := fmt.Sprintf("%s:%d", config.App.Global.Host, config.App.Global.Port)
		return e.Start(address)
	},
}
