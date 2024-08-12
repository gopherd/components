package http

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gopherd/core/component"

	"github.com/gopherd/components/httpserver/http/httpapi"
)

// Name is the unique identifier for the HTTPServerComponent.
const Name = "github.com/gopherd/component/httpserver/http"

func init() {
	// Register the HTTPServerComponent implementation.
	component.Register(Name, func() component.Component {
		return new(httpServerComponent)
	})
}

// Ensure httpServerComponent implements HTTPServerComponent interface.
var _ httpapi.Component = (*httpServerComponent)(nil)

type httpServerComponent struct {
	component.BaseComponent[struct {
		Addr   string // Addr is the address to listen on.
		Block  bool   // Block indicates whether the Start method should block.
		NewMux bool   // NewMux indicates whether to create a new ServeMux.
	}]
	mux    *http.ServeMux
	server *http.Server
}

// Init implements the component.Component interface.
func (c *httpServerComponent) Init(ctx context.Context) error {
	if c.Options().NewMux {
		c.mux = http.NewServeMux()
	} else {
		c.mux = http.DefaultServeMux
	}
	return nil
}

// Start implements the component.Component interface.
func (c *httpServerComponent) Start(ctx context.Context) error {
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
func (c *httpServerComponent) Shutdown(ctx context.Context) error {
	if c.server != nil {
		return c.server.Shutdown(ctx)
	}
	return nil
}

// Handle implements the HTTPServerComponent interface.
func (c *httpServerComponent) Handle(methods []string, path string, handler http.Handler) {
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
func (c *httpServerComponent) HandleFunc(methods []string, path string, handler http.HandlerFunc) {
	c.Handle(methods, path, handler)
}
