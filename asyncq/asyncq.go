// Package asyncq provides asynchronous event processing functionality.
package asyncq

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/gopherd/core/component"
	"github.com/gopherd/core/event"
	"github.com/gopherd/core/lifecycle"
	"github.com/gopherd/core/op"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/asyncq"

// Options represents the configuration options for the asyncq component.
type Options struct {
	// LockThread determines if the consumer goroutine should be bound to an OS thread.
	LockThread bool

	// MaxSize is the maximum number of requests allowed in the queue.
	// Requests exceeding this limit will be discarded.
	MaxSize int
}

func (o *Options) OnLoaded() error {
	op.SetOr(&o.MaxSize, 1<<20)
	return nil
}

func init() {
	component.Register(Name, func() component.Component {
		return &asyncqComponent{}
	})
}

var (
	// errFull is returned when the asyncq queue is at capacity.
	errFull = errors.New("asyncq: queue is at capacity")

	// errNotRunning is returned when trying to send an event to a non-running component.
	errNotRunning = errors.New("asyncq: component is not running")
)

// Ensure asyncqComponent implements asyncq.Component interface.
var _ event.Dispatcher[reflect.Type] = (*asyncqComponent)(nil)

// asyncqComponent implements the asyncq.Component interface for handling asynchronous events.
type asyncqComponent struct {
	component.BaseComponent[Options]
	dispatcher event.Dispatcher[reflect.Type]

	mutex sync.Mutex
	queue *queue
	cond  *sync.Cond

	status     int32         // Running status
	quit, wait chan struct{} // Channels for shutdown signaling and completion

	maxSizeEver int // Peak number of requests in the queue
}

// Init initializes the asyncq component.
func (c *asyncqComponent) Init(ctx context.Context) error {
	c.dispatcher = event.NewDispatcher[reflect.Type](true)
	c.queue = newQueue(128)
	c.status = int32(lifecycle.Running)
	c.quit = make(chan struct{})
	c.wait = make(chan struct{})
	c.cond = sync.NewCond(&c.mutex)
	go c.run()
	return nil
}

// Uninit shuts down the asyncq component.
func (c *asyncqComponent) Uninit(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&c.status, int32(lifecycle.Running), int32(lifecycle.Stopping)) {
		c.Logger().Error("asyncq component not running")
		return errNotRunning
	}
	close(c.quit)
	c.cond.Signal()
	c.Logger().Info("asyncq component waiting for shutdown")
	<-c.wait
	atomic.StoreInt32(&c.status, int32(lifecycle.Closed))
	return nil
}

// run is the main loop for processing events.
func (c *asyncqComponent) run() {
	options := c.Options()
	c.Logger().Info("asyncq component running", "lockThread", options.LockThread)
	if options.LockThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}
	ctx := context.Background()
	for {
		c.cond.L.Lock()
		for c.queue.size() == 0 {
			c.cond.Wait()
		}
		front := c.queue.pop()
		size := c.queue.size()
		c.cond.L.Unlock()

		if front != nil {
			c.dispatcher.DispatchEvent(ctx, front)
			if size > 0 {
				continue
			}
		}

		select {
		case <-c.quit:
			c.Logger().Info("asyncq component quitting")
			c.clean()
			c.Logger().Info("asyncq component cleanup complete")
			close(c.wait)
			return
		default:
		}
	}
}

// clean processes remaining events in the queue during shutdown.
func (c *asyncqComponent) clean() {
	c.Logger().Info("asyncq component cleaning up")
	ctx := context.Background()
	for {
		c.cond.L.Lock()
		if c.queue.size() == 0 {
			c.cond.L.Unlock()
			break
		}
		front := c.queue.pop()
		c.cond.L.Unlock()

		if front != nil {
			c.dispatcher.DispatchEvent(ctx, front)
		}
	}
}

// AddListener implements the event.Dispatcher interface.
func (c *asyncqComponent) AddListener(listener event.Listener[reflect.Type]) event.ListenerID {
	return c.dispatcher.AddListener(listener)
}

// RemoveListener implements the event.Dispatcher interface.
func (c *asyncqComponent) RemoveListener(id event.ListenerID) bool {
	return c.dispatcher.RemoveListener(id)
}

// HasListener implements the event.Dispatcher interface.
func (c *asyncqComponent) HasListener(id event.ListenerID) bool {
	return c.dispatcher.HasListener(id)
}

// DispatchEvent implements the event.Dispatcher interface.
func (c *asyncqComponent) DispatchEvent(ctx context.Context, e event.Event[reflect.Type]) error {
	if atomic.LoadInt32(&c.status) != int32(lifecycle.Running) {
		c.Logger().Error("asyncq component not running")
		return errNotRunning
	}

	options := c.Options()
	c.mutex.Lock()
	size := c.queue.size()
	if options.MaxSize > 0 && size >= options.MaxSize {
		c.mutex.Unlock()
		c.Logger().Warn(
			"event discarded because the queue is full",
			slog.Int("maxSize", options.MaxSize),
			slog.Any("event", e),
		)
		return errFull
	}

	c.queue.push(e)
	size = c.queue.size()
	oldMaxSizeEver := c.updateMaxSizeEver(size)
	c.mutex.Unlock()

	const warningSizeMask = 1<<15 - 1
	if size == 1 {
		c.cond.Signal()
	} else if size&warningSizeMask == 0 && size > oldMaxSizeEver {
		c.Logger().Warn("queue size reached new peak", "size", size)
	}
	return nil
}

// updateMaxSizeEver updates the peak number of requests in the queue.
func (c *asyncqComponent) updateMaxSizeEver(size int) int {
	maxSizeEver := c.maxSizeEver
	if size > c.maxSizeEver || (size<<1) < c.maxSizeEver {
		c.maxSizeEver = size
	}
	return maxSizeEver
}
