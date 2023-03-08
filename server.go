package main

import (
	"context"
	"net/http"

	"github.com/firodj/pspsora/internal"
	"github.com/labstack/echo/v4"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func serveCommand(doc *internal.SoraDocument) *ffcli.Command {
	return &ffcli.Command{
		Name: "serve",
		Exec: func(ctx context.Context, args []string) error {
			e := echo.New()
			e.GET("/", func(c echo.Context) error {
				return c.String(http.StatusOK, "Hello, World!")
			})
			return e.Start(":1357")
		},
	}
}
