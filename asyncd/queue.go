package asyncd

import (
	"github.com/gopherd/core/math/mathutil"
)

// queue represents an internal event queue with circular buffer implementation.
type queue struct {
	buf []Event
	len int
	cap int
	pos int
	cur int
}

// newQueue creates a new queue with the given capacity.
// The capacity is rounded up to the nearest power of 2.
func newQueue(cap int) *queue {
	cap = int(mathutil.UpperPow2(cap))
	return &queue{
		buf: make([]Event, cap),
		cap: cap,
	}
}

// size returns the current number of events in the queue.
func (q *queue) size() int {
	return q.len
}

// push adds an Event to the queue and returns the new size.
// If the queue is full, it will expand automatically.
func (q *queue) push(e Event) int {
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
func (q *queue) pop() Event {
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
func (q *queue) index(n int) int {
	return n & (q.cap - 1)
}

// expand doubles the capacity of the queue.
func (q *queue) expand() {
	oldCap := q.cap
	newBuf := make([]Event, oldCap*2)

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
