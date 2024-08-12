// Package timeflowapi defines the interface for time flow management components.
package timeflowapi

import "time"

// Component represents a time flow management component.
// It provides methods to manage time offsets and adjustments.
type Component interface {
	// Offset returns the current time offset.
	Offset() time.Duration

	// SetOffset sets a new time offset.
	SetOffset(duration time.Duration)

	// Now returns the current time adjusted by the component's offset.
	Now() time.Time

	// Adjust applies the component's offset to the given time.
	Adjust(t time.Time) time.Time
}
