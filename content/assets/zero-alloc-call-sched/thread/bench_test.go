package thread_test

import (
	"testing"
	"x/thread"
)

func BenchmarkThreadCall(b *testing.B) {
	th := thread.New()
	f := func() {}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		th.Call(f)
	}
}
