package component

import (
	"context"
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
			com.Entity().Logger().Infof("failed to start http server %s: %v", addr, err)
			return err
		}
	case <-time.After(1 * time.Second):
		com.Entity().Logger().Infof("http server started at %s", addr)
	}
	return nil
}

func (com *httpserverComponent) Shutdown(ctx context.Context) error {
	return com.engine.Shutdown(ctx)
}

func (com *httpserverComponent) Engine() *echo.Echo {
	return com.engine
}
