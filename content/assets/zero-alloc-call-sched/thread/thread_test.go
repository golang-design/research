// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

// +build linux

package thread_test

import (
	"testing"
	"x/thread"

	"golang.org/x/sys/unix"
)

func TestThread(t *testing.T) {
	lastThId := 0

	th := thread.New()
	th.Call(func() {
		lastThId = unix.Gettid()
		t.Logf("thread id: %d", lastThId)
	})
	var failed bool
	th.Call(func() {
		if unix.Gettid() != lastThId {
			failed = true
		}
		lastThId = unix.Gettid()
		t.Logf("thread id: %d", lastThId)
	})
	if failed {
		t.Fatalf("failed to schedule function on the same thread.")
	}
	th.Terminate()
	th.Terminate()

	th.Call(func() {
		panic("unexpected call")
	})
	th.Call(func() {
		panic("unexpected call")
	})

	th = thread.New()
	th.Call(func() {
		if unix.Gettid() == lastThId {
			failed = true
		}
		lastThId = unix.Gettid()
		t.Logf("thread id: %d", lastThId)
	})
	if failed {
		t.Fatalf("failed to schedule function on a different thread.")
	}
	th.Call(func() {
		if unix.Gettid() != lastThId {
			failed = true
		}
		lastThId = unix.Gettid()
		t.Logf("thread id: %d", lastThId)
	})
	if failed {
		t.Fatalf("failed to schedule function on the same thread.")
	}
}
