// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"runtime"
	"runtime/trace"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	run := flag.Bool("run", false, "start test")
	traceF := flag.String("trace", "trace.out", "trace file, default: trace.out")
	traceT := flag.String("d", "10s", "trace duration, default: 10s")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: go run main.go -run [-trace FILENAME -d DURATION]
options:
`)
		flag.PrintDefaults()
	}
	flag.Parse()
	if !*run {
		flag.Usage()
		os.Exit(2)
	}

	d, err := time.ParseDuration(*traceT)
	if err != nil {
		flag.Usage()
		os.Exit(2)
	}

	err = glfw.Init()
	if err != nil {
		panic(err)
	}

	w, err := glfw.CreateWindow(800, 600, "image", nil, nil)
	if err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
	w.MakeContextCurrent()

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	f, err := os.Open("../ou.png")
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	go func() {
		f, _ := os.Create(*traceF)
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
		time.Sleep(d)
		w.SetShouldClose(true)
	}()

	total := time.Duration(0)
	count := 0
	for !w.ShouldClose() {
		t := time.Now()
		flush(w, img, img.Bounds())
		total += time.Now().Sub(t)
		count++

		glfw.WaitEventsTimeout(1.0 / 30)
	}
	timePerFlush := total / time.Duration(count)
	fmt.Printf("%v, fps: %v\n", timePerFlush, int(time.Second/timePerFlush))
}

func flush(w *glfw.Window, img image.Image, canvas image.Rectangle) {
	bounds := img.Bounds()
	canvas = canvas.Intersect(bounds)
	if canvas.Empty() {
		return
	}

	tmp := image.NewRGBA(canvas)
	draw.Draw(tmp, canvas, img, canvas.Min, draw.Src)
	gl.DrawBuffer(gl.FRONT)
	gl.Viewport(
		int32(bounds.Min.X), int32(bounds.Min.Y),
		int32(bounds.Dx()), int32(bounds.Dy()),
	)
	gl.RasterPos2d(
		-1+2*float64(canvas.Min.X)/float64(bounds.Dx()),
		+1-2*float64(canvas.Min.Y)/float64(bounds.Dy()),
	)
	gl.PixelZoom(1, -1)
	gl.DrawPixels(
		int32(canvas.Dx()), int32(canvas.Dy()),
		gl.RGBA, gl.UNSIGNED_BYTE,
		unsafe.Pointer(&tmp.Pix[0]),
	)
	gl.Flush()
}
