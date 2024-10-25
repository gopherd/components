package internal

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/gopherd/core/component"

	"github.com/gopherd/components/blocker"
	"github.com/gopherd/components/httpserver"
)

func init() {
	component.Register(blocker.Name, func() component.Component {
		return &BlockerComponent{}
	})
}

// BlockerComponent implements the component.Component interface to block
// process exit on specific signals.
type BlockerComponent struct {
	component.BaseComponentWithRefs[blocker.Options, struct {
		HTTPServer component.OptionalReference[httpserver.Component]
	}]
	signals []os.Signal
	sigChan chan os.Signal
	wg      sync.WaitGroup
}

// Init initializes the blockexitComponent.
func (c *BlockerComponent) Init(ctx context.Context) error {
	c.signals = []os.Signal{os.Interrupt, os.Kill}
	c.sigChan = make(chan os.Signal, 1)
	if server := c.Refs().HTTPServer.Component(); server != nil {
		httpPath := c.Options().HTTPPath
		if httpPath == "" {
			httpPath = "/blocker"
		}
		if httpPath[0] != '/' {
			httpPath = "/" + httpPath
		}
		server.HandleFunc([]string{"POST"}, path.Join(httpPath, "stop"), c.stopHandler)
		server.HandleFunc([]string{"POST"}, path.Join(httpPath, "kill"), c.killHandler)
	}
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

func (c *BlockerComponent) stopHandler(w http.ResponseWriter, r *http.Request) {
	c.Logger().Info("Received stop request")
	select {
	case c.sigChan <- os.Interrupt:
		io.WriteString(w, "OK")
	case <-time.After(5 * time.Second):
		io.WriteString(w, "Timeout")
	}
}

func (c *BlockerComponent) killHandler(w http.ResponseWriter, r *http.Request) {
	c.Logger().Info("Received kill request")
	select {
	case c.sigChan <- os.Kill:
		io.WriteString(w, "OK")
	case <-time.After(5 * time.Second):
		io.WriteString(w, "Timeout")
	}
}
