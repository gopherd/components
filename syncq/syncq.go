// Package syncq provides a flexible event handling system.
package syncq

import (
	"context"
	"reflect"
	"sync"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
)

const Name = "github.com/gopherd/components/syncq"

func init() {
	component.Register(Name, func() component.Component {
		return new(SyncqComponent[reflect.Type])
	})
}

// Options defines the configuration options for the syncq component.
type Options struct {
	// Unordered determines if listeners should be invoked in the order they are added.
	Unordered bool

	// Concurrent determines if listeners should be invoked concurrently.
	Concurrent bool
}

// Ensure syncqComponent implements event.Dispatcher interface.
var _ event.EventSystem[reflect.Type] = (*SyncqComponent[reflect.Type])(nil)

// SyncqComponent is a component template that provides a flexible event handling system.
type SyncqComponent[T comparable] struct {
	component.BaseComponent[Options]
	eventSystem event.EventSystem[T]
	concurrent  bool
	mu          sync.RWMutex
}

// Init implements component.Component interface.
func (c *SyncqComponent[T]) Init(ctx context.Context) error {
	c.concurrent = c.Options().Concurrent
	c.eventSystem = event.NewEventSystem[T](!c.Options().Unordered)
	return nil
}

// AddListener implements event.ListenerAdder interface.
func (c *SyncqComponent[T]) AddListener(listener event.Listener[T]) event.ListenerID {
	if c.concurrent {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	return c.eventSystem.AddListener(listener)
}

// RemoveListener implements event.ListenerRemover interface.
func (c *SyncqComponent[T]) RemoveListener(id event.ListenerID) bool {
	if c.concurrent {
		c.mu.Lock()
		defer c.mu.Unlock()
	}
	return c.eventSystem.RemoveListener(id)
}

// HasListener implements event.ListenerChecker interface.
func (c *SyncqComponent[T]) HasListener(id event.ListenerID) bool {
	if c.concurrent {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}
	return c.eventSystem.HasListener(id)
}

// DispatchEvent implements event.Dispatcher interface.
func (c *SyncqComponent[T]) DispatchEvent(ctx context.Context, event event.Event[T]) error {
	if c.concurrent {
		c.mu.RLock()
		defer c.mu.RUnlock()
	}
	return c.eventSystem.DispatchEvent(ctx, event)
}
