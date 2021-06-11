// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

package main

/*
#include <stdint.h>

extern void GoFunc(uintptr_t handle);
void cFunc(uintptr_t handle);
*/
import "C"
import (
	"cgo-handle/cgo"
	"log"
)

var meaningOfLife = 42

//export GoFunc
func GoFunc(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	ch := h.Value().(chan int)
	ch <- meaningOfLife
}

func main() {
	// Say we would like to pass the channel to a C function, then pass
	// it back from C to Go side and send some value.
	ch := make(chan int)

	h := cgo.NewHandle(ch)
	go func() {
		C.cFunc(C.uintptr_t(h))
	}()

	v := <-ch
	if meaningOfLife != v {
		log.Fatalf("unexpected receiving value: got %d, want %d", v, meaningOfLife)
	}
	h.Delete()
}