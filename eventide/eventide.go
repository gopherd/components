// Package eventnexus provides a flexible event handling system.
package eventnexus

import (
	"context"
	"log/slog"
	"reflect"
	"sync"

	eventideapi "github.com/gopherd/components/eventide/api"
	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
)

// Name is the unique identifier for the eventnexus component.
const Name = "github.com/gopherd/components/eventide"

// Options defines the configuration options for the eventnexus component.
type Options struct {
	Ordered bool // Determines if listeners should be invoked in order of registration
}

// Ensure eventNexusComponent implements eventnexusapi.Component interface.
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

func (com *eventideComponent) Start(ctx context.Context) error {
	slog.Info("EventNexus system started")
	return nil
}

func (com *eventideComponent) Shutdown(ctx context.Context) error {
	com.mu.Lock()
	defer com.mu.Unlock()
	com.dispatcher = nil
	slog.Info("EventNexus system shut down")
	return nil
}

func (com *eventideComponent) On(listener event.Listener[reflect.Type]) event.ListenerID {
	com.mu.RLock()
	defer com.mu.RUnlock()
	return com.dispatcher.AddListener(listener)
}

func (com *eventideComponent) Off(id event.ListenerID) {
	com.mu.RLock()
	defer com.mu.RUnlock()
	com.dispatcher.RemoveListener(id)
}

func (com *eventideComponent) Emit(ctx context.Context, event event.Event[reflect.Type]) error {
	com.mu.RLock()
	defer com.mu.RUnlock()
	return com.dispatcher.DispatchEvent(ctx, event)
}
