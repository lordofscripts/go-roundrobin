/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *				Copyright (C)2025 Muhammad H. Hosseinpour,
 *				Copyright (C)2025 Dídimo Grimaldo T.
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * A generic interface to the RingQueue implementations (plain & safe)
 * offered by this GO module.
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"fmt"
	"io"
	"time"
)

/* ----------------------------------------------------------------
 *						G l o b a l s
 *-----------------------------------------------------------------*/

const ( // what happens when Push() on a full circular buffer
	WhenFullError WhenFull = iota
	WhenFullOverwrite
)

const ( // what happens when Pop() on an empty circular buffer
	WhenEmptyError WhenEmpty = iota
	WhenEmptyBlock
)

var ( // module errors
	ErrFullQueue   = fmt.Errorf("ring buffer is full")
	ErrEmptyQueue  = fmt.Errorf("ring buffer is empty")
	ErrClosed      = fmt.Errorf("ring buffer is closed")
	ErrBadDeadline = fmt.Errorf("deadline only possible for WhenEmptyBlock")
)

/* ----------------------------------------------------------------
 *				I n t e r f a c e s
 *-----------------------------------------------------------------*/

type IRingQueue[T any] interface {
	fmt.Stringer
	io.Closer

	SetPopDeadline(t time.Time) error
	SetWhenFull(a WhenFull) IRingQueue[T]
	SetOnClose(callback OnCloseCallback[T]) IRingQueue[T]

	Size() int
	Cap() int

	Push(element T) (newLen int, err error)
	Pop() (element T, newLen int, err error)
	Peek() (element T, len int, err error)

	Reset()
}

/* ----------------------------------------------------------------
 *				P u b l i c		T y p e s
 *-----------------------------------------------------------------*/

type WhenEmpty int
type WhenFull int

type OnCloseCallback[T any] func(data T)
