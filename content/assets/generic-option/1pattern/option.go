// Copyright 2022 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

package main

import "fmt"

type A struct {
	v1 int
	// Removed, now moved to v3.
	// v2 int
	v3 int
}

type Option func(*A)

func V1(v1 int) Option {
	return func(a *A) {
		a.v1 = v1
	}
}

// Deprecated: Use V3 instead.
func V2(v2 int) Option {
	return func(a *A) {
		// no effects anymore
		// a.v2 = v2
	}
}

func V3(v3 int) Option {
	return func(a *A) {
		a.v3 = v3
	}
}

func NewA(opts ...Option) *A {
	a := &A{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func main() {
	fmt.Printf("%#v\n", NewA())               // &main.A{v1:0, v2:0}
	fmt.Printf("%#v\n", NewA(V1(42)))         // &main.A{v1:42, v2:0}
	fmt.Printf("%#v\n", NewA(V2(42)))         // &main.A{v1:0, v2:0}
	fmt.Printf("%#v\n", NewA(V1(42), V2(42))) // &main.A{v1:42, v2:0}
}
