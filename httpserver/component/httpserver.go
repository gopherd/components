package component

import (
	"context"
	"time"

	"github.com/gopherd/core/component"
	"github.com/gopherd/log"
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

func (com *httpserverComponent) Init(ctx context.Context, entity component.Entity) error {
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
			log.Info().
				String("addr", addr).
				Error("err", err).
				Print("failed to start http server")
			return err
		}
	case <-time.After(1 * time.Second):
		log.Info().String("addr", addr).Print("http server started")
	}
	return nil
}

func (com *httpserverComponent) Shutdown(ctx context.Context) error {
	return com.engine.Shutdown(ctx)
}
