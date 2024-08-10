package httpserver

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gopherd/core/component"

	httpapi "github.com/gopherd/components/httpserver/http/api"
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
func (com *httpServerComponent) Init(ctx context.Context) error {
	if com.Options().NewMux {
		com.mux = http.NewServeMux()
	} else {
		com.mux = http.DefaultServeMux
	}
	return nil
}

// Start implements the component.Component interface.
func (com *httpServerComponent) Start(ctx context.Context) error {
	addr := cmp.Or(com.Options().Addr, ":http")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	com.server = &http.Server{Addr: addr, Handler: com.mux}
	if com.Options().Block {
		com.Logger().Info("http server started", "addr", addr)
		return com.server.Serve(ln)
	}
	go func() {
		com.Logger().Info("http server started", "addr", addr)
		com.server.Serve(ln)
	}()
	return nil
}

// Shutdown implements the component.Component interface.
func (com *httpServerComponent) Shutdown(ctx context.Context) error {
	if com.server != nil {
		return com.server.Shutdown(ctx)
	}
	return nil
}

// Handle implements the HTTPServerComponent interface.
func (com *httpServerComponent) Handle(methods []string, path string, handler http.Handler) {
	com.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
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
func (com *httpServerComponent) HandleFunc(methods []string, path string, handler http.HandlerFunc) {
	com.Handle(methods, path, handler)
}
