package asyncdapi

import (
	"reflect"

	"github.com/gopherd/core/event"
)

// Component defines the interface for the asyncd component.
type Component interface {
	// On registers a listener for events in the component.
	// It returns an event.ID that can be used to unregister the listener.
	On(listener event.Listener[reflect.Type]) event.ID

	// Send submits an event to the component for processing.
	// It returns an error if the event cannot be sent.
	Send(event.Event[reflect.Type]) error
}
