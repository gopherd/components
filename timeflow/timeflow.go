// Package timeflow provides functionality for managing time offsets and adjustments.
package timeflow

import (
	"context"
	"net/http"
	"path"
	"sync/atomic"
	"time"

	"github.com/gopherd/core/component"

	"github.com/gopherd/components/httpserver/http/httpapi"
	"github.com/gopherd/components/timeflow/timeflowapi"
)

// Name is the unique identifier for the timeflow component.
const Name = "github.com/gopherd/components/timeflow"

func init() {
	component.Register(Name, func() component.Component {
		return &timeflowComponent{}
	})
}

// Ensure timeflowComponent implements timeflowapi.Component interface.
var _ timeflowapi.Component = (*timeflowComponent)(nil)

// Options defines the configuration options for the timeflow component.
type Options struct {
	InitialOffset time.Duration
	HTTPPath      string
}

// timeflowComponent implements the timeflow functionality.
type timeflowComponent struct {
	component.BaseComponentWithRefs[Options, struct {
		HTTPServer component.OptionalReference[httpapi.Component]
	}]
	offset atomic.Int64 // Stores nanosecond-level offset
}

// Init initializes the TimeFlowComponent with the provided context.
func (c *timeflowComponent) Init(ctx context.Context) error {
	c.offset.Store(int64(c.Options().InitialOffset))
	return nil
}

func (c *timeflowComponent) Start(ctx context.Context) error {
	if server := c.Refs().HTTPServer.Component(); server != nil {
		if root := c.Options().HTTPPath; root != "" {
			c.Logger().Info(
				"register HTTP handler",
				"get", path.Join(root, "/get"),
				"set", path.Join(root, "/set"),
			)
			server.HandleFunc([]string{http.MethodGet}, path.Join(root, "/get"), c.handleGetOffset)
			server.HandleFunc([]string{http.MethodPost}, path.Join(root, "/set"), c.handleSetOffset)
		}
	}
	return nil
}

func (c *timeflowComponent) handleGetOffset(w http.ResponseWriter, r *http.Request) {
	offset := c.offset.Load()
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(time.Duration(offset).String() + "\n"))
}

func (c *timeflowComponent) handleSetOffset(w http.ResponseWriter, r *http.Request) {
	offset := r.FormValue("offset")
	if offset == "" {
		http.Error(w, "missing offset", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(offset)
	if err != nil {
		http.Error(w, "invalid offset", http.StatusBadRequest)
		return
	}
	c.offset.Store(int64(d))
	c.Logger().Info("set time offset", "offset", d)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(d.String() + "\n"))
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
