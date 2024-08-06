package logger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gopherd/core/api/logging"
	"github.com/gopherd/core/component"
	"github.com/gopherd/core/raw"
)

// Name is the unique identifier for the logger component.
const Name = "github.com/gopherd/components/logger"

// Options defines the configuration options for the logger component.
type Options struct {
	// Name is the name of the log provider which is registered by log.Register.
	Name string
	// Options used to create the log provider.
	Options raw.Object
}

// DefaultOptions returns the default options.
func DefaultOptions(modifier func(*Options)) Options {
	options := Options{
		Name: "stderr",
		Options: raw.MustJSON(logging.StdOptions{
			AddSource: true,
		}),
	}
	if modifier != nil {
		modifier(&options)
	}
	return options
}

func init() {
	component.Register(Name, func() component.Component {
		return &loggerComponent{}
	})
}

type loggerComponent struct {
	component.BaseComponent[Options]
	provider logging.Provider
}

func (com *loggerComponent) Init(ctx context.Context) error {
	f := logging.Lookup(com.Options().Name)
	if f == nil {
		return fmt.Errorf("unknown log provider: %s", com.Options().Name)
	}
	p, err := f(com.Options().Options.Bytes())
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
