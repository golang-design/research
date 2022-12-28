package main

import (
	"testing"
	"unsafe"
)

// Go 1.17: func Slice(ptr *ArbitraryType, len IntegerType) []ArbitraryType
// Go 1.20: func SliceData(slice []ArbitraryType) *ArbitraryType
// Go 1.20: func String(ptr *byte, len IntegerType) string
// Go 1.20: func StringData(str string) *byte

func string2bytes(x string) []byte { return unsafe.Slice(unsafe.StringData(x), len(x)) }
func bytes2string(x []byte) string { return unsafe.String(unsafe.SliceData(x), len(x)) }

func BenchmarkSliceToString(b *testing.B) {
	x := []byte("this is a string")

	b.Run("type-cast", func(b *testing.B) {
		var s string
		for i := 0; i < b.N; i++ {
			s = string(x)
		}
		_ = s
	})

	b.Run("unsafe-conv", func(b *testing.B) {
		var s string
		for i := 0; i < b.N; i++ {
			s = bytes2string(x)
		}
		_ = s
	})
}

func BenchmarkStringToSlice(b *testing.B) {
	x := "this is a string"

	b.Run("type-cast", func(b *testing.B) {
		var s []byte
		for i := 0; i < b.N; i++ {
			s = []byte(x)
		}
		_ = s
	})

	b.Run("unsafe-conv", func(b *testing.B) {
		var s []byte
		for i := 0; i < b.N; i++ {
			s = string2bytes(x)
		}
		_ = s
	})
}
