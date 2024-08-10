package echohttpserverapi

import (
	"github.com/gin-gonic/gin"

	httpapi "github.com/gopherd/components/httpserver/http/api"
)

// Component defines the http server component API
type Component interface {
	httpapi.Component

	// Engine returns the underlying gin engine
	Engine() *gin.Engine
}
