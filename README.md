[![Go Reference](https://pkg.go.dev/badge/github.com/lordofscripts/go-roundrobin.svg)](https://pkg.go.dev/github.com/lordofscripts/go-roundrobin)
[![GitHub release (with filter)](https://img.shields.io/github/v/release/lordofscripts/go-roundrobin)](https://github.com/lordofscripts/go-roundrobin/releases/latest)
[![GitHub License](https://img.shields.io/github/license/lordofscripts/go-roundrobin)](https://github.com/lordofscripts/go-roundrobin/blob/master/LICENSE)


# go-roundrobin
Ring (Circular or Round-Robin) queue-like container with fixed memory footprint.

This brings great memories of decades ago when I worked in a high-tech European
company writing telecommunications firmware. Back then C was a luxury and we
had to mix C with Assembly to meet memory space (RAM & ROM) and timing requirements.
We used a Round-Robin buffer in some part. Later, for fun, I spiced up the
round-robin code with some features **similar** to what we see in this module.

**NOTE:** *Despite what it says above that it is N commits behind Serge's repo,
I have already fixed that here but in a different way. I will have to disconnect
the reference because our repositories are quite different.*

## What a Ring Queue is and Why use it
So, what is a Ring Queue and why it's useful:
1. It uses a static / fixed array buffer (no costly allocations)
1. It does not shift elements on element removal (no costly memcopy)
1. It can start anywhere in the buffer and loop over the end

![img](https://sergetoro.com/images/RQ.svg)

## Using go-roundrobin

To add it on your module or GO application:

`go get https://github.com/lordofscripts/go-roundrobin`

Then in your source code:

```go
	import "lordofscripts/go-roundrobin"
```	

### The programming interface

While it has three different implementations, you decide which one is more
suitable for your requirements. All three implementations comply with:

```go
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
```

If you plan to use it with Go Routines, or need to specify `onClose` actions
and define behaviors for Full and Empty queues, or even set a deadline for
a blocking Pop operation:

```go
	const CAPACITY int = 100
	var intCallback roundrobin.OnCloseCallback[int] = somefunctionOrMethod
	intRingbuffer := roundrobin.NewSafeRingBuffer[int](CAPACITY, WhenFullError, WhenEmptyBlock, intCallback)
	intRingbuffer.SetPopDeadline(time.Now().Add(5 * time.Second))

	charRingbuffer := roundrobin.NewSafeRingBuffer[rune](CAPACITY, WhenFullOverwrite, WhenEmptyError, nil)
```

If you don't need any of the extra features (full/empty behavior, close callback, 
thread-safe, etc.):

```go
	const CAPACITY int = 100
	intRingbuffer := roundrobin.NewRingBuffer[int](CAPACITY)
	charRingbuffer := roundrobin.NewRingBuffer[rune](CAPACITY)
```

And in the remote case you need something suitable for characters/runes whether
single or multi-byte, not for concurrent use but with the ability to have a
user-defined behavior when the buffer is full:

```go
	const CAPACITY int = 100
	charRingbuffer := roundrobin.NewRuneRingBuffer(CAPACITY).
			SetWhenFull(WhenFullOverwrite)  // not required!
```

### Improvements over original code

As you may have noticed, this is a *forked* repository from [Serge](https://github.com/sombr/go-container-roundrobin).
I was looking for something like that without the need to redesign the wheel, just improve it.

In this forked version, the original has been enhanced with a slightly different interface `IRingQueue`. So this
repository is a hybrid between Serge's original, and [Haddi's](https://github.com/hadi77ir/go-ringqueue) 
much improved version that introduced a change in the original signatures and the addition of a "safe" version
that uses a synchronization mechanism.

**Serge's** wonderful [implementation](https://github.com/sombr/go-container-roundrobin) can be
used with the `NewRingQueue[T]()` *constructor*. It has these features:

* Safe for **single-threaded** applications
* Based on **generic types**
* Benchmarked the differences between a plain array-based ring queue and his generics ring queue.
* Read his full post here: https://www.sergetoro.com/golang-round-robin-queue-from-scratch

**Muhammad's** [repository](https://github.com/hadi77ir/go-ringqueue) can be 
instantiated with the `NewSafeRingQueue[T]()` constructor. It brings in great
new features:

* Has a synchronization mechanism making it suitable for go-routines (multi-threaded).
* The user can specify the behavior when `Push()` onto a **full** queue: `WhenFullError` or
  `WhenFullOverwrite`.
* The user can specify the behavior when `Pop()` from an **empty** queue: `WhenEmptyError` or
  `WhenEmptyBlock` (block until data available).
* The user can specify a callback to invoke when the Ring Queue is closed. Once closed
  all operations on it will fail. He added this to both his thread-safe version and the
  original thread-unsafe version. Additionally he backported the `WhenFull` feature
  into the original (thread-unsafe) version.
* He introduced a *generic interface* implemented by both the single and 
  multi-threaded versions.
* Given that a `Pop()` on an empty queue *may* block (user-defined), he
  introduced a user-defined *deadline*.
* Consolidated errors that are globally-defined  
* Has a bug in its `Len()` method though! which Serge already fixed after I filed an issue.
  Unfortunately, M. Hadi's repository has disabled issues.
* I believe he also has a bug in his Push method that instead checks `whenEmpty`
  rather than `whenFull`, but again, there is no way to report problems with his code.  

I needed a ring queue for runes only, and loved the benchmarking. I was curious about
the performance differences between Serge's generics version, and a concrete rune
version based entirely on his type, but without the generics. Then I discovered
Muhammand's great enhancements and decided it would be nice to bring all together but
adapted to my needs.

So, I, **Lord of Scripts** created this hybrid fork. My rune-specific type can
be instantiated with the `NewRuneRingQueue()` constructor. While I didn't add great
functionality other than adding a new (basic) type for curiosity, I also did this:

* Added tests for different scenarios of `WhenFull` and `WhenEmpty`
* Fixed several size issues from both Haddi's & Serge's code that returned the
  wrong size. I decided to use a thread-safe size counter instead of the unreliable
  size based on start/end comparisons (that bug remains in both their repos.)
* I renamed the *interface* to `IRingQueue[T any]` and all three objects implement
  this interface.
* `RuneRingQueue` is only suitable for single-threaded applications, like Serge's.
* I simplified Muhammed's constructors without compromising functionality (see below).
* I added the `SetWhenFull(WhenFull)` fluent API interface method.
* I added the `SetOnClose(func OnCloseCallback[T])` fluent API interface method.
* Added `Reset()` which is handy when reusing a buffer of the same size
* All sources use a standards template, sorry, it comes from my times working at
  Software Process Improvement and Standarization.
* Refactored the tests so that they are split into several files depending on
  who implemented it.
* Parallelized and modernized the Timing tests. Now they use a feature introduced
  in GO v1.24.


## Performance

If you are interested, here are some benchmarks we ran on both the original code
and the improved code.

## RingQueue Performance
Do you like benchmarks? I love them, even though many cases they are non-exhaustive, relatively synthetic and might give a skewed view of reality :)
So, let's see how our implementation performs against a simple Go array.
Remember, what we're looking for is an array slow down due to excessive copying.

```go
func BenchmarkRR(b *testing.B) {
	rr := NewRingQueue[int](100_000)

	for n := 0; n < b.N; n++ {
		if rr.IsFull() {
			rr.Pop()
		}
		rr.Push(n)
	}
}

func BenchmarkArray(b *testing.B) {
	var ar [100_000]int
	size := 0

	for n := 0; n < b.N; n++ {
		if size >= 100_000 {
			copy(ar[0:], ar[1:])
			size--
		}

		ar[size] = n
		size++
	}
}
```

Which one do you thing would be the fastest? 😉 Let's run it!

```bash
go test -bench=. -benchmem
```

Here's results on my Intel NUC Ubuntu Linux machine:

```bash
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i7-10710U CPU @ 1.10GHz
BenchmarkRR-12          59798865                18.50 ns/op
BenchmarkArray-12        1000000             17321.00 ns/op
PASS
```

If I run the same test with just 10_000 queue size instead of 100_000:

```bash
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i7-10710U CPU @ 1.10GHz
BenchmarkRR-12          57184480                18.65 ns/op
BenchmarkArray-12        1000000              1068 ns/op
PASS
```

Let's plot the relation of run times and queue lengths:
![img](https://sergetoro.com/images/RQbench.svg)

## On Lord of Script's old (2010) laptop

And old laptop from 2010 which works quite fine with Linux but it
crawls to a halt with Windows 10 after they dumped zillions of
bloatware and delayware to force people to buy new computers. 

My system is an Intel i5-M430 with 4 cores at 2.27GHz instead of Serge's
Intel i7 with 12 cores at 1.10GHz.

The **benchmark** on the specialized `RuneRingQueue` which spares
you from the overhead of generics:

```bash
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5 CPU       M 430  @ 2.27GHz
BenchmarkRR-4          22,946,761                55.32 ns/op
BenchmarkArray-4        1,000,000              1,068 ns/op
PASS
```

Whereas the **timing** tests on a 100K Ring buffer using a plain
array, the concrete Rune implementation (no generics) and the
generics implementation respectively:

```bash
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5 CPU       M 430  @ 2.27GHz
TestSizes_Array        100K elements       116.96s
TestSizes_RuneConcrete 100K elements       0.055031723s
TestSizes_GenericRune  100K elements       0.060474519s
TestSizes_GenericInt   100K elements       0.055532279s
PASS
```