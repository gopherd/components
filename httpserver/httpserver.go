package httpserver

import (
	"github.com/labstack/echo/v4"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/httpserver"

// Options represents the options of the component.
type Options struct {
	Addr string `json:"addr"`
}

// Component defines the http server component API
type Component interface {
	// Engine returns the underlying echo engine
	Engine() *echo.Echo
}
