/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2023 Serge Toro, and
 *				Copyright (C)2025 Didimo Grimaldo
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * A rune-specific version that does not have the overhead of generics.
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"errors"
	"fmt"
	"time"
)

/* ----------------------------------------------------------------
 *				I n t e r f a c e s
 *-----------------------------------------------------------------*/

var _ IRingQueue[rune] = (*RuneRingQueue)(nil)

/* ----------------------------------------------------------------
 *				P u b l i c		T y p e s
 *-----------------------------------------------------------------*/

/**
 * A specialized version of the generics RingQueue[T]. First born out
 * of curiousity for the benchmark between this and the generic version,
 * and also because I was precisely looking for a rune ring buffer for
 * my pet project.
 */
type RuneRingQueue struct {
	data   []rune // container data of runes
	isFull bool   // disambiguate whether the queue is full or empty
	start  int    // start index (inclusive, i.e. first element)
	end    int    // end index (exclusive, i.e. next after last element)

	whenFull WhenFull
}

/* ----------------------------------------------------------------
 *				C o n s t r u c t o r s
 *-----------------------------------------------------------------*/

/**
 * A specific (non-generic) Ring Queue to hold unicode runes.
 */
func NewRuneRingQueue(capacity int) *RuneRingQueue {
	return &RuneRingQueue{
		data:     make([]rune, capacity),
		isFull:   false,
		start:    0,
		end:      0,
		whenFull: WhenFullError,
	}
}

/* ----------------------------------------------------------------
 *				P u b l i c		M e t h o d s
 *-----------------------------------------------------------------*/

func (r *RuneRingQueue) Reset() {
	r.start = 0
	r.end = 0
	r.isFull = false
	clear(r.data)
}

// @implements fmt.Stringer
func (r *RuneRingQueue) String() string {
	return fmt.Sprintf(
		"[RuneRQ full:%v size:%d start:%d end:%d data:%v]",
		r.isFull,
		len(r.data),
		r.start,
		r.end,
		r.data)
}

func (r *RuneRingQueue) Push(elem rune) (int, error) {
	if r.isFull {
		return r.Size(), ErrFullQueue
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

func (r *RuneRingQueue) Pop() (rune, int, error) {
	var res rune // "zero" element (respective of the type)
	if !r.isFull && r.start == r.end {
		return res, 0, ErrEmptyQueue
	}

	res = r.data[r.start]                 // copy over the first element in the queue
	r.start = (r.start + 1) % len(r.data) // move the start of the queue
	r.isFull = false                      // since we're removing elements, we can never be full

	return res, r.Size(), nil
}

func (r *RuneRingQueue) Peek() (rune, int, error) {
	var res rune // "zero" element (respective of the type)
	if !r.isFull && r.start == r.end {
		return res, 0, ErrEmptyQueue
	}

	return r.data[r.start], r.Size(), nil
}

func (r *RuneRingQueue) Size() int {
	res := r.end - r.start
	if res == 0 && r.isFull {
		res = len(r.data)
	} else if res < 0 {
		res = len(r.data) - res*-1
	}

	return res
}

func (r *RuneRingQueue) Cap() int {
	return len(r.data)
}

func (r *RuneRingQueue) IsFull() bool {
	return r.isFull
}

/**
 * Sets the behaviour when pushing onto a full Ring Queue.
 * It can throw an error or overwrite old data.
 */
func (r *RuneRingQueue) SetWhenFull(a WhenFull) IRingQueue[rune] {
	r.whenFull = a
	return r
}

/**
 * Throws ErrUnsupported. Simply complies with the interface.
 * @implement roundrobin.IRingQueue[rune]
 */
func (r *RuneRingQueue) SetPopDeadline(t time.Time) error {
	return errors.ErrUnsupported
}

/**
 * Does nothing, simply complies with the interface.
 * @implement roundrobin.IRingQueue[rune]
 */
func (r *RuneRingQueue) SetOnClose(callback OnCloseCallback[rune]) IRingQueue[rune] {
	return r
}

/**
 * Does nothing, simply complies with the interface.
 * @implement io.Closer
 */
func (r *RuneRingQueue) Close() error {
	return nil
}
