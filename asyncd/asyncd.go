// Package asyncd provides asynchronous event processing functionality.
package asyncd

import (
	"errors"
	"reflect"

	"github.com/gopherd/core/event"
)

// Name is the fully qualified package name.
const Name = "github.com/gopherd/components/asyncd"

// Options represents the configuration options for the asyncd component.
type Options struct {
	// LockThread determines if the consumer goroutine should be bound to an OS thread.
	LockThread bool `json:"lock_thread"`

	// MaxSize is the maximum number of requests allowed in the queue.
	// Requests exceeding this limit will be discarded.
	MaxSize int `json:"max_size"`
}

// DefaultOptions returns the default configuration for the asyncd component.
func DefaultOptions() Options {
	return Options{
		MaxSize: 1 << 20, // 1,048,576
	}
}

var (
	// ErrFull is returned when the asyncd queue is at capacity.
	ErrFull = errors.New("asyncd: queue is at capacity")

	// ErrNotRunning is returned when trying to send an event to a non-running component.
	ErrNotRunning = errors.New("asyncd: component is not running")
)

// Event is an exported alias for event.Event[reflect.Type].
type Event = event.Event[reflect.Type]

// Component defines the interface for the asyncd component.
type Component interface {
	// On registers a listener for events in the component.
	// It returns an event.ID that can be used to unregister the listener.
	On(listener event.Listener[reflect.Type]) event.ID

	// Send submits an event to the component for processing.
	// It returns an error if the event cannot be sent.
	Send(Event) error
}
