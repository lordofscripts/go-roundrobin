/* -----------------------------------------------------------------
 *				   P u b l i c   D o m a i n / F O S
 *  			Copyright (C)2023 Serge Toro, and
 *				Copyright (C)2025 Muhammad H. Hosseinpour
 * - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
 * Benchmarking the RingQueue.
 *-----------------------------------------------------------------*/
package roundrobin

/*
	go test -bench=. -benchmem
	DEGT: Added b.Loop() RRR tests
	https://betterstack.com/community/guides/scaling-go/golang-benchmarking/
*/

import (
	"testing"
)

/* ----------------------------------------------------------------
 *					B e n c h m a r k s
 *-----------------------------------------------------------------*/

/**
 * Benchmarking a generic RingQueue[rune]
 */
func BenchmarkRingQueueR(b *testing.B) {
	rr := NewRingQueue[rune](1_000)

	for b.Loop() {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push('ẞ')
	}
}

/**
 * Benchmarking (out of curiosity) a plain RingQueue
 * specialized (rather than generic) in runes. We fill
 * it with the same multi-byte rune.
 */
func BenchmarkRuneRingQueue(b *testing.B) {
	rr := NewRuneRingQueue(1_000)

	for b.Loop() {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push('ẞ')
	}
}

/**
 * Benchmarking plain GENERIC RingQueue[int]
 */
func BenchmarkRingQueue(b *testing.B) {
	rr := NewRingQueue[int](1_000)

	for n := 0; b.Loop(); n++ {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push(n)
	}
}

/**
 * Benchmarking Array-based circular buffer
 */
func BenchmarkArray(b *testing.B) {
	var ar [1_000]int
	size := 0

	for n := 0; b.Loop(); n++ {
		if size >= len(ar) {
			copy(ar[0:], ar[1:])
			size--
		}

		ar[size] = n
		size++
	}
}
