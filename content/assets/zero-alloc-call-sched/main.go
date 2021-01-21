// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/trace"
	"time"
	"x/app"
	"x/mainthread"
)

func main() {
	mainthread.Init(fn)
}

func fn() {
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

	w, err := app.NewWindow()
	if err != nil {
		panic(err)
	}
	defer app.Terminate()

	go func() {
		f, _ := os.Create(*traceF)
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
		time.Sleep(d)
		w.Stop()
	}()
	for !w.Closed() {
		w.Update()
	}
}
