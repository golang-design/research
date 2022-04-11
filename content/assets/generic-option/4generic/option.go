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

type Option[T A | B] func(a *T)

func V1[T A | B](v1 int) Option[T] {
	return func(a *T) {
		switch x := any(a).(type) {
		case *A:
			x.v1 = v1
		case *B:
			x.v1 = v1
		default:
			panic("unexpected use")
		}
	}
}

func V2[T B](v2 int) Option[T] {
	return func(a *T) {
		switch x := any(a).(type) {
		case *B:
			x.v2 = v2
		default:
			panic("unexpected use")
		}
	}
}

func NewA[T A](opts ...Option[T]) *T {
	t := new(T)
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func NewB[T B](opts ...Option[T]) *T {
	t := new(T)
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func main() {
	fmt.Printf("%#v\n", NewA())                     // &main.A{v1:0}
	fmt.Printf("%#v\n", NewA(V1[A](42)))            // &main.A{v1:42}
	fmt.Printf("%#v\n", NewB())                     // &main.B{v1:0, v2:0}
	fmt.Printf("%#v\n", NewB(V1[B](42)))            // &main.B{v1:42, v2:0}
	fmt.Printf("%#v\n", NewB(V2[B](42)))            // &main.B{v1:0, v2:42}
	fmt.Printf("%#v\n", NewB(V1[B](42), V2[B](42))) // &main.B{v1:42, v2:42}

	// _ = NewA(V2[B](42))            // ERROR: B does not implement A
	// _ = NewA(V2[A](42))            // ERROR: A does not implement B
	// _ = NewB(V1[A](42), V2[B](42)) // ERROR: type Option[B] of V2[B](42) does not match inferred type Option[A] for Option[T]
	// _ = NewB(V1[B](42), V2[A](42)) // ERROR: type Option[A] of V2[A](42) does not match inferred type Option[B] for Option[T]
}
