package httpserverapi

import "github.com/labstack/echo/v4"

// Component defines the http server component API
type Component interface {
	// Engine returns the underlying echo engine
	Engine() *echo.Echo
}
