// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package app

import (
	mainthread "x/mainthread-opt2"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// Init initializes an app environment.
func Init() (err error) {
	mainthread.Call(func() { err = glfw.Init() })
	return
}

// Terminate terminates the entire application.
func Terminate() {
	mainthread.Call(glfw.Terminate)
}

// Win is a window.
type Win struct {
	win *glfw.Window
}

// NewWindow constructs a new graphical window.
func NewWindow() (*Win, error) {
	var (
		w   = &Win{}
		err error
	)
	mainthread.Call(func() {
		w.win, err = glfw.CreateWindow(640, 480, "golang.design/research", nil, nil)
		if err != nil {
			return
		}
	})
	if err != nil {
		return nil, err
	}

	w.win.MakeContextCurrent()
	return w, nil
}

// Run runs the given window and blocks until it is destroied.
func (w *Win) Run() {
	for !w.win.ShouldClose() {
		w.win.SwapBuffers()
		mainthread.Call(func() {
			// This function must be called from the main thread.
			glfw.WaitEventsTimeout(1.0 / 30)
		})
	}
	// This function must be called from the mainthread.
	mainthread.Call(w.win.Destroy)
}

// Stop stops and closes the given window.
func (w *Win) Stop() {
	w.win.SetShouldClose(true)
}
