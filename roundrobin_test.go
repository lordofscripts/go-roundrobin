/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2023 Serge Toro, and
 *				Copyright (C)2025 Muhammad H. Hosseinpour,
 *				Copyright (C)2025 Dídimo Grimaldo T.
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * Tests for methods on plain (Serge's) RingQueue
 *-----------------------------------------------------------------*/
package roundrobin

/*
	go test -bench=. -benchmem
	DEGT: Added b.Loop() RRR tests
	https://betterstack.com/community/guides/scaling-go/golang-benchmarking/
*/

import (
	"fmt"
	"testing"
)

/* ----------------------------------------------------------------
 *						T e s t s
 *-----------------------------------------------------------------*/

func Test_Stringer(t *testing.T) {
	obj := NewRingQueue[int](10)
	expected := "[RRQ full:false size:10 start:0 end:0 data:[0 0 0 0 0 0 0 0 0 0]]"
	actual := fmt.Sprint(obj)

	if actual != expected {
		t.Fatalf("Mismatch, expected:%s, found:%s", expected, actual)
	}
}

func Test_PushEnough(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 10; idx++ {
		size, err := obj.Push(idx)
		if err != nil {
			t.Fatalf("unexpected error in adding an element with index %d", idx)
		} else if size != idx+1 {
			t.Fatalf("push returned wrong size. got %d exp %d", size, idx+1)
		}
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func Test_PushOver(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 10; idx++ {
		size, err := obj.Push(idx)
		if err != nil {
			t.Fatalf("Unexpected error in adding an element with index %d", idx)
		} else if size != idx+1 {
			t.Fatalf("push returned wrong size. got %d exp %d", size, idx+1)
		}
	}

	_, err := obj.Push(100)
	if err == nil {
		t.Fatalf("Expected overflow error")
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}
}

func Test_PushPop(t *testing.T) {
	obj := NewRingQueue[int](10)
	for idx := 0; idx < 8; idx++ {
		obj.Push(idx)
	}

	expSize := obj.Size()
	for idx := 0; idx < 5; idx++ {
		e, size, err := obj.Pop()
		if err != nil || e != idx {
			t.Fatalf("inconsistent behavior")
		} else if size != expSize-idx-1 {
			t.Fatalf("pop returned wrong size. got %d exp %d", size, expSize-idx-1)
		}
	}

	for idx := 0; idx < 7; idx++ {
		obj.Push(100 + idx)
	}

	expected := []int{102, 103, 104, 105, 106, 5, 6, 7, 100, 101}

	if !eqSlices(obj.data, expected) {
		t.Fatalf("Container data mismatch, expected:%v, found:%v", expected, obj.data)
	}

	if obj.Size() != 10 {
		t.Fatalf("inconsistent size: %d", obj.Size())
	}

	expSize = obj.Size()
	for idx := 0; idx < 10; idx++ {
		e, size, _ := obj.Pop()
		if e != expected[(5+idx)%10] {
			t.Fatalf("inconsistent behavior")
		} else if size != expSize-idx-1 {
			t.Fatalf("pop #%d returned wrong size. got %d exp %d", idx+1, size, expSize-idx-1)
		}
	}
}

/* ----------------------------------------------------------------
 *					F u n c t i o n s
 *-----------------------------------------------------------------*/

func eqSlices[T comparable](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for idx := 0; idx < len(a); idx++ {
		if a[idx] != b[idx] {
			return false
		}
	}

	return true
}
