// Package blockexit provides a component that prevents the process from exiting
// immediately upon receiving a termination signal.
package blockexit

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/gopherd/core/component"
)

// Name is the unique identifier for the blockexit component.
const Name = "github.com/gopherd/components/blockexit"

// Options defines the configuration options for the blockexit component.
// Currently, it's empty but can be extended in the future if needed.
type Options struct{}

func init() {
	component.Register(Name, func() component.Component {
		return &blockexitComponent{}
	})
}

// blockexitComponent implements the component.Component interface to block
// process exit on specific signals.
type blockexitComponent struct {
	component.BaseComponent[Options]
	signals []os.Signal
	sigChan chan os.Signal
	wg      sync.WaitGroup
}

// Init initializes the blockexitComponent.
func (com *blockexitComponent) Init(ctx context.Context) error {
	com.signals = []os.Signal{os.Interrupt}
	com.sigChan = make(chan os.Signal, 1)
	return nil
}

// Uninit performs cleanup for the blockexitComponent.
func (com *blockexitComponent) Uninit(ctx context.Context) error {
	close(com.sigChan)
	return nil
}

// Start begins listening for signals and blocks until a signal is received
// or the context is cancelled.
func (com *blockexitComponent) Start(ctx context.Context) error {
	signal.Notify(com.sigChan, com.signals...)
	com.wg.Add(1)
	go func() {
		defer com.wg.Done()
		select {
		case sig := <-com.sigChan:
			com.Logger().Info("Received signal", "signal", sig.String())
		case <-ctx.Done():
			com.Logger().Info("Context cancelled")
		}
	}()
	com.wg.Wait()
	return nil
}
