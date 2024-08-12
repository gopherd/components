package echo

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gopherd/core/component"
	"github.com/labstack/echo/v4"

	"github.com/gopherd/components/httpserver/echo/echoapi"
)

// Name is the unique identifier for the httpserver component.
const Name = "github.com/gopherd/components/httpserver/echo"

func init() {
	component.Register(Name, func() component.Component {
		return &echoComponent{}
	})
}

// Options defines the configuration options for the httpserver component.
type Options struct {
	Addr  string // Addr is the address to listen on.
	Block bool   // Block indicates whether the Start method should block.
}

// Ensure httpserverComponent implements httpserverapi.Component interface.
var _ echoapi.Component = (*echoComponent)(nil)

type echoComponent struct {
	component.BaseComponent[Options]
	engine *echo.Echo
}

func (c *echoComponent) Init(ctx context.Context) error {
	c.engine = echo.New()
	return nil
}

func (c *echoComponent) Start(ctx context.Context) error {
	addr := cmp.Or(c.Options().Addr, ":http")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	c.engine.Listener = ln

	if c.Options().Block {
		c.Logger().Info("http server started", "addr", addr)
		return c.engine.Start(addr)
	}
	go func() {
		c.Logger().Info("http server started", "addr", addr)
		c.engine.Start(addr)
	}()
	return nil
}

func (c *echoComponent) Shutdown(ctx context.Context) error {
	return c.engine.Shutdown(ctx)
}

func (c *echoComponent) Handle(methods []string, path string, h http.Handler) {
	if len(methods) > 0 {
		c.engine.Match(methods, path, echo.WrapHandler(h))
	} else {
		c.engine.Any(path, echo.WrapHandler(h))
	}
}

func (c *echoComponent) HandleFunc(methods []string, path string, h http.HandlerFunc) {
	c.Handle(methods, path, h)
}

func (c *echoComponent) Engine() *echo.Echo {
	return c.engine
}
