// Copyright 2021 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.
//
// Written by Changkun Ou <changkun.de>

// WARNING: This example contains a deadlock.
package main

import (
	"fmt"
	"math/rand"
	"time"
)

type ResizeEvent struct {
	width, height int
}

type renderProfile struct {
	id     int
	width  int
	height int
}

// Draw executes a draw call by the given render profile
func (p *renderProfile) Draw() interface{} {
	return fmt.Sprintf("draw-%d-%dx%d", p.id, p.width, p.height)
}

func main() {
	// draw is a channel for receiving finished draw calls.
	draw := make(chan interface{})

	// Solution 2 (step 1):
	// drawIn, drawOut := MakeChan()

	// change is a channel to receive notification of the change of
	// rendering settings.
	change := make(chan ResizeEvent)

	// Rendering Thread
	//
	// Sending drawcalls to the event thread in order to draw pictures.
	// The thread sends darwcalls to the draw channel, using the same
	// rendering setting id. If there is a change of rendering setting,
	// the event thread notifies the rendering setting change, and here
	// increases the rendering setting id.
	go func() {
		p := &renderProfile{id: 0, width: 800, height: 500}
		for {
			select {
			case size := <-change:
				// Modify rendering profile.
				p.id++
				p.width = size.width
				p.height = size.height
			default:
				draw <- p.Draw()
				// Solution 1:
				// select {
				// case draw <- p.Draw():
				// default:
				// }
				// Solution 2 (step 2):
				// drawIn <- p.Draw()
			}
		}
	}()

	// Event Thread
	//
	// Process events every 100 ms. Otherwise, process drawcall request
	// upon-avaliable.
	event := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case id := <-draw:
			// Solution 2 (step 3):
			// case id := <-drawOut:
			println(id)
		case <-event.C:
			// Notify the rendering thread there is a change regarding
			// rendering settings. We simulate a random size at every
			// event processing loop.
			change <- ResizeEvent{
				width:  int(rand.Float64() * 100),
				height: int(rand.Float64() * 100),
			}
		}
	}
}
