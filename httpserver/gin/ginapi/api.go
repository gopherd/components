package ginapi

import (
	"github.com/gin-gonic/gin"

	"github.com/gopherd/components/httpserver/http/httpapi"
)

// Component defines the http server component API
type Component interface {
	httpapi.Component

	// Engine returns the underlying gin engine
	Engine() *gin.Engine
}
