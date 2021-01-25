// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package mainthread

import (
	"runtime"
	"sync"
)

func init() {
	runtime.LockOSThread()
}

var (
	funcQ    = make(chan funcdata, runtime.GOMAXPROCS(0))
	donePool = sync.Pool{New: func() interface{} {
		return make(chan struct{})
	}}
)

type funcdata struct {
	fn   func()
	done chan struct{}
}

// Init initializes the functionality of running arbitrary subsequent
// functions be called on the main system thread.
//
// Init must be called in the main.main function.
func Init(main func()) {
	done := donePool.Get().(chan struct{})
	defer donePool.Put(done)

	go func() {
		main()
		done <- struct{}{}
	}()

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

// Call calls f on the main thread and blocks until f finishes.
func Call(f func()) {
	done := donePool.Get().(chan struct{})
	defer donePool.Put(done)

	funcQ <- funcdata{fn: f, done: done}
	<-done
}
