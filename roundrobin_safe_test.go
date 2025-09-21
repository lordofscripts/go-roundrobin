/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2025 Muhammad H. Hosseinpour
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * Tests of Hadi's functional enhancements
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

/* ----------------------------------------------------------------
 *						T e s t s
 *-----------------------------------------------------------------*/

func Test_Deadline(t *testing.T) {
	obj := NewSafeRingQueue[int](10, WhenFullError, WhenEmptyBlock, nil)
	timeBefore := time.Now()
	obj.SetPopDeadline(time.Now().Add(1 * time.Second))
	_, _, err := obj.Pop()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded error")
	}
	if time.Since(timeBefore) < 1*time.Second {
		t.Fatalf("Expected 1s timeout")
	}
}

func Test_Deadline2(t *testing.T) {
	obj := NewSafeRingQueue[int](10, WhenFullError, WhenEmptyBlock, nil)
	timeBefore := time.Now()
	for i := 0; i < 10; i++ {
		t.Log("push ", i)
		obj.Push(i)
	}
	for i := 0; i < 10; i++ {
		t.Log("pop ", i)
		obj.Pop()
	}
	obj.SetPopDeadline(time.Now().Add(5 * time.Second))
	_, _, err := obj.Pop()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected DeadlineExceeded error")
	}
	if time.Since(timeBefore) < 5*time.Second {
		t.Fatalf("Expected 1s timeout")
	}
}

func Test_OnClose(t *testing.T) {
	type testCase struct {
		name           string
		pushCount      int
		popCount       int
		onCloseCount   int
		wantErrInClose bool
		wantMismatch   bool
	}
	tests := []testCase{
		{
			name:           "under-push",
			pushCount:      1,
			popCount:       1,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "over-push",
			pushCount:      15,
			popCount:       10,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "over-push and under-pop",
			pushCount:      15,
			popCount:       5,
			onCloseCount:   5,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "under-push and under-pop",
			pushCount:      7,
			popCount:       5,
			onCloseCount:   2,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "over-push and over-pop",
			pushCount:      15,
			popCount:       15,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
		{
			name:           "mismatch",
			pushCount:      10,
			popCount:       0,
			onCloseCount:   0,
			wantErrInClose: false,
			wantMismatch:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onCloseCount := 0
			rr := NewRingQueue[int](10)
			rr.SetWhenFull(WhenFullError)
			rr.SetOnClose(func(data int) {
				onCloseCount++
			})

			for i := 0; i < tt.pushCount; i++ {
				rr.Push(i)
			}
			for i := 0; i < tt.popCount; i++ {
				rr.Pop()
			}
			if err := rr.Close(); (err != nil) != tt.wantErrInClose {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErrInClose)
			}
			if onCloseCount != tt.onCloseCount && !tt.wantMismatch {
				t.Errorf("onCloseCount = %v, wanted onCloseCount %v", onCloseCount, tt.onCloseCount)
			}
		})
	}
}

// If RQ is full and whenFull action is set to WhenFullError, pushing an
// extra value should report an error and leave RQ unmodified.
func Test_PushFull_WhenFullError(t *testing.T) {
	obj := NewSafeRingQueue[int](5, WhenFullError, WhenEmptyBlock, nil)
	for i := range 5 {
		obj.Push(i)
	}

	assertSize(obj, 5, t)

	newSize, err := obj.Push(15) // should return an error and preserve the RingQueue
	assertSize(obj, 5, t)
	if newSize != 5 {
		t.Errorf("pushing on full should report full size. exp: %d got: %d", 5, newSize)
	}

	switch err {
	case nil:
		t.Errorf("pushing onto full buffer with WheFullError should return an error, got nil")
	case ErrFullQueue:
		fmt.Printf("error as expected: %v", err)
	default:
		t.Errorf("unexpected error %v", err)
	}
}

