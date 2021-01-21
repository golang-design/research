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
	0x0000 f2 0f 10 44 24 28 f2 0f 10 4c 24 08 f2 0f 58 c1  ...D$(...L$...X.
	0x0010 f2 0f 11 44 24 48 f2 0f 10 44 24 30 f2 0f 10 4c  ...D$H...D$0...L
	0x0020 24 10 f2 0f 58 c1 f2 0f 11 44 24 50 f2 0f 10 44  $...X....D$P...D
	0x0030 24 38 f2 0f 10 4c 24 18 f2 0f 58 c1 f2 0f 11 44  $8...L$...X....D
	0x0040 24 58 f2 0f 10 44 24 40 f2 0f 10 4c 24 20 f2 0f  $X...D$@...L$ ..
	0x0050 58 c1 f2 0f 11 44 24 60 c3                       X....D$`.
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
	0x0000 48 8b 44 24 10 f2 0f 10 00 48 8b 4c 24 08 f2 0f  H.D$.....H.L$...
	0x0010 58 01 f2 0f 10 48 08 f2 0f 58 49 08 f2 0f 10 51  X....H...XI....Q
	0x0020 10 f2 0f 58 50 10 f2 0f 10 58 18 f2 0f 58 59 18  ...XP....X...XY.
	0x0030 f2 0f 11 01 f2 0f 11 49 08 f2 0f 11 51 10 f2 0f  .......I....Q...
	0x0040 11 59 18 48 89 4c 24 18 c3                       .Y.H.L$..
