# Pointer Type May Not Be Ideal for Parameters

Author(s): [Changkun Ou](https://changkun.de)

Last updated: 2020-10-27

## Introduction

We are aware that using pointers for passing parameters can avoid data copy,
which will benefit the performance. Nevertheless, there are always some
edge cases we might need concern.

Let's take this as an example:

```go
// vec.go
type vec struct {
	x, y, z, w float64
}

func (v vec) addv(u vec) vec {
	return vec{v.x + u.x, v.y + u.y, v.z + u.z, v.w + u.w}
}

func (v *vec) addp(u *vec) *vec {
	v.x, v.y, v.z, v.w = v.x+u.x, v.y+u.y, v.z+u.z, v.w+u.w
	return v
}
```

Which vector addition runs faster?
Intuitively, we might consider that `vec.addp` is faster than `vec.addv`
because its parameter `u` uses pointer form. There should be no copies
of the data, whereas `vec.addv` involves data copy both when passing and
returning.

However, if we do a micro-benchmark:

```go
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
```

And run as follows: 

```sh
$ perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee new.txt
$ benchstat new.txt
```

The `benchstat` will give you the following result:

```
name         time/op
Vec/addv-16  0.25ns ± 2%
Vec/addp-16  2.20ns ± 0%

name         alloc/op
Vec/addv-16   0.00B     
Vec/addp-16   0.00B     

name         allocs/op
Vec/addv-16    0.00     
Vec/addp-16    0.00
```

How is this happening?

## Inlining Optimization

This is all because of compiler optimization, and mostly because of inlining.

If we disable inline from the `addv` and `addp`:

```go
//go:noinline
func (v vec) addv(u vec) vec {
	return vec{v.x + u.x, v.y + u.y, v.z + u.z, v.w + u.w}
}

//go:noinline
func (v *vec) addp(u *vec) *vec {
	v.x, v.y, v.z, v.w = v.x+u.x, v.y+u.y, v.z+u.z, v.w+u.w
	return v
}
```

Then run the benchmark and compare the perf with the previous one:

```sh
$ perflock -governor 80% go test -v -run=none -bench=. -count=10 | tee old.txt
$ benchstat old.txt new.txt
name         old time/op    new time/op    delta
Vec/addv-16    4.99ns ± 1%    0.25ns ± 2%  -95.05%  (p=0.000 n=9+10)
Vec/addp-16    3.35ns ± 1%    2.20ns ± 0%  -34.37%  (p=0.000 n=10+8)
```

The inline optimization transforms the `vec.addv`:

```go
v1 := vec{1, 2, 3, 4}
v2 := vec{4, 5, 6, 7}
v1 = v1.addv(v2)
```

to a direct assign statement:

```go
v1 := vec{1, 2, 3, 4}
v2 := vec{4, 5, 6, 7}
v1 = vec{1+4, 2+5, 3+6, 4+7}
```

And for the `vec.addp`'s case:

```go
v1 := &vec{1, 2, 3, 4}
v2 := &vec{4, 5, 6, 7}
v1 = v1.addp(v2)
```

to a direct manipulation:

```go
v1 := vec{1, 2, 3, 4}
v2 := vec{4, 5, 6, 7}
v1.x, v1.y, v1.z, v1.w = v1.x+v2.x, v1.y+v2.y, v1.z+v2.z, v1.w+v2.w 
```

## Addressing Modes

If we check the compiled assembly, the reason reveals quickly:

```sh
$ mkdir asm && go tool compile -S vec.go > asm/vec.s
```

The dumped assumbly code is as follows:

```asm
"".vec.addv STEXT nosplit size=89 args=0x60 locals=0x0 funcid=0x0
	0x0000 00000 (vec.go:7)	TEXT	"".vec.addv(SB), NOSPLIT|ABIInternal, $0-96
	0x0000 00000 (vec.go:7)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (vec.go:7)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (vec.go:8)	MOVSD	"".u+40(SP), X0
	0x0006 00006 (vec.go:8)	MOVSD	"".v+8(SP), X1
	0x000c 00012 (vec.go:8)	ADDSD	X1, X0
	0x0010 00016 (vec.go:8)	MOVSD	X0, "".~r1+72(SP)
	0x0016 00022 (vec.go:8)	MOVSD	"".u+48(SP), X0
	0x001c 00028 (vec.go:8)	MOVSD	"".v+16(SP), X1
	0x0022 00034 (vec.go:8)	ADDSD	X1, X0
	0x0026 00038 (vec.go:8)	MOVSD	X0, "".~r1+80(SP)
	0x002c 00044 (vec.go:8)	MOVSD	"".u+56(SP), X0
	0x0032 00050 (vec.go:8)	MOVSD	"".v+24(SP), X1
	0x0038 00056 (vec.go:8)	ADDSD	X1, X0
	0x003c 00060 (vec.go:8)	MOVSD	X0, "".~r1+88(SP)
	0x0042 00066 (vec.go:8)	MOVSD	"".u+64(SP), X0
	0x0048 00072 (vec.go:8)	MOVSD	"".v+32(SP), X1
	0x004e 00078 (vec.go:8)	ADDSD	X1, X0
	0x0052 00082 (vec.go:8)	MOVSD	X0, "".~r1+96(SP)
	0x0058 00088 (vec.go:8)	RET
"".(*vec).addp STEXT nosplit size=73 args=0x18 locals=0x0 funcid=0x0
	0x0000 00000 (vec.go:11)	TEXT	"".(*vec).addp(SB), NOSPLIT|ABIInternal, $0-24
	0x0000 00000 (vec.go:11)	FUNCDATA	$0, gclocals·522734ad228da40e2256ba19cf2bc72c(SB)
	0x0000 00000 (vec.go:11)	FUNCDATA	$1, gclocals·69c1753bd5f81501d95132d08af04464(SB)
	0x0000 00000 (vec.go:12)	MOVQ	"".u+16(SP), AX
	0x0005 00005 (vec.go:12)	MOVSD	(AX), X0
	0x0009 00009 (vec.go:12)	MOVQ	"".v+8(SP), CX
	0x000e 00014 (vec.go:12)	ADDSD	(CX), X0
	0x0012 00018 (vec.go:12)	MOVSD	8(AX), X1
	0x0017 00023 (vec.go:12)	ADDSD	8(CX), X1
	0x001c 00028 (vec.go:12)	MOVSD	16(CX), X2
	0x0021 00033 (vec.go:12)	ADDSD	16(AX), X2
	0x0026 00038 (vec.go:12)	MOVSD	24(AX), X3
	0x002b 00043 (vec.go:12)	ADDSD	24(CX), X3
	0x0030 00048 (vec.go:12)	MOVSD	X0, (CX)
	0x0034 00052 (vec.go:12)	MOVSD	X1, 8(CX)
	0x0039 00057 (vec.go:12)	MOVSD	X2, 16(CX)
	0x003e 00062 (vec.go:12)	MOVSD	X3, 24(CX)
	0x0043 00067 (vec.go:13)	MOVQ	CX, "".~r1+24(SP)
	0x0048 00072 (vec.go:13)	RET
```

The `addv` implementation uses values from the previous stack frame and
writes the result directly to the return; whereas `addp` needs MOVQ that
copies the parameter to different registers (e.g., copy pointers to AX and CX,),
then write back when returning. Therefore, another unexpected cost in
`addp` is caused by the indirect addressing mode for accessing the memory unit.

## Further Reading Suggestions

- Changkun Ou. Conduct Reliable Benchmarking in Go. March 26, 2020. https://golang.design/s/gobench
- Dave Cheney. Mid-stack inlining in Go. May 2, 2020. https://dave.cheney.net/2020/05/02/mid-stack-inlining-in-go
- Dave Cheney. Inlining optimisations in Go. April 25, 2020. https://dave.cheney.net/2020/04/25/inlining-optimisations-in-go
- MOVSD. Move or Merge Scalar Double-Precision Floating-Point Value. Last access: 2020-10-27. https://www.felixcloutier.com/x86/movsd
- ADDSD. Add Scalar Double-Precision Floating-Point Values. Last access: 2020-10-27. https://www.felixcloutier.com/x86/addsd
- MOVEQ. Move Quadword. Last access: 2020-10-27. https://www.felixcloutier.com/x86/movq

## License

Copyright &copy; 2020 The [golang.design](https://golang.design) Authors.
