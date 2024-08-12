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
func (c *blockexitComponent) Init(ctx context.Context) error {
	c.signals = []os.Signal{os.Interrupt}
	c.sigChan = make(chan os.Signal, 1)
	return nil
}

// Uninit performs cleanup for the blockexitComponent.
func (c *blockexitComponent) Uninit(ctx context.Context) error {
	close(c.sigChan)
	return nil
}

// Start begins listening for signals and blocks until a signal is received
// or the context is cancelled.
func (c *blockexitComponent) Start(ctx context.Context) error {
	signal.Notify(c.sigChan, c.signals...)
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		select {
		case sig := <-c.sigChan:
			c.Logger().Info("Received signal", "signal", sig.String())
		case <-ctx.Done():
			c.Logger().Info("Context cancelled")
		}
	}()
	c.wg.Wait()
	return nil
}
