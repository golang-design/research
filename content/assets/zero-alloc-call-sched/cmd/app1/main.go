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
	app "x/app-naive"
	mainthread "x/mainthread-opt2"
)

func main() {
	mainthread.Init(fn)
}

func fn() {
	d := parseArgs()

	err := app.Init()
	if err != nil {
		panic(err)
	}
	defer app.Terminate()
	w, err := app.NewWindow()
	if err != nil {
		panic(err)
	}

	done := make(chan struct{}, 2)
	go func() {
		f, _ := os.Create(*traceF)
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
		time.Sleep(d)
		w.Stop()
	}()
	go func() {
		w.Run()
		done <- struct{}{}
	}()
	<-done
}

var (
	run    *bool
	traceF *string
	traceT *string
)

func parseArgs() time.Duration {
	run = flag.Bool("run", false, "start test")
	traceF = flag.String("trace", "trace.out", "trace file, default: trace.out")
	traceT = flag.String("d", "2s", "trace duration, default: 10s")
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
	return d
}
