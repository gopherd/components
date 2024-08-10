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

	eventideapi "github.com/gopherd/components/eventide/api"
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

// DefaultOptions returns the default configuration for the asyncq component.
func DefaultOptions() Options {
	return Options{
		MaxSize: 1 << 20, // 1,048,576
	}
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

// Event is an exported alias for event.Event[reflect.Type].
type Event = event.Event[reflect.Type]

// Ensure asyncqComponent implements asyncq.Component interface.
var _ eventideapi.Component = (*asyncqComponent)(nil)

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
func (com *asyncqComponent) Init(ctx context.Context) error {
	com.dispatcher = event.NewDispatcher[reflect.Type](true)
	com.queue = newQueue(128)
	com.status = int32(lifecycle.Running)
	com.quit = make(chan struct{})
	com.wait = make(chan struct{})
	com.cond = sync.NewCond(&com.mutex)
	go com.run()
	return nil
}

// Uninit shuts down the asyncq component.
func (com *asyncqComponent) Uninit(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&com.status, int32(lifecycle.Running), int32(lifecycle.Stopping)) {
		com.Logger().Error("asyncq component not running")
		return errNotRunning
	}
	close(com.quit)
	com.cond.Signal()
	com.Logger().Info("asyncq component waiting for shutdown")
	<-com.wait
	atomic.StoreInt32(&com.status, int32(lifecycle.Closed))
	return nil
}

// run is the main loop for processing events.
func (com *asyncqComponent) run() {
	options := com.Options()
	com.Logger().Info("asyncq component running", "lockThread", options.LockThread)
	if options.LockThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}
	ctx := context.Background()
	for {
		com.cond.L.Lock()
		for com.queue.size() == 0 {
			com.cond.Wait()
		}
		front := com.queue.pop()
		size := com.queue.size()
		com.cond.L.Unlock()

		if front != nil {
			com.dispatcher.DispatchEvent(ctx, front)
			if size > 0 {
				continue
			}
		}

		select {
		case <-com.quit:
			com.Logger().Info("asyncq component quitting")
			com.clean()
			com.Logger().Info("asyncq component cleanup complete")
			close(com.wait)
			return
		default:
		}
	}
}

// clean processes remaining events in the queue during shutdown.
func (com *asyncqComponent) clean() {
	com.Logger().Info("asyncq component cleaning up")
	ctx := context.Background()
	for {
		com.cond.L.Lock()
		if com.queue.size() == 0 {
			com.cond.L.Unlock()
			break
		}
		front := com.queue.pop()
		com.cond.L.Unlock()

		if front != nil {
			com.dispatcher.DispatchEvent(ctx, front)
		}
	}
}

// On adds an event listener to the component.
func (com *asyncqComponent) On(listener event.Listener[reflect.Type]) event.ListenerID {
	return com.dispatcher.AddListener(listener)
}

// Off removes an event listener from the component.
func (com *asyncqComponent) Off(id event.ListenerID) {
	com.dispatcher.RemoveListener(id)
}

// Emit sends an event to the component for processing.
func (com *asyncqComponent) Emit(_ context.Context, e Event) error {
	if atomic.LoadInt32(&com.status) != int32(lifecycle.Running) {
		com.Logger().Error("asyncq component not running")
		return errNotRunning
	}

	options := com.Options()
	com.mutex.Lock()
	size := com.queue.size()
	if options.MaxSize > 0 && size >= options.MaxSize {
		com.mutex.Unlock()
		com.Logger().Warn(
			"event discarded because the queue is full",
			slog.Int("maxSize", options.MaxSize),
			slog.Any("event", e),
		)
		return errFull
	}

	com.queue.push(e)
	size = com.queue.size()
	oldMaxSizeEver := com.updateMaxSizeEver(size)
	com.mutex.Unlock()

	const warningSizeMask = 1<<15 - 1
	if size == 1 {
		com.cond.Signal()
	} else if size&warningSizeMask == 0 && size > oldMaxSizeEver {
		com.Logger().Warn("queue size reached new peak", "size", size)
	}
	return nil
}

// updateMaxSizeEver updates the peak number of requests in the queue.
func (com *asyncqComponent) updateMaxSizeEver(size int) int {
	maxSizeEver := com.maxSizeEver
	if size > com.maxSizeEver || (size<<1) < com.maxSizeEver {
		com.maxSizeEver = size
	}
	return maxSizeEver
}
