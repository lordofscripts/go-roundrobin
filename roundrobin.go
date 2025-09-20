/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2023 Serge Toro, and
 *				Copyright (C)2025 Muhammad H. Hosseinpour
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * Serge Toro's original RingQueue with M. Hadi's enhancements.
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

/* ----------------------------------------------------------------
 *				I n t e r f a c e s
 *-----------------------------------------------------------------*/

var _ IRingQueue[int] = (*RingQueue[int])(nil)

/* ----------------------------------------------------------------
 *				P u b l i c		T y p e s
 *-----------------------------------------------------------------*/

type RingQueue[T any] struct {
	data   []T  // container data of a generic type T
	isFull bool // disambiguate whether the queue is full or empty
	start  int  // start index (inclusive, i.e. first element)
	end    int  // end index (exclusive, i.e. next after last element)

	// Hadi's enhancements
	whenFull  WhenFull
	closed    bool
	onClose   OnCloseCallback[T]
	closeOnce sync.Once
}

/* ----------------------------------------------------------------
 *				C o n s t r u c t o r s
 *-----------------------------------------------------------------*/

func NewRingQueue[T any](capacity int) *RingQueue[T] {
	return &RingQueue[T]{
		data:   make([]T, capacity),
		isFull: false,
		start:  0,
		end:    0,

		whenFull: WhenFullError,
		closed:   false,
	}
}

/* ----------------------------------------------------------------
 *				P u b l i c		M e t h o d s
 *-----------------------------------------------------------------*/

func (r *RingQueue[T]) SetWhenFull(a WhenFull) IRingQueue[T] {
	r.whenFull = a
	return r
}

func (r *RingQueue[T]) SetOnClose(callback OnCloseCallback[T]) IRingQueue[T] {
	r.onClose = callback
	return r
}

func (r *RingQueue[T]) Reset() {
	r.start = 0
	r.end = 0
	r.isFull = false
	clear(r.data)
}

// @implements fmt.Stringer interface
func (r *RingQueue[T]) String() string {
	return fmt.Sprintf(
		"[RRQ full:%v size:%d start:%d end:%d data:%v]",
		r.isFull,
		len(r.data),
		r.start,
		r.end,
		r.data)
}

func (r *RingQueue[T]) Push(elem T) (int, error) {
	if r.closed {
		return 0, ErrClosed
	}

	if r.isFull {
		switch r.whenFull {
		case WhenFullError:
			return 0, ErrFullQueue

		case WhenFullOverwrite:
			// continue pushing with loss of data
			break

		default:
			return 0, errors.ErrUnsupported
		}
	}

	r.data[r.end] = elem              // place the new element on the available space
	r.end = (r.end + 1) % len(r.data) // move the end forward by modulo of capacity
	r.isFull = r.end == r.start       // check if we're full now

	return r.Size(), nil
}

func (r *RingQueue[T]) Pop() (T, int, error) {
	var res T // "zero" element (respective of the type)
	if r.closed {
		return res, 0, ErrClosed
	}

	if !r.isFull && r.start == r.end {
		return res, 0, ErrEmptyQueue
	}

	res = r.data[r.start]                 // copy over the first element in the queue
	r.start = (r.start + 1) % len(r.data) // move the start of the queue
	r.isFull = false                      // since we're removing elements, we can never be full

	return res, r.Size(), nil
}

func (r *RingQueue[T]) Peek() (T, int, error) {
	var res T // "zero" element (respective of the type)
	if r.closed {
		return res, 0, ErrClosed
	}

	if !r.isFull && r.start == r.end {
		return res, 0, fmt.Errorf("empty queue")
	}

	return r.data[r.start], r.Size(), nil
}

func (r *RingQueue[T]) Size() int {
	if r.closed {
		return 0
	}

	res := r.end - r.start
	if res == 0 && r.isFull {
		res = len(r.data)
	} else if res < 0 {
		res = len(r.data) - res*-1
	}

	return res
}

func (r *RingQueue[T]) Cap() int {
	if r.closed {
		return 0
	}

	return len(r.data)
}

func (r *RingQueue[T]) IsFull() bool {
	if r.closed {
		return false
	}

	return r.isFull
}

func (r *RingQueue[T]) SetPopDeadline(t time.Time) error {
	return errors.ErrUnsupported
}

// @implement io.Closer
func (r *RingQueue[T]) Close() error {
	r.closeOnce.Do(func() {
		r.closed = true
		if r.onClose != nil {
			for (r.end - r.start) != 0 {
				res := r.data[r.start]
				r.start = (r.start + 1) % len(r.data)
				r.onClose(res)
			}
		}
		r.data = nil
	})
	return nil
}
