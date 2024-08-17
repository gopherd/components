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

func init() {
	component.Register(Name, func() component.Component {
		return new(AsyncqComponent[reflect.Type])
	})
}

// Options represents the configuration options for the asyncq component.
type Options struct {
	// LockThread determines if the consumer goroutine should be bound to an OS thread.
	LockThread bool

	// MaxSize is the maximum number of requests allowed in the queue.
	// Requests exceeding this limit will be discarded.
	MaxSize int

	// NumConsumers is the number of consumer goroutines to run concurrently.
	NumConsumers int
}

func (o *Options) OnLoaded() error {
	op.SetOr(&o.MaxSize, 1<<20)
	return nil
}

var (
	// ErrFull is returned when the asyncq queue is at capacity.
	ErrFull = errors.New("asyncq: queue is at capacity")

	// ErrClosed is returned when trying to send an event to a non-running component.
	ErrClosed = errors.New("asyncq: component is closed")
)

// Ensure asyncqComponent implements asyncq.Component interface.
var _ event.EventSystem[reflect.Type] = (*AsyncqComponent[reflect.Type])(nil)

// AsyncqComponent implements the asyncq.Component interface for handling asynchronous events.
type AsyncqComponent[T comparable] struct {
	component.BaseComponent[Options]
	eventSystem event.EventSystem[T]

	mutex sync.Mutex
	queue *queue[T]
	cond  *sync.Cond

	status     int32         // Running status
	quit, wait chan struct{} // Channels for shutdown signaling and completion

	maxSizeEver int // Peak number of requests in the queue
}

// Init initializes the asyncq component.
func (c *AsyncqComponent[T]) Init(ctx context.Context) error {
	c.eventSystem = event.NewEventSystem[T](true)
	c.queue = newQueue[T](128)
	c.status = int32(lifecycle.Running)
	c.quit = make(chan struct{})
	c.wait = make(chan struct{})
	c.cond = sync.NewCond(&c.mutex)
	go c.run()
	return nil
}

// Uninit shuts down the asyncq component.
func (c *AsyncqComponent[T]) Uninit(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&c.status, int32(lifecycle.Running), int32(lifecycle.Stopping)) {
		c.Logger().Error("asyncq component not running")
		return ErrClosed
	}
	close(c.quit)
	c.cond.Signal()
	c.Logger().Info("asyncq component waiting for shutdown")
	<-c.wait
	atomic.StoreInt32(&c.status, int32(lifecycle.Closed))
	return nil
}

// run is the main loop for processing events.
func (c *AsyncqComponent[T]) run() {
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
			c.eventSystem.DispatchEvent(ctx, front)
			if size > 0 {
				continue
			}
		}

		select {
		case <-c.quit:
			c.Logger().Info("asyncq component quiting")
			c.clean()
			c.Logger().Info("asyncq component cleanup complete")
			close(c.wait)
			return
		default:
		}
	}
}

// clean processes remaining events in the queue during shutdown.
func (c *AsyncqComponent[T]) clean() {
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
			c.eventSystem.DispatchEvent(ctx, front)
		}
	}
}

// AddListener implements the event.Dispatcher interface.
func (c *AsyncqComponent[T]) AddListener(listener event.Listener[T]) event.ListenerID {
	return c.eventSystem.AddListener(listener)
}

// RemoveListener implements the event.Dispatcher interface.
func (c *AsyncqComponent[T]) RemoveListener(id event.ListenerID) bool {
	return c.eventSystem.RemoveListener(id)
}

// HasListener implements the event.Dispatcher interface.
func (c *AsyncqComponent[T]) HasListener(id event.ListenerID) bool {
	return c.eventSystem.HasListener(id)
}

// DispatchEvent implements the event.Dispatcher interface.
func (c *AsyncqComponent[T]) DispatchEvent(ctx context.Context, e event.Event[T]) error {
	if atomic.LoadInt32(&c.status) != int32(lifecycle.Running) {
		c.Logger().Error("asyncq component not running")
		return ErrClosed
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
		return ErrFull
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
func (c *AsyncqComponent[T]) updateMaxSizeEver(size int) int {
	maxSizeEver := c.maxSizeEver
	if size > c.maxSizeEver || (size<<1) < c.maxSizeEver {
		c.maxSizeEver = size
	}
	return maxSizeEver
}
