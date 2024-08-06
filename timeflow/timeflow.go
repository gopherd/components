// Package timeflow provides functionality for managing time offsets and adjustments.
package timeflow

import (
	"context"
	"sync/atomic"
	"time"

	timeflowapi "github.com/gopherd/components/timeflow/api"
	"github.com/gopherd/core/component"
)

// Name is the unique identifier for the timeflow component.
const Name = "github.com/gopherd/components/timeflow"

// Options defines the configuration options for the timeflow component.
type Options struct {
	InitialOffset time.Duration
}

// Ensure timeflowComponent implements timeflowapi.Component interface.
var _ timeflowapi.Component = (*timeflowComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &timeflowComponent{}
	})
}

// timeflowComponent implements the timeflow functionality.
type timeflowComponent struct {
	component.BaseComponent[Options]
	offset atomic.Int64 // Stores nanosecond-level offset
}

// Init initializes the TimeFlowComponent with the provided context.
func (c *timeflowComponent) Init(ctx context.Context) error {
	c.offset.Store(int64(c.Options().InitialOffset))
	return nil
}

// Offset returns the current time offset.
func (c *timeflowComponent) Offset() time.Duration {
	return time.Duration(c.offset.Load())
}

// SetOffset sets a new time offset.
func (c *timeflowComponent) SetOffset(duration time.Duration) {
	c.offset.Store(int64(duration))
}

// Now returns the current time adjusted by the offset.
func (c *timeflowComponent) Now() time.Time {
	return time.Now().Add(time.Duration(c.offset.Load()))
}

// Adjust applies the current offset to the given time.
func (c *timeflowComponent) Adjust(t time.Time) time.Time {
	return t.Add(time.Duration(c.offset.Load()))
}
