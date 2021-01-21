// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package mainthread_test

import (
	"testing"

	"x/mainthread"
)

/*
bench: run benchmarks under 90% cpufreq...
bench: go test -run=^$ -bench=. -count=10
goos: linux
goarch: amd64
pkg: x/mainthread
cpu: Intel(R) Core(TM) i9-9900K CPU @ 3.60GHz

name      time/op
Call-16   391ns ±0%
CallV-16  398ns ±0%

name      alloc/op
Call-16                    0.00B
CallV-16                   0.00B

name      allocs/op
Call-16                     0.00
CallV-16                    0.00
*/

func BenchmarkCall(b *testing.B) {
	f := func() {}
	mainthread.Init(func() {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mainthread.Call(f)
		}
	})
}

func BenchmarkCallV(b *testing.B) {
	f := func() interface{} {
		return true
	}
	mainthread.Init(func() {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mainthread.CallV(f).(bool)
		}
	})
}
