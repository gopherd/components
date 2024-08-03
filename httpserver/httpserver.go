package httpserver

import (
	"github.com/labstack/echo/v4"
)

const ComponentName = "httpserver"

type Options struct {
	Addr string `json:"addr"`
}

// Component defines the http server component API
type Component interface {
	// Engine returns the underlying echo engine
	Engine() *echo.Echo
}
