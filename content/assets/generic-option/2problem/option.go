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

type OptionA func(a *A)
type OptionB func(a *B)

func V1ForA(v1 int) OptionA {
	return func(a *A) {
		a.v1 = v1
	}
}

func V1ForB(v1 int) OptionB {
	return func(b *B) {
		b.v1 = v1
	}
}

func V2ForB(v2 int) OptionB {
	return func(b *B) {
		b.v2 = v2
	}
}

func NewA(opts ...OptionA) *A {
	a := &A{}

	for _, opt := range opts {
		opt(a)
	}
	return a
}

func NewB(opts ...OptionB) *B {
	b := &B{}

	for _, opt := range opts {
		opt(b)
	}
	return b
}

func main() {
	fmt.Printf("%#v\n", NewA())                       // &main.A{v1:0}
	fmt.Printf("%#v\n", NewA(V1ForA(42)))             // &main.A{v1:42}
	fmt.Printf("%#v\n", NewB())                       // &main.B{v1:0, v2:0}
	fmt.Printf("%#v\n", NewB(V1ForB(42)))             // &main.B{v1:42, v2:0}
	fmt.Printf("%#v\n", NewB(V2ForB(42)))             // &main.B{v1:0, v2:42}
	fmt.Printf("%#v\n", NewB(V1ForB(42), V2ForB(42))) // &main.B{v1:42, v2:42}
}
