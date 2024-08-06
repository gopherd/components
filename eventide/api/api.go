package api

import (
	"context"
	"reflect"

	"github.com/gopherd/core/event"
)

// Component defines the interface for the eventide component.
type Component interface {
	// On registers a listener for events in the component.
	// It returns an event.ID that can be used to unregister the listener.
	On(listener event.Listener[reflect.Type]) event.ListenerID

	// Off unregisters a listener for events in the component.
	Off(id event.ListenerID)

	// Emit sends an event to the component for processing.
	Emit(context.Context, event.Event[reflect.Type]) error
}
