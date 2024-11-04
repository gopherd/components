package internal

import (
	"context"
	"net/http"
	"path"
	"sync/atomic"
	"time"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/typing"

	"github.com/gopherd/components/httpserver"
	"github.com/gopherd/components/timeflow"
)

func init() {
	component.Register(timeflow.Name, func() component.Component {
		return &TimeFlowComponent{}
	})
}

// Ensure TimeFlowComponent implements timeflow.Component interface.
var _ timeflow.Component = (*TimeFlowComponent)(nil)

// TimeFlowComponent implements the timeflow functionality.
type TimeFlowComponent struct {
	component.BaseComponentWithRefs[timeflow.Options, struct {
		HTTPServer component.OptionalReference[httpserver.Component]
	}]
	offset atomic.Int64 // Stores nanosecond-level offset
}

// Init initializes the TimeFlowComponent with the provided context.
func (c *TimeFlowComponent) Init(ctx context.Context) error {
	c.offset.Store(int64(c.Options().InitialOffset))
	return nil
}

func (c *TimeFlowComponent) Start(ctx context.Context) error {
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

func (c *TimeFlowComponent) handleGetOffset(w http.ResponseWriter, r *http.Request) {
	offset := c.offset.Load()
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(time.Duration(offset).String() + "\n"))
}

func (c *TimeFlowComponent) handleSetOffset(w http.ResponseWriter, r *http.Request) {
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
func (c *TimeFlowComponent) Offset() typing.Duration {
	return typing.Duration(c.offset.Load())
}

// SetOffset sets a new time offset.
func (c *TimeFlowComponent) SetOffset(duration typing.Duration) {
	c.offset.Store(int64(duration))
}

// Now returns the current time adjusted by the offset.
func (c *TimeFlowComponent) Now() time.Time {
	return time.Now().Add(time.Duration(c.offset.Load()))
}

// Adjust applies the current offset to the given time.
func (c *TimeFlowComponent) Adjust(t time.Time) time.Time {
	return t.Add(time.Duration(c.offset.Load()))
}
