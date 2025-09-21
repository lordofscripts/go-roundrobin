/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2023 Serge Toro, and
 *				Copyright (C)2025 Dídimo Grimaldo T.
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * Timing tests (expanded and modernized)
 *-----------------------------------------------------------------*/
package roundrobin

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

const (
	mbRUNE rune = 'ß'
)

/* ----------------------------------------------------------------
 *						T e s t s
 *-----------------------------------------------------------------*/

func TestSizes_Array(t *testing.T) {
	fmt.Println("\tArray")

	allStressed := []struct {
		name     string
		capacity int
	}{
		{"Array C=1", 1},
		//{"Array C=10", 10},
		{"Array C=100", 100},
		{"Array C=1K", 1000},
		//{"Array C=10K", 10_000},
		{"Array C=100K", 100_000},
	}

	for _, tt := range allStressed {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sim(tt.capacity)
		})
	}
}

func TestSizes_GenericInt(t *testing.T) {
	fmt.Println("\tRoundRobin[int]")

	allStressed := []struct {
		name     string
		capacity int
	}{
		{"RoundRobin[T] C=1", 1},
		//{"RoundRobin[T] C=10", 10},
		{"RoundRobin[T] C=100", 100},
		{"RoundRobin[T] C=1K", 1000},
		//{"RoundRobin[T] C=10K", 10_000},
		{"RoundRobin[T] C=100K", 100_000},
	}

	for _, tt := range allStressed {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			simRingQueueInt(tt.capacity)
		})
	}
}

func TestSizes_GenericRune(t *testing.T) {
	fmt.Println("\tRoundRobin[rune]")

	allStressed := []struct {
		name     string
		capacity int
	}{
		{"RoundRobin[T] C=1", 1},
		//{"RoundRobin[T] C=10", 10},
		{"RoundRobin[T] C=100", 100},
		{"RoundRobin[T] C=1K", 1000},
		//{"RoundRobin[T] C=10K", 10_000},
		{"RoundRobin[T] C=100K", 100_000},
	}

	for _, tt := range allStressed {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			simRingQueueRune(tt.capacity)
		})
	}
}

func TestSizes_RuneConcrete(t *testing.T) {
	fmt.Println("\tRuneRoundRobin")

	allStressed := []struct {
		name     string
		capacity int
	}{
		{"RuneRoundRobin C=1", 1},
		//{"RuneRoundRobin C=10", 10},
		{"RuneRoundRobin C=100", 100},
		{"RuneRoundRobin C=1K", 1000},
		//{"RuneRoundRobin C=10K", 10_000},
		{"RuneRoundRobin C=100K", 100_000},
	}

	for _, tt := range allStressed {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			simRuneRingQueue(tt.capacity)
		})
	}
}

func TestPrimitiveAsymptoticPerformance(t *testing.T) {
	fmt.Println("Standard array")
	for idx := 7; idx < 14; idx++ {
		sim(1 << idx)
	}

	fmt.Println("RoundRobin (ring) queue")
	for idx := 7; idx < 14; idx++ {
		simRingQueueInt(1 << idx)
	}
}

/* ----------------------------------------------------------------
 *					F u n c t i o n s
 *-----------------------------------------------------------------*/

func Format(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

/**
 * Timing of Array version
 */
func sim(capacity int) {
	ar := make([]int, capacity)
	size := 0

	start := time.Now()
	for n := range 1_000_000 {
		if size >= len(ar) {
			copy(ar[0:], ar[1:])
			size--
		}

		ar[size] = n
		size++
	}

	fmt.Printf("%12s took %v\n", Format(int64(capacity)), time.Since(start).Seconds())
}

/**
 * Timing of plain RingQueue[T] generics version
 */
func simRingQueueInt(capacity int) {
	rr := NewRingQueue[int](capacity)

	start := time.Now()
	for n := range 1_000_000 {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push(n)
	}

	fmt.Printf("%12s took %v\n", Format(int64(capacity)), time.Since(start).Seconds())
}

func simRingQueueRune(capacity int) {
	rr := NewRingQueue[rune](capacity)

	start := time.Now()
	for range 1_000_000 {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push(mbRUNE)
	}

	fmt.Printf("%12s took %v\n", Format(int64(capacity)), time.Since(start).Seconds())
}

/**
 * Timing of concrete RuneRingQueue rune version
 */
func simRuneRingQueue(capacity int) {
	rr := NewRuneRingQueue(capacity)

	start := time.Now()
	for range 1_000_000 {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push(mbRUNE)
	}

	fmt.Printf("%12s took %v\n", Format(int64(capacity)), time.Since(start).Seconds())
}
