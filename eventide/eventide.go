// Package eventide provides a flexible event handling system.
package eventide

import (
	"context"
	"reflect"
	"sync"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
	"github.com/gopherd/core/op"
	"github.com/gopherd/core/types"
)

// Name is the unique identifier for the eventide component.
const Name = "github.com/gopherd/components/eventide"

// Options defines the configuration options for the eventide component.
type Options struct {
	Ordered    *types.Bool // Determines if listeners should be invoked in order of registration
	Concurrent *types.Bool // Determines if listeners should be invoked concurrently
}

// Ensure eventideComponent implements eventideapi.Component interface.
var _ event.Dispatcher[reflect.Type] = (*eventideComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &eventideComponent{}
	})
}

type eventideComponent struct {
	component.BaseComponent[Options]
	dispatcher event.Dispatcher[reflect.Type]
	concurrent bool
	mu         sync.RWMutex
}

func (com *eventideComponent) Init(ctx context.Context) error {
	com.concurrent = op.IfFunc(com.Options().Concurrent == nil, true, com.Options().Concurrent.Deref)
	com.dispatcher = event.NewDispatcher[reflect.Type](op.IfFunc(com.Options().Ordered == nil, true, com.Options().Ordered.Deref))
	return nil
}

// AddListener implements event.Dispatcher interface.
func (com *eventideComponent) AddListener(listener event.Listener[reflect.Type]) event.ListenerID {
	if com.concurrent {
		com.mu.Lock()
		defer com.mu.Unlock()
	}
	return com.dispatcher.AddListener(listener)
}

// RemoveListener implements event.Dispatcher interface.
func (com *eventideComponent) RemoveListener(id event.ListenerID) bool {
	if com.concurrent {
		com.mu.Lock()
		defer com.mu.Unlock()
	}
	return com.dispatcher.RemoveListener(id)
}

// HasListener implements event.Dispatcher interface.
func (com *eventideComponent) HasListener(id event.ListenerID) bool {
	if com.concurrent {
		com.mu.RLock()
		defer com.mu.RUnlock()
	}
	return com.dispatcher.HasListener(id)
}

// DispatchEvent implements event.Dispatcher interface.
func (com *eventideComponent) DispatchEvent(ctx context.Context, event event.Event[reflect.Type]) error {
	if com.concurrent {
		com.mu.RLock()
		defer com.mu.RUnlock()
	}
	return com.dispatcher.DispatchEvent(ctx, event)
}
