package echoapi

import (
	"github.com/labstack/echo/v4"

	httpapi "github.com/gopherd/components/httpserver/http/api"
)

// Component defines the http server component API
type Component interface {
	httpapi.Component

	// Engine returns the underlying echo engine
	Engine() *echo.Echo
}
