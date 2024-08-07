// Package eventide provides a flexible event handling system.
package eventide

import (
	"context"
	"reflect"
	"sync"

	eventideapi "github.com/gopherd/components/eventide/api"
	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
)

// Name is the unique identifier for the eventide component.
const Name = "github.com/gopherd/components/eventide"

// Options defines the configuration options for the eventide component.
type Options struct {
	Ordered    bool // Determines if listeners should be invoked in order of registration
	Concurrent bool // Determines if listeners should be invoked concurrently
}

// Ensure eventideComponent implements eventideapi.Component interface.
var _ eventideapi.Component = (*eventideComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &eventideComponent{}
	})
}

type eventideComponent struct {
	component.BaseComponent[Options]
	dispatcher event.Dispatcher[reflect.Type]
	mu         sync.RWMutex
}

func (com *eventideComponent) Init(ctx context.Context) error {
	com.dispatcher = event.NewDispatcher[reflect.Type](com.Options().Ordered)
	return nil
}

func (com *eventideComponent) On(listener event.Listener[reflect.Type]) event.ListenerID {
	if com.Options().Concurrent {
		com.mu.Lock()
		defer com.mu.Unlock()
	}
	return com.dispatcher.AddListener(listener)
}

func (com *eventideComponent) Off(id event.ListenerID) {
	if com.Options().Concurrent {
		com.mu.Lock()
		defer com.mu.Unlock()
	}
	com.dispatcher.RemoveListener(id)
}

func (com *eventideComponent) Emit(ctx context.Context, event event.Event[reflect.Type]) error {
	if com.Options().Concurrent {
		com.mu.RLock()
		defer com.mu.RUnlock()
	}
	return com.dispatcher.DispatchEvent(ctx, event)
}
