# Pointer Type May Not Be Ideal for Parameters

Author(s): [Changkun Ou](https://changkun.de)

Last updated: 2020-10-27

## Introduction

We are aware of that using pointers for passing parameters can avoid data copy, which will benefit the prformance. But there are always some edge cases you might need concern.

Let's check this example:

```go
// vec.go
type vec1 struct {
	x, y, z, w float64
}

func (v vec1) add(u vec1) vec1 {
	return vec1{v.x + u.x, v.y + u.y, v.z + u.z, v.w + u.w}
}

type vec2 struct {
	x, y, z, w float64
}

func (v *vec2) add(u *vec2) *vec2 {
	v.x += u.x
	v.y += u.y
	v.z += u.z
	v.w += u.w
	return v
}
```

Which `add` implementation runs faster?
Intuitively, we might think that `vec2` is faster because its parameter `u` uses pointer and there should have no copies on the data, whereas `vec1` involves data copy both when passing and returning.

However, if we write a benchmark:

```go
func BenchmarkVec(b *testing.B) {
	b.ReportAllocs()
	b.Run("vec1", func(b *testing.B) {
		v1 := vec1{1, 2, 3, 4}
		v2 := vec1{4, 5, 6, 7}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				v1 = v1.add(v2)
			} else {
				v2 = v2.add(v1)
			}
		}
	})
	b.Run("vec2", func(b *testing.B) {
		v1 := vec2{1, 2, 3, 4}
		v2 := vec2{4, 5, 6, 7}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				v1.add(&v2)
			} else {
				v2.add(&v1)
			}
		}
	})
}
```

And run as follows: 

```sh
$ perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee new.txt
$ benchstat new.txt
```

The `benchstat` will give you the following result:

```
name         time/op
Vec/vec1-16  0.25ns ± 1%
Vec/vec2-16  2.20ns ± 0%
```

How is this happening?

## Inlining Optimization

This is all because of compiler optimization, and mostly because of inlining.

If we disable inline from the `add`:

```go
// vec.go
type vec1 struct {
	x, y, z, w float64
}

//go:noinline
func (v vec1) add(u vec1) vec1 {
	return vec1{v.x + u.x, v.y + u.y, v.z + u.z, v.w + u.w}
}

type vec2 struct {
	x, y, z, w float64
}

//go:noinline
func (v *vec2) add(u *vec2) *vec2 {
	v.x += u.x
	v.y += u.y
	v.z += u.z
	v.w += u.w
	return v
}
```

Run the benchmark and compare the perf with the previous one:

```sh
$ perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee old.txt
$ benchstat old.txt new.txt
name         old time/op  new time/op  delta
Vec/vec1-16  4.92ns ± 1%  0.25ns ± 1%  -95.01%  (p=0.000 n=10+9)
Vec/vec2-16  2.89ns ± 1%  2.20ns ± 0%  -23.77%  (p=0.000 n=10+8)
```

The inline optimization transforms the code:

```go
v1 := vec1{1, 2, 3, 4}
v2 := vec1{4, 5, 6, 7}
v1 = v1.add(v2)
```

to a direct assign statement:

```go
v1 := vec1{1, 2, 3, 4}
v2 := vec1{4, 5, 6, 7}
v1 = vec1{1+4, 2+5, 3+6, 4+7}
```

And for the `vec2`'s case:

```go
v1 := vec2{1, 2, 3, 4}
v2 := vec2{4, 5, 6, 7}
v1 = v1.add(v2)
```

to a direct manipulation:

```go
v1 := vec2{1, 2, 3, 4}
v2 := vec2{4, 5, 6, 7}
v1.x += v2.x
v1.y += v2.y
v1.z += v2.z
v1.w += v2.w
```

## Unoptimized Move Semantics

If we check the compiled assembly, the reason reveals quickly:

```sh
$ go tool compile -S vec.go > vec.s
```

The dumped assumbly code is as follows:

```asm
"".vec1.add STEXT nosplit size=89 args=0x60 locals=0x0
	0x0000 00000 (vec.go:8)	TEXT	"".vec1.add(SB), NOSPLIT|ABIInternal, $0-96
	0x0000 00000 (vec.go:8)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (vec.go:8)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (vec.go:9)	MOVSD	"".u+40(SP), X0
	0x0006 00006 (vec.go:9)	MOVSD	"".v+8(SP), X1
	0x000c 00012 (vec.go:9)	ADDSD	X1, X0
	0x0010 00016 (vec.go:9)	MOVSD	X0, "".~r1+72(SP)
	0x0016 00022 (vec.go:9)	MOVSD	"".u+48(SP), X0
	0x001c 00028 (vec.go:9)	MOVSD	"".v+16(SP), X1
	0x0022 00034 (vec.go:9)	ADDSD	X1, X0
	0x0026 00038 (vec.go:9)	MOVSD	X0, "".~r1+80(SP)
	0x002c 00044 (vec.go:9)	MOVSD	"".u+56(SP), X0
	0x0032 00050 (vec.go:9)	MOVSD	"".v+24(SP), X1
	0x0038 00056 (vec.go:9)	ADDSD	X1, X0
	0x003c 00060 (vec.go:9)	MOVSD	X0, "".~r1+88(SP)
	0x0042 00066 (vec.go:9)	MOVSD	"".u+64(SP), X0
	0x0048 00072 (vec.go:9)	MOVSD	"".v+32(SP), X1
	0x004e 00078 (vec.go:9)	ADDSD	X1, X0
	0x0052 00082 (vec.go:9)	MOVSD	X0, "".~r1+96(SP)
	0x0058 00088 (vec.go:9)	RET
"".(*vec2).add STEXT nosplit size=73 args=0x18 locals=0x0
	0x0000 00000 (vec.go:17)	TEXT	"".(*vec2).add(SB), NOSPLIT|ABIInternal, $0-24
	0x0000 00000 (vec.go:17)	FUNCDATA	$0, gclocals·8f9cec06d1ae35cc9900c511c5e4bdab(SB)
	0x0000 00000 (vec.go:17)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (vec.go:18)	MOVQ	"".u+16(SP), AX
	0x0005 00005 (vec.go:18)	MOVSD	(AX), X0
	0x0009 00009 (vec.go:18)	MOVQ	"".v+8(SP), CX
	0x000e 00014 (vec.go:18)	ADDSD	(CX), X0
	0x0012 00018 (vec.go:18)	MOVSD	X0, (CX)
	0x0016 00022 (vec.go:19)	MOVSD	8(AX), X0
	0x001b 00027 (vec.go:19)	ADDSD	8(CX), X0
	0x0020 00032 (vec.go:19)	MOVSD	X0, 8(CX)
	0x0025 00037 (vec.go:20)	MOVSD	16(CX), X0
	0x002a 00042 (vec.go:20)	ADDSD	16(AX), X0
	0x002f 00047 (vec.go:20)	MOVSD	X0, 16(CX)
	0x0034 00052 (vec.go:21)	MOVSD	24(AX), X0
	0x0039 00057 (vec.go:21)	ADDSD	24(CX), X0
	0x003e 00062 (vec.go:21)	MOVSD	X0, 24(CX)
	0x0043 00067 (vec.go:22)	MOVQ	CX, "".~r1+24(SP)
	0x0048 00072 (vec.go:22)	RET
```

The `add` implementation of `vec1` uses values from the previous stack frame and writes the result directly to the return;
whereas `vec2` needs MOVQ that copies the parameter to different registers (e.g., copy pointers to AX and CX,), then write back to the return.

The unexpected move cost in `vec2` is the additional two `MOVQ` instructions and read operations on the two pointer addresses.

## Further Reading Suggestions

- Changkun Ou. Conduct Reliable Benchmarking in Go. March 26, 2020. https://golang.design/s/gobench
- Dave Cheney. Mid-stack inlining in Go. May 2, 2020. https://dave.cheney.net/2020/05/02/mid-stack-inlining-in-go
- Dave Cheney. Inlining optimisations in Go. April 25, 2020. https://dave.cheney.net/2020/04/25/inlining-optimisations-in-go
- MOVSD. Move or Merge Scalar Double-Precision Floating-Point Value. Last access: 2020-10-27. https://www.felixcloutier.com/x86/movsd
- ADDSD. Add Scalar Double-Precision Floating-Point Values. Last access: 2020-10-27. https://www.felixcloutier.com/x86/addsd
- MOVEQ. Move Quadword. Last access: 2020-10-27. https://www.felixcloutier.com/x86/movq

## License

Copyright &copy; 2020 The [golang.design](https://golang.design) Authors.
