---
date: 2021-08-09T09:02:42+02:00
toc: true
slug: /ultimate-channel
tags:
  - Go
  - Synchronization
  - Deadlock
title: The Ultimate Channel Abstraction
---

Author(s): [Changkun Ou](https://changkun.de)

Permalink: https://golang.design/research/ultimate-channel

Recently, I have been rethinking the programming patterns regarding computer graphics applications, and have been written a 3D graphics package in Go, called [polyred](https://poly.red).
While I was designing the rendering execution pipeline, a tricky deadlock struggled with me for a long time and led to creating an unbounded channel as a workaround solution eventually.

<!--more-->

## The problem

At the beginning of my design, I had to deal with [OpenGL](https://github.com/go-gl/gl) where a chunk of APIs must be executed on the mainthread and issue a draw call is one of those
infamous. The common pattern in graphics programming is as follows:

```go
app := newApp()
driver := initDriver()
ctx := driver.Context()

for !app.IsClosed() {
	ctx.Clear()
	processingDrawCalls(ctx)
	processingInputEvents()
}
```

The entire GUI application is executed in an infinite loop that contains two parts: draw call processing and event processing.

Typically, all these codes run on the CPU, and the actual rendering computation executes on a GPU. That means the graphics API provided by a graphic driver (such as OpenGL, Vulkan, Metal, Direct X) is just a communication command send from the CPU to the GPU or even waiting for a response from the GPU.
For some special reasons, the [polyred](https://poly.red) is limited to software implementation, a pure-CPU implementation. Hence, the execution should utilize the full power of CPU parallelization. It makes much more sense to execute rendering on a separate goroutine so that it won't block the event processing thread.

*_In fact, to guarantee an application's responsiveness, it is ideal not to block the event processing since there might also be system invocation._

Nevertheless, I turned the rendering loop into a separate goroutine and sent the rendering result to the event processing loop to be flushed to the hardware display. The entire application works as the following code snippet:

```go
// WARNING: This example contains a deadlock.
package main

import (
	"fmt"
	"math/rand"
	"time"
)

type ResizeEvent struct {
	width, height int
}

type renderProfile struct {
	id     int
	width  int
	height int
}

// Draw executes a draw call by the given render profile
func (p *renderProfile) Draw() interface{} {
	return fmt.Sprintf("draw-%d-%dx%d", p.id, p.width, p.height)
}

func main() {
	// draw is a channel for receiving finished draw calls.
	draw := make(chan interface{})
	// change is a channel to receive notification of the change of rendering settings.
	change := make(chan ResizeEvent)

	// Rendering Thread
	//
	// Sending draw calls to the event thread in order to draw pictures.
	// The thread sends darw calls to the draw channel, using the same
	// rendering setting id. If there is a change of rendering setting,
	// the event thread notifies the rendering setting change, and here
	// increases the rendering setting id.
	go func() {
		p := &renderProfile{id: 0, width: 800, height: 500}
		for {
			select {
			case size := <-change:
				// Modify rendering profile.
				p.id++
				p.width = size.width
				p.height = size.height
			default:
				draw <- p.Draw()
			}
		}
	}()

	// Event Thread
	//
	// Process events every 100 ms. Otherwise, process drawcall request
	// upon-avaliable.
	event := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case id := <-draw:
			println(id)
		case <-event.C:
			// Notify the rendering thread there is a change regarding
			// rendering settings. We simulate a random size at every
			// event processing loop.
			change <- ResizeEvent{
				width:  int(rand.Float64() * 100),
				height: int(rand.Float64() * 100),
			}
		}
	}
}
```

As one can observe from the above example, it simulates a resize event of a GUI window at every event processing loop. Whenever the size of the GUI window is changed, the underlying rendering should adapt to that, for instance, reallocating the rendering buffers. To let the rendering thread understand the change, another channel is used to communicate from the event thread to the rendering thread.

It sounds like a perfect design, but a nasty deadlock is hidden in the dark if one executes the program, and the program will freeze until a manual interruption:

```
draw-0-800x500
...
draw-0-800x500
draw-1-60x94
...
draw-1-60x94
^Csignal: interrupt
```

If we take a closer look into the program pattern:

1. Two infinite `select` loops (say `E` and `R`) running on different goroutines (threads).
2. The `E` thread receives communication from the `R` thread
3. The `R` thread receives communication from the `E` thread

Found the problem? The problem happens in the two-way communication:
If the communication channels are unbuffered channel (wait until the
receive is complete), the deadlock happens when `E` is waiting for `R` to
complete the receive, and `R` is also waiting for `E` to complete the receive.

One may argue instantly that the deadlock can be resolved using a
buffered channel:

```diff
-draw := make(chan interface{})
+draw := make(chan interface{}, 100)
-change := make(chan ResizeEvent)
+change := make(chan ResizeEvent, 100)
```

But unfortunately, it remains problematic. Let's do a thought experiment: if `E` is too busy, and quickly exploits the entire buffer of the communication channel `change`, then the communication channel falls back to an unbuffered channel. Then `E` starts to wait to proceed; On the otherwise, `R` is busy working on the draw call, when it is finished, `R` tries to send the draw call to `E`.  However, at this moment. the `E` is already waiting for `R` to receive the `change` signal. Hence, we will fall back to the same case -- deadlock.

Is the problem a producer-consumer scenario? Indeed, the case is quite similar but not entirely identical. The producer-consumer scenario focuses on producing content for the buffer while the consumer consumes the buffer. If the buffer is full, it is easy to send either producer or consumer to sleep. However, the key difference here is: On the two sides of communication, they both play the role of producer and consumer simoutainiously, and they both relying on each other.

What can we do to solve the above deadlock? There are two
approaches.

## Solution 1: Send in select's case

The first approach is a simple one. We utilize the power of the select statement, that a send operation to any channel won't block if there is a default statement. Hence, we could simply turn the draw call sends statement into a nested select statement:

```diff
go func() {
	p := &renderProfile{id: 0, width: 800, height: 500}
	for {
		select {
		case size := <-change:
			// Modify rendering profile.
			p.id++
			p.width = size.width
			p.height = size.height
		default:
-			draw <- p.Draw()
+			select {
+			case draw <- p.Draw():
+			default:
+			}
		}
	}
}()
```

In this case, if the `draw <- p.Draw()` is blocking, the newly introduced `select` statement will not block on the send and execute the default statement then resolves the deadlock.

However, there are two drawbacks to this approach: 

1. If a draw call is skipped, there will be one frame loss of rendering. Because the next loop will start to calculate a new frame.
2. The event thread remains blocked until a frame rendering in the rendering thread is complete. Because the new select statement can only be executed after all rendering calculation is complete.

These two drawbacks are there intrinsically, and with this approach, it seems there is no better way to improve it. What else could we do?

## Solution 2: Unbounded Channel

We may first come up with this idea: can we make a channel that contains an infinite-sized buffer, i.e. unbounded channel? Though the language, it is not possible yet. However, such a pattern can be easily constructed:

```go
// MakeChan returns a sender and a receiver of a buffered channel
// with infinite capacity.
func MakeChan() (chan<- interface{}, <-chan interface{}) {
	in, out := make(chan interface{}), make(chan interface{})

	go func() {
		var q []interface{}
		for {
			e, ok := <-in
			if !ok {
				close(out)
				return
			}
			q = append(q, e)
			for len(q) > 0 {
				select {
				case out <- q[0]:
					q = q[1:]
				case e, ok := <-in:
					if ok {
						q = append(q, e)
						break
					}
					for _, e := range q {
						out <- e
					}
					close(out)
					return
				}
			}
		}
	}()
	return in, out
}
```

In the above implementation, we created two unbuffered channels. To not block the communication, a separate goroutine is created from the call. Whenever there is a send operation, it appends to a buffer `q`. To send the value to the receiver, a nested select loop that checks whether send is possible or not. If not, it keeps appending the data to the queue `q`.

When the input channel is closed, an additional loop over the queue `q` is used to run out all cached elements, then close the output channel.

Hence, another fix of the deadlock using unbounded channel would be:

```diff
func main() {
-	draw := make(chan interface{})
+	drawIn, drawOut := MakeChan()

	...

	// Rendering Thread
	go func() {
		...
		for {
			select {
			case size := <-change:
				...
			default:
-				draw <- p.Draw()
+				drawIn <- p.Draw()
			}
		}
	}()

	// Event Thread
	event := time.NewTicker(100 * time.Millisecond)
	for {
		select {
-		case id := <-draw:
+		case id := <-drawOut:
			println(id)
		case <-event.C:
			...
		}
	}
}
```

This unbounded channel is very similar to the commonly used standard graphics API pattern: CommandBuffer, a buffer that caches a series of draw calls, and does batch execution of a chunk of draw calls.

## Conclusion: A Generic Channel Abstraction

In this article, we discussed a form of deadlock in the select statement and two possible ways to address it. In the second approach, we discussed a possible way of implementing an unbounded channel construction. The implementation constructs an `interface{}` typed channel. 

We may ask ourselves, does unbounded make sense to have in the Go language with this particular example? Does the Go team ever consider such usage?

The answer to the second question is: Yes. They do, see [golang/go#20352](https://golang.org/issue/20352). The discussion thread shows that unbounded channels indeed serve a certain application, but clear drawbacks may hurt the application.
The major drawback is that an unbounded channel may run out of memory (OOM). If there is a concurrency bug, the running application will keep eats memory from OS and eventually leads to OOM. Developers argue that an unbounded channel should be added to the language mainly because the `MakeChan` function is returning an `interface{}` typed channel which brings weakly typed flaw into the statically typed Go code. Eventually, Ian Lance Taylor from the Go team [clarifies](https://golang.org/issue/20352#issuecomment-365438524) that an unbounded channel may have a sort of usage but is unworthy to be added to the language. As long as we have generics, a type-safe unbounded channel can be easily implemented in a library, answering the first question.

As of Go 1.18, soon we have type parameters, the above difficulty finally can be resolved. Here I provide a generic channel abstraction that is able to construct a type-safe, arbitrary sized channel:

```go
// MakeChan is a generic implementation that returns a sender and a
// receiver of an arbitrarily sized channel of an arbitrary type.
//
// If the given size is positive, the returned channel is a regular
// fix-sized buffered channel.
// If the given size is zero, the returned channel is an unbuffered channel.
// If the given size is -1, the returned an unbounded channel with an
// infinite-sized buffer.
func MakeChan[T any](size int) (chan<- T, <-chan T) {
	switch {
	case size == 0:
		ch := make(chan T)
		return ch, ch
	case size > 0:
		ch := make(chan T, size)
		return ch, ch
	case size != -1:
		panic("unbounded buffer size should be specified using -1")
	default:
		// size == -1
	}

	in, out := make(chan T), make(chan T)

	go func() {
		var q []T
		for {
			e, ok := <-in
			if !ok {
				close(out)
				return
			}
			q = append(q, e)
			for len(q) > 0 {
				select {
				case out <- q[0]:
					q = q[1:]
				case e, ok := <-in:
					if ok {
						q = append(q, e)
						break
					}
					for _, e := range q {
						out <- e
					}
					close(out)
					return
				}
			}
		}
	}()
	return in, out
}
```

```go
func main() {
	in, out := MakeChan[int](1)
	// in, out := MakeChan[int](0)
	// in, out := MakeChan[int](-1)

	go func() {
		in <- 42
	}()

	println(<-out)
}
```

This code is executable on go2go playground: https://go2goplay.golang.org/p/krLWm7ZInnL.

## Further Reading Suggestions

- Ian Lance Taylor. Type Parameters. March 19, 2021. https://golang.org/design/43651-type-parameters
- rgooch. proposal: spec: add support for unlimited capacity channels. 13 May 2017. https://golang.org/issue/20352
