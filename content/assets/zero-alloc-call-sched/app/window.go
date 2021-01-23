// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package app

import (
	"x/mainthread"
	"x/thread"

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
	th  *thread.Thread
}

// NewWindow constructs a new graphical window.
func NewWindow() (*Win, error) {
	var (
		w   = &Win{th: thread.New()}
		err error
	)
	mainthread.Call(func() {
		w.win, err = glfw.CreateWindow(640, 480, "", nil, nil)
		if err != nil {
			return
		}
	})

	// This function can be called from any thread.
	w.th.Call(w.win.MakeContextCurrent)
	return w, nil
}

// Run runs the given window and blocks until it is destroied.
func (w *Win) Run() {
	for !w.closed() {
		w.update()
	}
	w.destroy()
}

// Stop stops the given window.
func (w *Win) Stop() {
	// This function can be called from any threads.
	w.th.Call(func() { w.win.SetShouldClose(true) })
}

// closed asserts if the given window is closed.
func (w *Win) closed() bool {
	// This function can be called from any thread.
	var stop bool
	w.th.Call(func() { stop = w.win.ShouldClose() })
	return stop
}

// Update updates the frame buffer of the given window.
func (w *Win) update() {
	mainthread.Call(func() {
		w.win.SwapBuffers()
		// This function must be called from the main thread.
		glfw.WaitEventsTimeout(1.0 / 30)
	})
}

// destroy destructs the given window.
func (w *Win) destroy() {
	// This function must be called from the mainthread.
	mainthread.Call(w.win.Destroy)
	w.th.Terminate()
}
