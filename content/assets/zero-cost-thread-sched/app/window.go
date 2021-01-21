// Copyright (c) 2021 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

package app

import (
	"x/mainthread"

	"github.com/go-gl/glfw/v3.3/glfw"
)

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
		err = glfw.Init()
		if err != nil {
			return
		}

		w.win, err = glfw.CreateWindow(640, 480, "win", nil, nil)
		if err != nil {
			return
		}
	})
	return w, nil
}

// Terminate terminates the entire application.
func Terminate() {
	mainthread.Call(func() {
		glfw.Terminate()
	})
}

// Closed asserts if the given window is closed.
func (w *Win) Closed() (stop bool) {
	return mainthread.CallV(func() interface{} {
		return w.win.ShouldClose()
	}).(bool)
}

// Update updates the frame buffer of the given window.
func (w *Win) Update() {
	mainthread.Call(func() {
		w.win.SwapBuffers()
		// glfw.WaitEventsTimeout(1.0 / 30)
		glfw.PollEvents()
	})
}

// Stop stops the given window.
func (w *Win) Stop() {
	mainthread.Call(func() {
		w.win.SetShouldClose(true)
	})
}
