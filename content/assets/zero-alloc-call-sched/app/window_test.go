package app

import (
	"testing"
	"x/mainthread"
	"x/thread"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// This test will crash!
func TestApp(t *testing.T) {
	mainthread.Init(func() {
		w, _ := NewWindow()
		defer Terminate()

		renderThread := thread.New()
		renderThread.Call(func() {
			// Initialize gl from a different thread
			gl.Init()
			// gl.MakeContextCurrent()
			gl.GoStr(gl.GetString(gl.VERSION))
		})
		mainthread.Call(func() {
			// Initialize gl from the main thread can
			// prevent crash from the GetVersion crash.
			// gl.Init()
			gl.GoStr(gl.GetString(gl.VERSION))
		})
		w.Stop()
	})
}
