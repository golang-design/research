package time_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkAtomic(b *testing.B) {
	var v int32
	var n = 1000000
	for k := 1; k < n; k *= 10 {
		b.Run(fmt.Sprintf("n-%d", k), func(b *testing.B) {
			atomic.StoreInt32(&v, 0)
			b.Run("with-timer", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					b.StartTimer()
					for j := 0; j < k; j++ {
						atomic.AddInt32(&v, 1)
					}
				}
			})
			atomic.StoreInt32(&v, 0)
			b.Run("w/o-timer", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for j := 0; j < k; j++ {
						atomic.AddInt32(&v, 1)
					}
				}
			})
		})
	}
}

func TestSolution(t *testing.T) {
	var v int32
	atomic.StoreInt32(&v, 0)
	r := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// ... do extra stuff ...
			b.StartTimer()
			atomic.AddInt32(&v, 1)
		}
	})

	// do calibration that removes the overhead of calling time.Now().
	calibrate := func(d time.Duration, n int) time.Duration {
		since := time.Duration(0)
		for i := 0; i < n; i++ {
			start := time.Now()
			since += time.Since(start)
		}
		return (d - since) / time.Duration(n)
	}

	fmt.Printf("%v ns/op\n", calibrate(r.T, r.N))
}
