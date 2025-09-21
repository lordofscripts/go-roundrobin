/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *				Copyright (C)2025 Muhammad H. Hosseinpour,
 *				Copyright (C)2025 Dídimo Grimaldo T.
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * M. Hadi's enhancements on the original RingQueue with my changes
 * to improve the constructor and implement new IRingQueue methods.
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"context"
	"sync"
	"time"

	"github.com/pion/transport/v3/deadline"
)

/* ----------------------------------------------------------------
 *				I n t e r f a c e s
 *-----------------------------------------------------------------*/

var _ IRingQueue[int] = (*safeRQ[int])(nil)

/* ----------------------------------------------------------------
 *				P r i v a t e	T y p e s
 *-----------------------------------------------------------------*/

type safeRQ[T any] struct {
	rq    *RingQueue[T]
	mutex sync.Mutex

	closed    chan struct{}
	closeOnce sync.Once

	deadline *deadline.Deadline

	whenEmpty WhenEmpty
	available chan struct{}
}

/* ----------------------------------------------------------------
 *				C o n s t r u c t o r s
 *-----------------------------------------------------------------*/

func NewSafeRingQueue[T any](capacity int, whenFull WhenFull, whenEmpty WhenEmpty, onCloseFunc OnCloseCallback[T]) *safeRQ[T] {
	rq := NewRingQueue[T](capacity)
	rq.SetWhenFull(whenFull).SetOnClose(onCloseFunc)
	rq.SetOnClose(onCloseFunc)

	if whenEmpty != WhenEmptyBlock && whenEmpty != WhenEmptyError {
		return nil
	}

	if whenFull != WhenFullOverwrite && whenFull != WhenFullError {
		return nil
	}

	return &safeRQ[T]{
		rq:        rq,
		available: make(chan struct{}, 1),
		deadline:  deadline.New(),
		closed:    make(chan struct{}),
		whenEmpty: whenEmpty,
	}
}

/* ----------------------------------------------------------------
 *				P u b l i c		M e t h o d s
 *-----------------------------------------------------------------*/

func (s *safeRQ[T]) SetPopDeadline(t time.Time) error {
	if s.whenEmpty != WhenEmptyBlock {
		return ErrBadDeadline
	}

	s.deadline.Set(t)

	return nil
}

func (s *safeRQ[T]) SetOnClose(callback OnCloseCallback[T]) IRingQueue[T] {
	s.rq.SetOnClose(callback)
	return s
}

func (s *safeRQ[T]) SetWhenFull(a WhenFull) IRingQueue[T] {
	s.rq.SetWhenFull(a)
	return s
}

func (s *safeRQ[T]) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.rq.Reset()
	s.resetChannel(s.available)
	s.available = make(chan struct{}, 1)
}

// @implement io.Closer
func (s *safeRQ[T]) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.closeOnce.Do(func() {
		close(s.closed)
	})

	return s.rq.Close()
}

// @implement fmt.Stringer
func (s *safeRQ[T]) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.rq.String()
}

func (s *safeRQ[T]) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.rq.Size()
}

func (s *safeRQ[T]) Cap() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.rq.Cap()
}

func (s *safeRQ[T]) Push(element T) (newLen int, err error) {
	newLen, err = s.guardedPush(element)

	if s.whenEmpty == WhenEmptyBlock {
		select {
		case <-s.closed:
			return 0, ErrClosed
		case s.available <- struct{}{}:
			return
		default:
		}
	}

	return
}

func (s *safeRQ[T]) Pop() (elem T, newLen int, err error) {
	elem, newLen, err = s.guardedPop()
	if err == nil {
		return
	}

	// we have an empty queue
	var empty T
	switch s.whenEmpty {
	case WhenEmptyError:
		return empty, 0, ErrEmptyQueue
	case WhenEmptyBlock:
		select {
		case <-s.closed:
			return empty, 0, ErrClosed
		case <-s.available:
			return s.Pop()
		case <-s.deadline.Done():
			return empty, 0, context.DeadlineExceeded
		}
	default:
		panic("unreachable")
	}
}

func (s *safeRQ[T]) Peek() (elem T, len int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.rq.Peek()
}

/* ----------------------------------------------------------------
 *				P r i v a t e	M e t h o d s
 *-----------------------------------------------------------------*/

func (s *safeRQ[T]) guardedPush(element T) (newLen int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newLen, err = s.rq.Push(element)

	return
}

func (s *safeRQ[T]) guardedPop() (elem T, newLen int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	elem, newLen, err = s.rq.Pop()

	return
}

func (s *safeRQ[T]) resetChannel(ch chan struct{}) {
	close(ch)
	// Drain the channel non-blockingly but only attempt
	// to read if channel is not closed (which we have)
	for {
		_, ok := <-ch
		if !ok {
			//fmt.Println("Channel closed.")
			break // Exit the loop if the channel is closed
		}
		//fmt.Println(value) // Process the received value
	}
}
