// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

// Package cgo is an implementation of golang.org/issue/37033.
//
// See golang.org/cl/294670 for code review discussion.
package cgo1

import (
	"reflect"
	"sync"
)

// Handle provides a way to pass values that contain Go pointers
// (pointers to memory allocated by Go) between Go and C without
// breaking the cgo pointer passing rules. A Handle is an integer
// value that can represent any Go value. A Handle can be passed
// through C and back to Go, and Go code can use the Handle to
// retrieve the original Go value.
//
// The underlying type of Handle is guaranteed to fit in an integer type
// that is large enough to hold the bit pattern of any pointer. The zero
// value of a Handle is not valid, and thus is safe to use as a sentinel
// in C APIs.
//
// For instance, on the Go side:
//
//	package main
//
//	/*
//	#include <stdint.h> // for uintptr_t
//
//	extern void MyGoPrint(uintptr_t handle);
//	void myprint(uintptr_t handle);
//	*/
//	import "C"
//	import "runtime/cgo"
//
//	//export MyGoPrint
//	func MyGoPrint(handle C.uintptr_t) {
//		h := cgo.Handle(handle)
//		val := h.Value().(string)
//		println(val)
//		h.Delete()
//	}
//
//	func main() {
//		val := "hello Go"
//		C.myprint(C.uintptr_t(cgo.NewHandle(val)))
//		// Output: hello Go
//	}
//
// and on the C side:
//
//	#include <stdint.h> // for uintptr_t
//
//	// A Go function
//	extern void MyGoPrint(uintptr_t handle);
//
//	// A C function
//	void myprint(uintptr_t handle) {
//	    MyGoPrint(handle);
//	}
type Handle uintptr

// NewHandle returns a handle for a given value.
//
// The handle is valid until the program calls Delete on it. The handle
// uses resources, and this package assumes that C code may hold on to
// the handle, so a program must explicitly call Delete when the handle
// is no longer needed.
//
// The intended use is to pass the returned handle to C code, which
// passes it back to Go, which calls Value.
func NewHandle(v interface{}) Handle {
	var k uintptr

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Slice,
		reflect.Map, reflect.Chan, reflect.Func:
		if rv.IsNil() {
			panic("cgo: cannot use Handle for nil value")
		}

		k = rv.Pointer()
	default:
		// Wrap and turn a value parameter into a pointer. This enables
		// us to always store the passing object as a pointer, and helps
		// to identify which of whose are initially pointers or values
		// when Value is called.
		v = &wrap{v}
		k = reflect.ValueOf(v).Pointer()
	}

	// v was escaped to the heap because of reflection. As Go do not have
	// a moving GC (and possibly lasts true for a long future), it is
	// safe to use its pointer address as the key of the global map at
	// this moment. The implementation must be reconsidered if moving GC
	// is introduced internally in the runtime.
	actual, loaded := m.LoadOrStore(k, v)
	if !loaded {
		return Handle(k)
	}

	arv := reflect.ValueOf(actual)
	switch arv.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Slice,
		reflect.Map, reflect.Chan, reflect.Func:
		// The underlying object of the given Go value already have
		// its existing handle.
		if arv.Pointer() == k {
			return Handle(k)
		}

		// If the loaded pointer is inconsistent with the new pointer,
		// it means the address has been used for different objects
		// because of GC and its address is reused for a new Go object,
		// meaning that the Handle does not call Delete explicitly when
		// the old Go value is not needed. Consider this as a misuse of
		// a handle, do panic.
		panic("cgo: misuse of a Handle")
	default:
		panic("cgo: Handle implementation has an internal bug")
	}
}

// Value returns the associated Go value for a valid handle.
//
// The method panics if the handle is invalid.
func (h Handle) Value() interface{} {
	v, ok := m.Load(uintptr(h))
	if !ok {
		panic("cgo: misuse of an invalid Handle")
	}
	if wv, ok := v.(*wrap); ok {
		return wv.v
	}
	return v
}

// Delete invalidates a handle. This method should only be called once
// the program no longer needs to pass the handle to C and the C code
// no longer has a copy of the handle value.
//
// The method panics if the handle is invalid.
func (h Handle) Delete() {
	_, ok := m.LoadAndDelete(uintptr(h))
	if !ok {
		panic("cgo: misuse of an invalid Handle")
	}
}

var m = &sync.Map{} // map[uintptr]interface{}

// wrap wraps a Go value.
type wrap struct{ v interface{} }
