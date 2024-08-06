package httpserver

import (
	"context"
	"log/slog"
	"time"

	"github.com/gopherd/core/component"
	"github.com/labstack/echo/v4"

	httpserverapi "github.com/gopherd/components/httpserver/api"
)

// Name is the unique identifier for the httpserver component.
const Name = "github.com/gopherd/components/httpserver"

// Options defines the configuration options for the httpserver component.
type Options struct {
	Addr string
}

// Ensure httpserverComponent implements httpserverapi.Component interface.
var _ httpserverapi.Component = (*httpserverComponent)(nil)

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
	var errChan = make(chan error)
	var addr = com.Options().Addr
	go func() {
		errChan <- com.engine.Start(addr)
	}()
	select {
	case err := <-errChan:
		if err != nil {
			slog.Error(
				"failed to start http server",
				slog.String("addr", addr),
				slog.Any("error", err),
			)
			return err
		}
	case <-time.After(1 * time.Second):
		slog.Info(
			"http server started",
			slog.String("addr", addr),
		)
	}
	return nil
}

func (com *httpserverComponent) Shutdown(ctx context.Context) error {
	return com.engine.Shutdown(ctx)
}

func (com *httpserverComponent) Engine() *echo.Echo {
	return com.engine
}
