package echoserver

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gopherd/core/component"
	"github.com/labstack/echo/v4"

	echoapi "github.com/gopherd/components/httpserver/echo/api"
)

// Name is the unique identifier for the httpserver component.
const Name = "github.com/gopherd/components/httpeserver/echo"

// Options defines the configuration options for the httpserver component.
type Options struct {
	Addr  string // Addr is the address to listen on.
	Block bool   // Block indicates whether the Start method should block.
}

// Ensure httpserverComponent implements httpserverapi.Component interface.
var _ echoapi.Component = (*httpserverComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &httpserverComponent{}
	})
}

type httpserverComponent struct {
	component.BaseComponent[Options]
	engine *echo.Echo
}

func (com *httpserverComponent) Init(ctx context.Context) error {
	com.engine = echo.New()
	return nil
}

func (com *httpserverComponent) Start(ctx context.Context) error {
	addr := cmp.Or(com.Options().Addr, ":http")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	com.engine.Listener = ln

	if com.Options().Block {
		com.Logger().Info("http server started", "addr", addr)
		return com.engine.Start(addr)
	}
	go func() {
		com.Logger().Info("http server started", "addr", addr)
		com.engine.Start(addr)
	}()
	return nil
}

func (com *httpserverComponent) Shutdown(ctx context.Context) error {
	return com.engine.Shutdown(ctx)
}

func (com *httpserverComponent) Handle(methods []string, path string, h http.Handler) {
	if len(methods) > 0 {
		com.engine.Match(methods, path, echo.WrapHandler(h))
	} else {
		com.engine.Any(path, echo.WrapHandler(h))
	}
}

func (com *httpserverComponent) HandleFunc(methods []string, path string, h http.HandlerFunc) {
	com.Handle(methods, path, h)
}

func (com *httpserverComponent) Engine() *echo.Echo {
	return com.engine
}
