package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

type TNotifier struct {
	App *Application
}

func (sw *TNotifier) SetApp(app *Application) {
	sw.App = app
}

func (sw *TNotifier) Init() {
	app := sw.App
	window := sw.App.Window
	fontSize = 14
	w := 500
	h := (fontSize + 10) * 15
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
}

func (sw *TNotifier) Draw() {
	app := sw.App
	c := make(chan Line)
	go GetTime(c)
	var lines []Line
	// window := sw.App.Window
	T := app.Widget
	w := T.Geometry.Width
	h := T.Geometry.Height
	T.Padding.Left = 30
	var r sdl.Rect
	// var line string
	// var n string
	for {
		ret := <-c
		if ret.Content == "end" {
			break
		}
		lines = append(lines, ret)
		T.Clear()
		for i, line := range lines {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w), int32(h)}
			T.DrawColoredText(line.Content,
				&r, "foreground", "default",
				line.Rules,
			)
		}

		app.DrawMode()
		T.Renderer.Clear()
		T.Renderer.Present()
		sdl.Delay(5)
		sw.App.Window.UpdateSurface()
	}
}

func (sw *TNotifier) Run() int {
	// app := sw.App
	// window := sw.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				// sw.Draw()
			}
		case *sdl.QuitEvent:
			return 0
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
			key := sdl.GetScancodeName(t.Keysym.Scancode)
			// 	// TODO: add cursor to TW
			// 	if (key == "N" && t.Keysym.Mod == 64) || key == "Down" {
			// 		if SELECTED < len(CLIENTS)-1 {
			// 			SELECTED++
			// 		} else {
			// 			SELECTED = 0
			// 		}
			// 		sw.Draw()
			// 	}
			// 	if (key == "P" && t.Keysym.Mod == 64) || key == "Up" {
			// 		if SELECTED > 0 {
			// 			SELECTED--
			// 		} else {
			// 			SELECTED = len(CLIENTS) - 1
			// 		}
			// 		sw.Draw()
			// 	}
			// 	if key == "X" && t.Keysym.Mod == 64 {
			// 		wid := CLIENTS[SELECTED].WID
			// 		ewmh.CloseWindow(X, wid)
			// 		// time.Sleep(1 * time.Second)
			// 		sdl.Delay(1000)
			// 		CLIENTS = GetClients()
			// 		sw.Draw()
			// 	}
			// 	if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
			// 		wid := CLIENTS[SELECTED].WID
			// 		ewmh.ActiveWindowReq(X, wid)
			// 		return 0
			// 	}
			if key == "Escape" || key == "CapsLock" {
				return 0
			}
		}
	}
	return 1
}
