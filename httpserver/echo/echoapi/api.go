package echoapi

import (
	"github.com/labstack/echo/v4"

	"github.com/gopherd/components/httpserver/http/httpapi"
)

// Component defines the http server component API
type Component interface {
	httpapi.Component

	// Engine returns the underlying echo engine
	Engine() *echo.Echo
}
