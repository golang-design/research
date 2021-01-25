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

// Init initializes the functionality of running arbitrary subsequent
// functions be called on the main system thread.
//
// Init must be called in the main.main function.
func Init(main func()) {
	done := donePool.Get().(chan struct{})
	defer donePool.Put(done)
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		main()
	}()

	for {
		select {
		case f := <-funcQ:
			if f.fn != nil {
				f.fn()
				f.done <- struct{}{}
			} else if f.fnv != nil {
				f.ret <- f.fnv()
			}
		case <-done:
			return
		}
	}
}

// Call calls f on the main thread and blocks until f finishes.
func Call(f func()) {
	done := donePool.Get().(chan struct{})
	defer donePool.Put(done)
	funcQ <- funcData{fn: f, done: done}
	<-done
}

// CallV calls f on the main thread. It returns what f returns.
func CallV(f func() interface{}) interface{} {
	ret := retPool.Get().(chan interface{})
	defer retPool.Put(ret)

	funcQ <- funcData{fnv: f, ret: ret}
	return <-ret
}

var (
	funcQ    = make(chan funcData, runtime.GOMAXPROCS(0))
	donePool = sync.Pool{New: func() interface{} {
		return make(chan struct{})
	}}
	retPool = sync.Pool{New: func() interface{} {
		return make(chan interface{})
	}}
)

type funcData struct {
	fn   func()
	done chan struct{}
	fnv  func() interface{}
	ret  chan interface{}
}
