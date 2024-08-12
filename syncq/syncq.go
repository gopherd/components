// Package syncq provides a flexible event handling system.
package syncq

import (
	"context"
	"reflect"
	"sync"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
	"github.com/gopherd/core/op"
	"github.com/gopherd/core/types"
)

// Name is the unique identifier for the syncq component.
const Name = "github.com/gopherd/components/syncq"

// Options defines the configuration options for the syncq component.
type Options struct {
	// Determines if listeners should be invoked in order of registration.
	// It is true by default.
	Ordered *types.Bool

	// Determines if listeners should be invoked concurrently.
	// It is true by default.
	Concurrent *types.Bool
}

// Ensure syncqComponent implements event.Dispatcher interface.
var _ event.Dispatcher[reflect.Type] = (*syncqComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &syncqComponent{}
	})
}

type syncqComponent struct {
	component.BaseComponent[Options]
	dispatcher event.Dispatcher[reflect.Type]
	concurrent bool
	mu         sync.RWMutex
}

func (c *syncqComponent) Init(ctx context.Context) error {
	c.concurrent = op.IfFunc(c.Options().Concurrent == nil, true, c.Options().Concurrent.Deref)
	c.dispatcher = event.NewDispatcher[reflect.Type](op.IfFunc(c.Options().Ordered == nil, true, c.Options().Ordered.Deref))
	return nil
}

// AddListener implements event.Dispatcher interface.
func (c *syncqComponent) AddListener(listener event.Listener[reflect.Type]) event.ListenerID {
	if c.concurrent {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	return c.dispatcher.AddListener(listener)
}

// RemoveListener implements event.Dispatcher interface.
func (c *syncqComponent) RemoveListener(id event.ListenerID) bool {
	if c.concurrent {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	return c.dispatcher.RemoveListener(id)
}

// HasListener implements event.Dispatcher interface.
func (c *syncqComponent) HasListener(id event.ListenerID) bool {
	if c.concurrent {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}
	return c.dispatcher.HasListener(id)
}

// DispatchEvent implements event.Dispatcher interface.
func (c *syncqComponent) DispatchEvent(ctx context.Context, event event.Event[reflect.Type]) error {
	if c.concurrent {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}
	return c.dispatcher.DispatchEvent(ctx, event)
}
