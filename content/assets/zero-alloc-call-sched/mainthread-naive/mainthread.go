// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package mainthread

import (
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

var funcQ = make(chan func(), runtime.GOMAXPROCS(0))

// Init initializes the functionality of running arbitrary subsequent
// functions be called on the main system thread.
//
// Init must be called in the main.main function.
func Init(main func()) {
	done := make(chan struct{})
	go func() {
		main()
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

// Call calls f on the main thread and blocks until f finishes.
func Call(f func()) {
	done := make(chan struct{})
	funcQ <- func() {
		f()
		done <- struct{}{}
	}
	<-done
}
