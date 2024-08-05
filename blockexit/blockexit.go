package blockexit

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/gopherd/core/component"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/blockexit"

// Options represents the options of the component.
type Options struct {
}

func init() {
	component.Register(Name, func() component.Component {
		return &blockexitComponent{}
	})
}

type blockexitComponent struct {
	component.BaseComponent[Options]

	signals []os.Signal
	sigChan chan os.Signal
	wg      sync.WaitGroup
}

func (com *blockexitComponent) Init(ctx context.Context) error {
	com.signals = []os.Signal{os.Interrupt}
	com.sigChan = make(chan os.Signal, 1)
	return nil
}

func (com *blockexitComponent) Uninit(ctx context.Context) error {
	close(com.sigChan)
	return nil
}

func (com *blockexitComponent) Start(ctx context.Context) error {
	signal.Notify(com.sigChan, com.signals...)

	com.wg.Add(1)
	go func() {
		defer com.wg.Done()
		select {
		case sig := <-com.sigChan:
			slog.Debug("received signal", slog.String("signal", sig.String()))
		case <-ctx.Done():
			slog.Debug("context cancelled")
		}
	}()
	com.wg.Wait()
	return nil
}
