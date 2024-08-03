package component

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/log"

	"github.com/gopherd/components/logger"
)

func init() {
	component.Register(logger.ComponentName, func() component.Component {
		return &loggerComponent{}
	})
}

type loggerComponent struct {
	component.BaseComponent[logger.Options]
	provider log.Provider
}

func (com *loggerComponent) Init(ctx context.Context) error {
	f := log.Lookup(com.Options().Name)
	if f == nil {
		return fmt.Errorf("unknown log provider: %s", com.Options().Name)
	}
	p, err := f(com.Options().Options)
	if err != nil {
		slog.Error("failed to create log provider", slog.Any("error", err))
		return err
	}
	com.provider = p
	slog.SetDefault(slog.New(com.provider.Handler()))
	return nil
}

func (com *loggerComponent) Uninit(ctx context.Context) error {
	return com.provider.Close()
}
