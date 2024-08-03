package component

import (
	"context"
	"log/slog"
	"time"

	"github.com/gopherd/core/component"
	"github.com/labstack/echo/v4"

	"github.com/gopherd/components/httpserver"
)

var _ httpserver.Component = (*httpserverComponent)(nil)

func init() {
	component.Register(httpserver.ComponentName, func() component.Component {
		return &httpserverComponent{}
	})
}

type httpserverComponent struct {
	component.BaseComponent[httpserver.Options]
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
