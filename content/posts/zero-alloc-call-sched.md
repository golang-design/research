---
date: 2021-01-26T13:11:00+01:00
toc: true
slug: /zero-alloc-call-sched
tags:
  - Channel
  - EscapeAnalysis
  - GUI
  - MainThread
  - Thread
  - Tracing
  - MemAlloc
title: Scheduling Function Calls with Zero Allocation
---

Author(s): [Changkun Ou](https://changkun.de)

GUI programming in Go is a little bit tricky. The infamous issue
regarding interacting with the legacy GUI frameworks is that most of
the graphics related APIs must be called from the main thread.
This basically violates the concurrent nature of Go: A goroutine may be
arbitrarily and randomly scheduled or rescheduled on different running
threads, i.e., the same pice of code will be called from different
threads over time, even without evolving the `go` keyword.

<!--more-->

## Background

In multi-threaded programming, operating systems provide space,
the so-called Thread Local Storage (TLS) for each thread of a process
to store their private and local content. In the era where multithreaded
programming and scheduling algorithms are not rich enough,
the TLS feature was very useful to avoid data race since this storage is
purely local and guaranteed by the operating system.

For example, a graphics rendering backend such as OpenGL Context was
designed to store the rendering context of each thread on TLS;
In macOS, the famous GUI framework Cocoa also requires rendering user
interfaces on a specific thread, that is the so-called *main thread*.

## The Main Thread

In Go, as we all know that a goroutine will be scheduled on different
threads due to its internal work-stealing scheduler [^work-steal] [^go11sched].

With work-tealing scheduler, goroutines are not promised to run on a specific
thread forever. Instead, whenever a goroutine goes to sleep, or endering a
system call, or the Go's runtime proactively interrupts the execution of
that goroutine, it is likely to be rescheduled to a different thread.
Therefore, if a rendering (OpenGL) context is stored on the old thread,
after switching to a new thread will cause the lose of that old context, too.
**Because such an interruption can happen at anytime and anywhere,
it is impossible to check if the goroutine remains on the same thread
when the execution resumes.**

The original intntion of designing such a scheduler is to eliminate
the concept of system thread and multiplexing it. In this way, users won't
suffer from the paying the cost of threads switch/sleep whereas threads
always in its full power mode that constantly running tasks either from
user or the runtime.

### Method `runtime.LockOSThread` and Package `mainthread`

If GUI applications must interact with the OS on the main thread,
how can we achieve the goal where we want run a specific thread perminantly?
Luckily, there is a method called `LockOSThread` offered from the
`runtime` package, provides the exact feature we want:

```go
// LockOSThread wires the calling goroutine to its current operating system thread.
// The calling goroutine will always execute in that thread,
// and no other goroutine will execute in it,
// until the calling goroutine has made as many calls to
// UnlockOSThread as to LockOSThread.
// If the calling goroutine exits without unlocking the thread,
// the thread will be terminated.
//
// All init functions are run on the startup thread. Calling LockOSThread
// from an init function will cause the main function to be invoked on
// that thread.
//
// A goroutine should call LockOSThread before calling OS services or
// non-Go library functions that depend on per-thread state.
func LockOSThread()
```

As the document of `LockOSThread` says: All `init` functions are run on
the startup thread. Calling `LockOSThread` from an `init` function will
cause the main function to be invoked on that thread.

If you think about that carefully, you will immediately realize this gives
us the opportunity to identify, at least, the main thread.
When we would like to wrapping thread scheduling as a package `mainthread`,
we can do something like the following:

```go {linenos=inline,hl_lines=[13,16],linenostart=1}
package mainthread // import "x/mainthread"

import "runtime"

func init() {
	runtime.LockOSThread()
}

// Init initializes the functionality of running arbitrary subsequent
// functions be called on the main system thread.
//
// Init must be called in the main.main function.
func Init(main func())

// Call calls f on the main thread and blocks until f finishes.
func Call(f func())
```

As a user of such a package, one can:

```go {linenos=inline,hl_lines=[15],linenostart=1}
package main

func main() {
	mainthread.Init(fn)
}

func fn() {
	// ... do what ever we want to do in main ...
}


func gn() {
	// Wherever gn is running, the call will be executed on the main thread.
	mainthread.Call(func() {
		// ... do whatever we want to run on the main thread ...
	})
}
```

Once we solved the problem of API design, the next question is:
How can we implement the `Init` and `Call`?

Well, it is not that difficult. Recall that we use `Init` method
to obtain the full control of the main thread, then we should never and
ever to give up such a power. Thus, creating another goroutine to run
what we initially wants to run, and use a channel to receive
the calls that we would like to schedule on the main thread
becomes our only option:

```go
// funcQ is a global channel that responsible for receiving function
// calls that needs to run on the main thread.
var funcQ = make(chan func(), runtime.GOMAXPROCS(0))

func Init(main func()) {
	done := make(chan struct{})
	go func() {
		main()

		// main function terminates, signal and terminate
		// the main thread too.
		done <- struct{}{}
	}()

	for {
		select {
		case f := <-funcQ:
				f()
		case <-done:
			return
		}
	}
}
```

Since we have the global `funcQ`, scheduling a function via that channel
becomes an easy work:

```go
// Call calls f on the main thread and blocks until f finishes.
func Call(f func()) {
	done := make(chan struct{})
	funcQ <- func() {
		f()
		done <- struct{}{}
	}
	<-done
}
```

> Note that we use empty struct as our channel signal, if you are not
> familiar with empty struct and channels, you might want read two great
> post from Dave Cheney [^empty-struct] [^curious-channels].

To use such a package, one can use `mainthread.Call` to schedule
a call to be executed on the main thread:

```go
package main

import "x/mainthread"

func main() {
	mainthread.Init(fn)
}

func fn() {
	done := make(chan struct{})
	go gn(done)
	<-done
}

func gn(done chan<- struct{}) {
	mainthread.Call(func() {
		println("call on the main thread.")
	})
	done <- struct{}{}
}
```

### Creating A Window with `glfw` using `mainthread`

Whenever we need to wrap a window package, such as initializing `glfw` [^glfw]:

```go
package app // import "x/app"

import (
	"x/mainthread"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// Init initializes an app environment.
func Init() (err error) {
	mainthread.Call(func() { err = glfw.Init() })
	return
}

// Terminate terminates the entire application.
func Terminate() {
	mainthread.Call(glfw.Terminate)
}
```

and make sure critical calls like `glfw.WaitEventsTimeout` inside
the rendering loop always be executed from the main thread:

```go
package app // import "x/app"

// Win is a window.
type Win struct {
	win *glfw.Window
}

// NewWindow constructs a new graphical window.
func NewWindow() (*Win, error) {
	var (
		w   = &Win{}
		err error
	)
	mainthread.Call(func() {
		w.win, err = glfw.CreateWindow(640, 480, "golang.design/research", nil, nil)
		if err != nil {
			return
		}
	})
	if err != nil {
		return nil, err
	}
	w.win.MakeContextCurrent()
	return w, nil
}

// Run runs the given window and blocks until it is destroied.
func (w *Win) Run() {
	for !w.win.ShouldClose() {
		w.win.SwapBuffers()
		mainthread.Call(func() {
			// This function must be called from the main thread.
			glfw.WaitEventsTimeout(1.0 / 30)
		})
	}
	// This function must be called from the mainthread.
	mainthread.Call(w.win.Destroy)
}
```

As a user of `app` package, can get rid of the understanding
and thought overhead about when and how do we call a function
on the main thread:

```go
package main

import (
	"x/app"
	"x/mainthread"
)

func main() {
	mainthread.Init(fn)
}

func fn() {
	err := app.Init()
	if err != nil {
		panic(err)
	}
	defer app.Terminate()
	w, err := app.NewWindow()
	if err != nil {
		panic(err)
	}
	w.Run()
}
```

![](../assets/zero-alloc-call-sched/app.png)

Now, we have an empty solid window and will never be crashed randomly 😄.

## Cost Analysis and Optimization

After implementing a first iteration of the `mainthread` package,
we might directly wonder about the performance of this package,
questions could be:

_If a function is sent from a thread to the main thread, what's the
latency when calling such a function?_

Let's write a few benchmark tests that can measure the performance of
such a call. The idea is very simple, we need a baseline to identify
the initial cost of calling a function, then measure the completion
time when we schedule the same function call on the main thread:

```go
var f = func() {}

// Baseline: call f() directly.
func BenchmarkDirectCall(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

// MainthreadCall: call f() on the mainthread.
func BenchmarkMainThreadCall(b *testing.B) {
	mainthread.Init(func() {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mainthread.Call(f)
		}
	})
}
```

Be careful with micro benchmarks here, as we discussed in our previous
research [^bench-time], let's use the `bench`
tool [^bench-tool] for benchmarking. `bench` is tool for executing
Go benchmarks reliably. It automatically locks machine's performance
and execute benchmarks 10x by default to eliminate system measurement error:

```
$ bench
goos: darwin
goarch: arm64
pkg: x/mainthread-naive
...

name              time/op
DirectCall-8      0.95ns ±1%
MainThreadCall-8   448ns ±0%

name              alloc/op
DirectCall-8       0.00B
MainThreadCall-8    120B ±0%

name              allocs/op
DirectCall-8        0.00
MainThreadCall-8    2.00 ±0%
```

The benchmark result indicates that calling an empty function directly in Go
will `1ns` whereas schedule the same empty function to the main thread
will spend `448ns`. Thus the cost is `447ns`.

Moreover, when we talk about cost,
we actually care about the cost of CPU as well as memory consumption.
According to the second report regarding `allocs/op`, the result shows
scheduling an empty function to the mainthread will cost `120B` allocation.

Allocation of `120B` per operation might not be a big deal from our first impression.
However, if we consider the actual use case of this package, i.e. managing GUI rendering 
calls, either CPU or memory allocation can be propagated to a huge cost over time.
If we are dealing with rendering, especially graphical rendering, the fresh rate
is typically minimum 25fps, ideally 30fps or even higher.

This means, every 5 minutes, without considering mouse button, movements,
and keystrokes, a GUI application will allocate at least:

$$
5 \times 60\times 30 \times 120 \text{byte} = 1.08 \text{MiB}
$$

A directly impact from an excessive allocation behavior is the runtime garbage
collector and the scavenger. With higher allocation rate, the garbage collector
is triggered more often, and the scavenger releases memory to the OS more often.
Because of more works are produced for the GC, the GC will also consume more
CPU from the system. It is good enough to say the entire application is
a vicious circle.

The following is a trace information of that above application runs in 6 minutes, the total heap allocation is actually 1.41 MiB (2113536-630784 byte), preety close to what we predicted before.

![](./../assets/zero-alloc-call-sched/naive-sched-trace-1.png)
![](./../assets/zero-alloc-call-sched/naive-sched-trace-2.png)

Where does the allocation come from?
How can we deal with these issues?
How to optimize the exisiting naive implementation?
Let's find out in the next section.

## Optimal Threading Control

The first optimization comes to the attempt of avoid allocating
channels. In our `Call` implementation, we allocate a signal channel
for every function that we need to call from the main thread:

```go {linenos=inline,hl_lines=[3],linenostart=1}
// Call calls f on the main thread and blocks until f finishes.
func Call(f func()) {
	done := make(chan struct{}) // allocation!
	funcQ <- func() {
		f()
		done <- struct{}{}
	}
	<-done
}
```

This means everytime when we call the `Call` method, we will have to
allocate at least 96 bytes for a channel due to the Go compiler will
uses `runtime.hchan` as the struct that represents the actual channel:

```go
// in src/runtime/chan.go

// the hchan struct needs 96 bytes.
type hchan struct {
	qcount   uint
	dataqsiz uint
	buf      unsafe.Pointer
	elemsize uint16
	closed   uint32
	elemtype *_type
	sendx    uint
	recvx    uint
	recvq    waitq
	sendq    waitq
	lock     mutex
}
```

A well known trick to avoid repetitive allocation is to use
the `sync.Pool`. One can:

```go {linenos=inline,hl_lines=["1-3", 6, 7],linenostart=1}
var donePool = sync.Pool{New: func() interface{} {
	return make(chan struct{})
}}

func Call(f func()) {
	done := donePool.Get().(chan struct{}) // reuse signal channel via sync.Pool!
	defer donePool.Put(done)

	funcQ <- func() {
		f()
		done <- struct{}{}
	}
	<-done
}
```

With that simple optimization, a rebenchmarked result indicates
an 80% reduction of memory usage:

```txt {linenos=inline,hl_lines=[3,7,11],linenostart=1}
name              old time/op    new time/op      delta
DirectCall-8      0.95ns ±1%         0.95ns ±1%    ~     (p=0.631 n=10+10)
MainThreadCall-8   448ns ±0%         440ns ±0%   -1.83%  (p=0.000 n=9+9)

name              old alloc/op   new alloc/op     delta
DirectCall-8       0.00B            0.00B          ~     (all equal)
MainThreadCall-8    120B ±0%          24B ±0%   -80.00%  (p=0.000 n=10+10)

name              old allocs/op  new allocs/op    delta
DirectCall-8        0.00             0.00          ~     (all equal)
MainThreadCall-8    2.00 ±0%         1.00 ±0%   -50.00%  (p=0.000 n=10+10)
```

Can we do it even better? The answer is yes. As you can notice that
there is still a 24B of allocation per operation. But to identify it becomes
a little bit tricky.

In Go, variables can be allocated from heap if:

1. Using `make` and `new` keywords explicitly, or 
2. Escape from the stack

The second case is a little bit advance from the normal use of Go. To be short,
escape from the execution stack to the heap is something that decided from
compile time. The Go's compiler will decide when should a vaiable be allocated
on the heap. The process of deciding allocate variables either on the stack or
the heap is called _escape analysis_.

The great thing about the Go is that this information is trackable and can be
enabled directly from the Go toolchain. One can use `-gcflags="-m"` to activate
the escape analysis and see the result from the compile time:

```shell
$ go build -gcflags="-m"
./mainthread.go:52:11: can inline Call.func1
./mainthread.go:48:11: leaking param: f
./mainthread.go:52:11: func literal escapes to heap
```

The compiler shows us that the sending function is leaking and the wrapper
function that sends via our `funcQ` is causing the func literal escaping to the
heap. The reason that func literal escapes to the heap is because a func literal
is considered as a pointer, and sending a pointer via channel will always
cause an escape by design.

To avoid the escaping fucntion literal, instead of using a function wrapper,
we can send a struct:

```go {linenos=inline,hl_lines=["1-4", 10],linenostart=1}
type funcdata struct {
	fn   func()
	done chan struct{}
}

func Call(f func()) {
	done := donePool.Get().(chan struct{})
	defer donePool.Put(done)

	funcQ <- funcdata{fn: f, done: done} // wrap the information
	<-done
}
```

and when we receives the `funcdata`:

```go {linenos=inline,hl_lines=["6-8"],linenostart=1}
func Init(main func()) {
	...

	for {
		select {
		case fdata := <-funcQ:
			fdata.fn()
			fdata.done <- struct{}{}
		case <-done:
			return
		}
	}
}
```

After such an optimization, a re-benchmarked result indicates that
we hint the zero-allocation goal:

```txt {linenos=table,hl_lines=[3,7,11],linenostart=1}
name              old time/op     new time/op     delta
DirectCall-8      0.95ns ±1%      0.95ns ±1%        ~      (p=0.896 n=10+10)
MainThreadCall-8   448ns ±0%       366ns ±1%     -18.17%   (p=0.000 n=9+9)

name              old alloc/op    new alloc/op    delta
DirectCall-8       0.00B             0.00B          ~      (all equal)
MainThreadCall-8    120B ±0%            0B      -100.00%   (p=0.000 n=10+10)

name              old allocs/op   new allocs/op   delta
DirectCall-8        0.00             0.00           ~      (all equal)
MainThreadCall-8    2.00 ±0%         0.00       -100.00%   (p=0.000 n=10+10)
```

Hooray! 🎉

## Verification and Discussion

Before we conclude our research, let's do a final verification on the real world
example that we had before: the GUI application.

While a re-evaluation, we can see from the trace file that entire application
is still allocating memory and the heap is still increasing:

![](../assets/zero-alloc-call-sched/opt-sched-trace.png)

Notably, the total allocated bytes during the application life cycle (6 minutes)
only allocates:

$$ 958464 - 622592 = 0.32 \text{MiB} $$

Comparing to the previous 1.41 MiB allocation, we optimized 1.08 MiB of memory
which is exactly what was predicted before.

We might still wondering, if scheduling is not allocating memory anymore,
who is still allocating the memory? To find out, we need a little bit help
from the `runtime` package. The compiler translates the allocation operation
to a runtime function `runtime.newobject`. One can add three more lines
and prints who is exactly calling this function using `runtime.FuncForPC`:

```go {linenos=inline,hl_lines=["3-5"],linenostart=1}
// src/runtime/malloc.go
func newobject(typ *_type) unsafe.Pointer {
	f := FuncForPC(getcallerpc())       // add this
	l, ll := f.FileLine(getcallerpc())  // add this
	println(typ.size, f.Name(), l, ll)  // add this
	return mallocgc(typ.size, typ, true)
}
```

In above, the `getcallerpc` is a runtime private helper.
If we execute the application again, we will see printed information similar
to below:

```
88 runtime.acquireSudog /Users/changkun/dev/godev/go-github/src/runtime/proc.go 375
88 runtime.acquireSudog /Users/changkun/dev/godev/go-github/src/runtime/proc.go 375
88 runtime.acquireSudog /Users/changkun/dev/godev/go-github/src/runtime/proc.go 375
...
```

This demonstrates how and why the allocation still happens:

```go {linenos=inline,hl_lines=[23],linenostart=1}
// ch <- elem
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	...
	gp := getg()
	mysg := acquireSudog()
	...
}

//go:nosplit
func acquireSudog() *sudog {
	mp := acquirem()
	pp := mp.p.ptr()
	if len(pp.sudogcache) == 0 {
		lock(&sched.sudoglock)
		for len(pp.sudogcache) < cap(pp.sudogcache)/2 && sched.sudogcache != nil {
			s := sched.sudogcache
			sched.sudogcache = s.next
			s.next = nil
			pp.sudogcache = append(pp.sudogcache, s)
		}
		unlock(&sched.sudoglock)
		if len(pp.sudogcache) == 0 {
			pp.sudogcache = append(pp.sudogcache, new(sudog)) // !here
		}
	}
	...
}
```

Unfortunately, this is completely outside the control of the userland.
We are not able to optimize here anymore. 
Nevertheless, we have reached our goal for today, and this is the best
of what we can do so far.

One more thing, if you take a closer look into how much the heap grows
for one step, you will get some calculation like this: 671744-663552=8192
This is actually the minimum allocation size of the runtime allocator,
which allocates a _page_. Due to the discussion of this
topic if much more out from our goal in this research, we leave that as
your future outlook to dig more on your own, there is a great blog post
[^mem-alloc] as your starting point.

## Conclusion

In this research, we covered the following topics:

1. Go's runtime scheduler
2. Scheduling on a specific thread, especially the main thread
3. Reliable benchmarking and allocations tracing techniques
4. Go's runtime memory allocator
5. Go's runtime garbage collector
6. Escape analysis
7. Go's channel implementation

There are several points we can summarize:

1. A channel allocates 96 bytes of memory
2. A function literal allocate 24 bytes of memory
3. Escape analysis can help us identify unexpected allocations, and function literal is considered as a pointer that always escapes to the heap
4. Sending information via a channel can cause allocation intrinsically from the runtime.
5. Go runtime grows the the heap 8K on each step as page allocation

Beyond this research, we also encapsulated all the abstractions derived from this
research, and we published two packages: `mainthread`[^mainthread] and `thread`[^thread].
These packages gives you the ability to schedule any function calls
either on the main thread, or a specific thread.

Have fun!

## References

[^work-steal]: Robert D. Blumofe and Charles E. Leiserson. 1999. "Scheduling multithreaded computations by work stealing." J. ACM 46, 5 (September 1999), 720-748. https://dl.acm.org/citation.cfm?id=324234
[^go11sched]: Dmitry Vyukov. "Scalable Go Scheduler Design Doc." May 2, 2012. https://golang.org/s/go11sched
[^glfw]: The glfw Library. https://www.glfw.org/
[^bench-time]: Changkun Ou. "Eliminating A Source of Measurement Errors in Benchmarks
." 30.09.2020. https://golang.design/research/bench-time/
[^bench-tool]: Changkun Ou. "bench: Reliable performance measurement for Go programs. All in one design." https://golang.design/s/bench
[^empty-struct]: Dave Cheney. "The empty struct." March 25, 2014. https://dave.cheney.net/2014/03/25/the-empty-struct
[^curious-channels]: Dave Cheney. "Curious Channels." April 30, 2013. https://dave.cheney.net/2013/04/30/curious-channels
[^mem-alloc]: Dave Cheney. "A few bytes here, a few there, pretty soon you’re talking real memory." Jan 05, 2021. https://dave.cheney.net/2021/01/05/a-few-bytes-here-a-few-there-pretty-soon-youre-talking-real-memory
[^mainthread]: Changkun Ou. "Package golang.design/x/mainthread." https://golang.design/s/mainthread
[^thread]: Changkun Ou. "Package golang.design/x/thread." https://golang.design/s/thread