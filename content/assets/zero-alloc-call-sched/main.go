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

	err = app.Init()
	if err != nil {
		panic(err)
	}
	defer app.Terminate()

	w1, err := app.NewWindow()
	if err != nil {
		panic(err)
	}
	w2, err := app.NewWindow()
	if err != nil {
		panic(err)
	}

	done := make(chan struct{}, 3)
	go func() {
		defer func() { done <- struct{}{} }()
		f, _ := os.Create(*traceF)
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
		time.Sleep(d)
		w1.Stop()
		time.Sleep(d)
		w2.Stop()
	}()

	go func() {
		w1.Run()
		done <- struct{}{}
	}()
	go func() {
		w2.Run()
		done <- struct{}{}
	}()
	<-done
	<-done
	<-done
}
