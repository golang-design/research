// Copyright (c) 2020 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package pparam

import (
	"testing"
)

func BenchmarkVec(b *testing.B) {
	b.Run("addv", func(b *testing.B) {
		v1 := vec{1, 2, 3, 4}
		v2 := vec{4, 5, 6, 7}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				v1 = v1.addv(v2)
			} else {
				v2 = v2.addv(v1)
			}
		}
	})
	b.Run("addp", func(b *testing.B) {
		v1 := &vec{1, 2, 3, 4}
		v2 := &vec{4, 5, 6, 7}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				v1 = v1.addp(v2)
			} else {
				v2 = v2.addp(v1)
			}
		}
	})
}