func Test_PushFull_WhenFullOverwrite(t *testing.T) {
	obj := NewSafeRingQueue[int](5, WhenFullOverwrite, WhenEmptyBlock, nil)
	const FIRST_POP_VALUE int = 0
	for i := range 5 { // 0..4
		obj.Push(i)
	}

	assertSize(obj, 5, t)

	// should accept and overwrite but size remains unchanged.
	// there will be no errors. By default, it overwrites oldest data (what we pop)
	const OVERWRITE_DATA int = 15
	newSize, err := obj.Push(OVERWRITE_DATA) // should accept new value and overwrite, no errors
	assertSize(obj, 5, t)
	if newSize != 5 {
		t.Errorf("pushing on full should report full size. exp: %d got: %d", 5, newSize)
	}

	if err != nil {
		t.Errorf("pushing onto full buffer with WheFullOverwrite should NOT return an error, got: %v", err)
	}

	// since it overwrote the oldest data, we should Pop that last push...
	val, newSize, _ := obj.Pop()
	if newSize != 4 {
		t.Errorf("pop did not decrement: exp 4 got %d", newSize)
	}
	if val != OVERWRITE_DATA {
		t.Errorf("pushing onto full buffer with WheFullError should Pop that same value. exp %d got %d", OVERWRITE_DATA, val)
	}
}

// If RQ is empty and whenEmpty action is set to WhenEmptyError, popping a
// value should report an error and leave RQ unmodified.
func Test_PopEmpty_WhenEmptyError(t *testing.T) {
	obj := NewSafeRingQueue[int](3, WhenFullError, WhenEmptyError, nil)
	for i := range 3 {
		obj.Push(i)
	}
	assertSize(obj, 3, t)

	for range 3 {
		obj.Pop()
	}
	assertSize(obj, 0, t)

	value, newLen, err := obj.Pop()
	if value != 0 {
		t.Errorf("pop on empty did not return default value for INT, got %d", value)
	}
	if newLen != 0 {
		t.Errorf("pop on empty should have size 0, got %d", newLen)
	}
	if err != ErrEmptyQueue {
		t.Errorf("pop on empty should return ErrEmptyQueue, got %v", err)
	}
}

// test popping on an empty queue configured for WhenEmptyBlock.
// It will block but we use a synchronization mechanism to test it.
func Test_PopEmpty_WhenEmptyBlock(t *testing.T) {
	const MAX int = 3
	obj := NewSafeRingQueue[int](MAX, WhenFullError, WhenEmptyBlock, nil)
	// Push() to full
	for i := range MAX {
		obj.Push(i)
	}
	assertSize(obj, MAX, t)

	// Pop() all
	for range MAX {
		obj.Pop()
	}
	assertSize(obj, 0, t)

	// Now Pop on empty should block until something available to Pop,
	// there will be none more but we use a timeout.

	done := make(chan struct{})

	go func() {
		obj.Pop()
		close(done)
	}()

	select {
	case <-done:
		t.Error("expected blocking function to remain blocked but it completed!")

	case <-time.After(2 * time.Second):
		// the Pop() function is still blocked, as expected
		fmt.Println("Pop() remains blocked as expected when empty and WhenEmptyBlock")
	}
}

// test popping on an empty queue configured for WhenEmptyBlock.
// It will block but we use a synchronization mechanism to test it.
// But in this variant, we use another Go Routing to push a value
// so that the Pop() unblocks when the new data is available.
func Test_PopEmpty_WhenEmptyBlock2(t *testing.T) {
	const MAX int = 3
	obj := NewSafeRingQueue[int](MAX, WhenFullError, WhenEmptyBlock, nil)
	// Push() to full
	for i := range MAX {
		obj.Push(i)
	}
	assertSize(obj, MAX, t)

	// Pop() all
	for range MAX {
		obj.Pop()
	}
	assertSize(obj, 0, t)

	// Now Pop on empty should block until something available to Pop,
	// there will be none more but we use a timeout.
	const WAITED_VALUE int = 100

	var wg sync.WaitGroup
	wg.Add(2)

	// we will push a value on the empty queue after 5 seconds
	go func() {
		defer wg.Done()

		fmt.Println("Pusher Go-routine")
		fmt.Println("\t...sleeping 5 secs. prior to Push on empty")
		time.Sleep(5 * time.Second)
		obj.Push(WAITED_VALUE)
		fmt.Println("\tPushed a value onto empty queue")
	}()

	// this will block for sometime because the queue is EMPTY
	go func() {
		defer wg.Done()

		fmt.Println("Popper Go-routine")
		fmt.Println("\t(pop) will wait for new data...")
		v, _, err := obj.Pop()
		fmt.Printf("... (pop) unblocked Pop with new data: %d\n", v)
		if v != WAITED_VALUE {
			t.Errorf("... (pop) waited but got a different value: %d", v)
		}
		if err != nil {
			t.Errorf("... (pop) unblocked Pop but with an error: %v", err)
		}
	}()

	wg.Wait()
}
