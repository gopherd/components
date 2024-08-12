package httpapi

import (
	"net/http"
)

// Component is the interface that wraps the Handle method.
type Component interface {
	// Handle registers the handler for the given pattern and methods.
	// If method is empty, it registers the handler for all methods.
	Handle(methods []string, path string, handler http.Handler)

	// HandleFunc registers the handler function for the given pattern and methods.
	// If method is empty, it registers the handler for all methods.
	HandleFunc(methods []string, path string, handler http.HandlerFunc)
}
