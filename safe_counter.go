/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2023 Serge Toro, and
 *				Copyright (C)2025 Muhammad H. Hosseinpour
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * Serge Toro's original RingQueue with M. Hadi's enhancements.
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"sync/atomic"
)

/* ----------------------------------------------------------------
 *				P u b l i c		T y p e s
 *-----------------------------------------------------------------*/

type SafeCounter struct {
	counter int64
}

/* ----------------------------------------------------------------
 *				C o n s t r u c t o r s
 *-----------------------------------------------------------------*/

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{0}
}

/* ----------------------------------------------------------------
 *				P u b l i c		M e t h o d s
 *-----------------------------------------------------------------*/

func (c *SafeCounter) Increment() int64 {
	return atomic.AddInt64(&c.counter, 1)
}

func (c *SafeCounter) Decrement() int64 {
	return atomic.AddInt64(&c.counter, -1)
}

func (c *SafeCounter) Value() int64 {
	return atomic.LoadInt64(&c.counter)
}

func (c *SafeCounter) IsZero() bool {
	return atomic.LoadInt64(&c.counter) == 0
}

func (c *SafeCounter) Clear() {
	atomic.StoreInt64(&c.counter, 0)
}
