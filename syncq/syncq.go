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

func (com *syncqComponent) Init(ctx context.Context) error {
	com.concurrent = op.IfFunc(com.Options().Concurrent == nil, true, com.Options().Concurrent.Deref)
	com.dispatcher = event.NewDispatcher[reflect.Type](op.IfFunc(com.Options().Ordered == nil, true, com.Options().Ordered.Deref))
	return nil
}

// AddListener implements event.Dispatcher interface.
func (com *syncqComponent) AddListener(listener event.Listener[reflect.Type]) event.ListenerID {
	if com.concurrent {
		com.mu.Lock()
		defer com.mu.Unlock()
	}
	return com.dispatcher.AddListener(listener)
}

// RemoveListener implements event.Dispatcher interface.
func (com *syncqComponent) RemoveListener(id event.ListenerID) bool {
	if com.concurrent {
		com.mu.Lock()
		defer com.mu.Unlock()
	}
	return com.dispatcher.RemoveListener(id)
}

// HasListener implements event.Dispatcher interface.
func (com *syncqComponent) HasListener(id event.ListenerID) bool {
	if com.concurrent {
		com.mu.RLock()
		defer com.mu.RUnlock()
	}
	return com.dispatcher.HasListener(id)
}

// DispatchEvent implements event.Dispatcher interface.
func (com *syncqComponent) DispatchEvent(ctx context.Context, event event.Event[reflect.Type]) error {
	if com.concurrent {
		com.mu.RLock()
		defer com.mu.RUnlock()
	}
	return com.dispatcher.DispatchEvent(ctx, event)
}
