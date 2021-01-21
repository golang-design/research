// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package thread

import (
	"runtime"
	"sync"
)

var donePool = sync.Pool{
	New: func() interface{} {
		return make(chan struct{})
	},
}

func init() {
	runtime.LockOSThread()
}

type funcData struct {
	fn   func()
	done chan struct{}
}

// Thread offers facilities to schedule function calls to run
// on a same thread.
type Thread struct {
	f         chan funcData
	terminate chan struct{}
}

// Call calls f on the given thread.
func (t *Thread) Call(f func()) bool {
	if f == nil {
		return false
	}
	select {
	case <-t.terminate:
		return false
	default:
		done := donePool.Get().(chan struct{})
		defer donePool.Put(done)
		defer func() {
			<-done
		}()

		t.f <- funcData{fn: f, done: done}
	}
	return true
}

// Terminate terminates the current thread.
func (t *Thread) Terminate() {
	select {
	case <-t.terminate:
		return
	default:
		t.terminate <- struct{}{}
	}
}

// New creates
func New() *Thread {
	t := Thread{
		f:         make(chan funcData),
		terminate: make(chan struct{}),
	}
	go func() {
		runtime.LockOSThread()
		for {
			select {
			case f := <-t.f:
				func() {
					defer func() {
						f.done <- struct{}{}
					}()
					f.fn()
				}()
			case <-t.terminate:
				close(t.terminate)
				return
			}
		}
	}()
	return &t
}
