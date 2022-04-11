// Copyright 2022 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

package main

import "fmt"

type A struct {
	v1 int
}

type B struct {
	v1 int
	v2 int
}

type Common interface {
	/* ... */
}

type Option func(c Common)

func V1(v1 int) Option {
	return func(c Common) {
		switch x := c.(type) {
		case *A:
			x.v1 = v1
		case *B:
			x.v1 = v1
		default:
			panic("unexpected use")
		}
	}
}

func V2(v2 int) Option {
	return func(c Common) {
		switch x := c.(type) {
		case *B:
			x.v2 = v2
		default:
			panic("unexpected use")
		}
	}
}

func NewA(opts ...Option) *A {
	a := &A{}

	for _, opt := range opts {
		opt(a)
	}
	return a
}

func NewB(opts ...Option) *B {
	b := &B{}

	for _, opt := range opts {
		opt(b)
	}
	return b
}

func main() {
	fmt.Printf("%#v\n", NewA())               // &main.A{v1:0}
	fmt.Printf("%#v\n", NewA(V1(42)))         // &main.A{v1:42}
	fmt.Printf("%#v\n", NewB())               // &main.B{v1:0, v2:0}
	fmt.Printf("%#v\n", NewB(V1(42)))         // &main.B{v1:42, v2:0}
	fmt.Printf("%#v\n", NewB(V2(42)))         // &main.B{v1:0, v2:42}
	fmt.Printf("%#v\n", NewB(V1(42), V2(42))) // &main.B{v1:42, v2:42}

	// fmt.Printf("%#v\n", NewA(V2(42))) // ERROR: panic
}
