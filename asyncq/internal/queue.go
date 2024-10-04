package internal

import (
	"github.com/gopherd/core/event"
	"github.com/gopherd/core/math/mathutil"
)

// queue represents an internal event queue with circular buffer implementation.
type queue[T comparable] struct {
	buf []event.Event[T]
	len int
	cap int
	pos int
	cur int
}

// newQueue creates a new queue with the given capacity.
// The capacity is rounded up to the nearest power of 2.
func newQueue[T comparable](cap int) *queue[T] {
	cap = int(mathutil.UpperPow2(cap))
	return &queue[T]{
		buf: make([]event.Event[T], cap),
		cap: cap,
	}
}

// size returns the current number of events in the queue.
func (q *queue[T]) size() int {
	return q.len
}

// push adds an Event to the queue and returns the new size.
// If the queue is full, it will expand automatically.
func (q *queue[T]) push(e event.Event[T]) int {
	if q.len == q.cap {
		q.expand()
	}
	q.buf[q.index(q.cur)] = e
	q.cur++
	q.len++
	return q.len
}

// pop removes and returns the oldest Event from the queue.
// If the queue is empty, it returns nil.
func (q *queue[T]) pop() event.Event[T] {
	if q.len == 0 {
		return nil
	}
	idx := q.index(q.pos)
	v := q.buf[idx]
	q.buf[idx] = nil // Allow GC to reclaim the memory
	q.pos++
	q.len--
	return v
}

// index calculates the actual index in the circular buffer.
func (q *queue[T]) index(n int) int {
	return n & (q.cap - 1)
}

// expand doubles the capacity of the queue.
func (q *queue[T]) expand() {
	oldCap := q.cap
	newBuf := make([]event.Event[T], oldCap*2)

	if q.cur > q.pos {
		copy(newBuf, q.buf[q.pos:q.cur])
	} else {
		n := copy(newBuf, q.buf[q.pos:])
		copy(newBuf[n:], q.buf[:q.cur])
	}

	q.buf = newBuf
	q.cap = len(newBuf)
	q.pos = 0
	q.cur = q.len
}
