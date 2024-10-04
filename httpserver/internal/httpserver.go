package http

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gopherd/core/component"

	"github.com/gopherd/components/httpserver"
)

func init() {
	// Register the HTTPServerComponent implementation.
	component.Register(httpserver.Name, func() component.Component {
		return new(HTTPServerComponent)
	})
}

// Ensure HTTPServerComponent implements httpserver.Component interface.
var _ httpserver.Component = (*HTTPServerComponent)(nil)

type HTTPServerComponent struct {
	component.BaseComponent[httpserver.Options]
	mux    *http.ServeMux
	server *http.Server
}

// Init implements the component.Component interface.
func (c *HTTPServerComponent) Init(ctx context.Context) error {
	if c.Options().NewMux {
		c.mux = http.NewServeMux()
	} else {
		c.mux = http.DefaultServeMux
	}
	return nil
}

// Start implements the component.Component interface.
func (c *HTTPServerComponent) Start(ctx context.Context) error {
	addr := cmp.Or(c.Options().Addr, ":http")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	c.server = &http.Server{Addr: addr, Handler: c.mux}
	if c.Options().Block {
		c.Logger().Info("http server started", "addr", addr)
		return c.server.Serve(ln)
	}
	go func() {
		c.Logger().Info("http server started", "addr", addr)
		c.server.Serve(ln)
	}()
	return nil
}

// Shutdown implements the component.Component interface.
func (c *HTTPServerComponent) Shutdown(ctx context.Context) error {
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

// Handle implements the HTTPServerComponent interface.
func (c *HTTPServerComponent) Handle(methods []string, path string, handler http.Handler) {
	c.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if len(methods) == 0 {
			handler.ServeHTTP(w, r)
			return
		}
		for _, method := range methods {
			if r.Method == method {
				handler.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
}

// HandleFunc implements the HTTPServerComponent interface.
func (c *HTTPServerComponent) HandleFunc(methods []string, path string, handler http.HandlerFunc) {
	c.Handle(methods, path, handler)
}
