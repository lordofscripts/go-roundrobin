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
