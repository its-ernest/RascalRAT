package main

import (
	"github.com/its-ernest/RascalRAT/internal/ws"
	"github.com/labstack/echo/v5"
)

func handleDispatchTask(c *echo.Context, hub *ws.Hub) error {
	return ws.DispatchTask(c, hub)
}
