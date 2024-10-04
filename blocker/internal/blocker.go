package internal

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/gopherd/core/component"

	"github.com/gopherd/components/blocker"
)

func init() {
	component.Register(blocker.Name, func() component.Component {
		return &BlockerComponent{}
	})
}

// BlockerComponent implements the component.Component interface to block
// process exit on specific signals.
type BlockerComponent struct {
	component.BaseComponent[blocker.Options]
	signals []os.Signal
	sigChan chan os.Signal
	wg      sync.WaitGroup
}

// Init initializes the blockexitComponent.
func (c *BlockerComponent) Init(ctx context.Context) error {
	c.signals = []os.Signal{os.Interrupt}
	c.sigChan = make(chan os.Signal, 1)
	return nil
}

// Uninit performs cleanup for the blockexitComponent.
func (c *BlockerComponent) Uninit(ctx context.Context) error {
	close(c.sigChan)
	return nil
}

// Start begins listening for signals and blocks until a signal is received
// or the context is cancelled.
func (c *BlockerComponent) Start(ctx context.Context) error {
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
