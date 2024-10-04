package echo

import (
	"cmp"
	"context"
	"net"
	"net/http"

	"github.com/gopherd/core/component"
	"github.com/labstack/echo/v4"

	"github.com/gopherd/components/echoserver"
)

func init() {
	component.Register(echoserver.Name, func() component.Component {
		return &EchoServerComponent{}
	})
}

// Ensure httpserverComponent implements httpserverapi.Component interface.
var _ echoserver.Component = (*EchoServerComponent)(nil)

type EchoServerComponent struct {
	component.BaseComponent[echoserver.Options]
	engine *echo.Echo
}

func (c *EchoServerComponent) Init(ctx context.Context) error {
	c.engine = echo.New()
	return nil
}

func (c *EchoServerComponent) Start(ctx context.Context) error {
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

func (c *EchoServerComponent) Shutdown(ctx context.Context) error {
	return c.engine.Shutdown(ctx)
}

func (c *EchoServerComponent) Handle(methods []string, path string, h http.Handler) {
	if len(methods) > 0 {
		c.engine.Match(methods, path, echo.WrapHandler(h))
	} else {
		c.engine.Any(path, echo.WrapHandler(h))
	}
}

func (c *EchoServerComponent) HandleFunc(methods []string, path string, h http.HandlerFunc) {
	c.Handle(methods, path, h)
}

func (c *EchoServerComponent) Engine() *echo.Echo {
	return c.engine
}
