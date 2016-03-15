package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type TNotifier struct {
	App   *Application
	Alias string
}

func (sw *TNotifier) SetApp(app *Application) {
	sw.App = app
	sw.Alias = "\uf017"
}

func (sw *TNotifier) GetAlias() string {
	return sw.Alias
}

func (sw *TNotifier) Init() WidgetSettings {
	app := sw.App
	window := sw.App.Window
	fontSize = 14
	w := 500
	h := (fontSize + 10) * 13
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
	return WidgetSettings{fontSize, Geometry{int32(w), int32(h)}, Padding{10, 10, 10}}
}

func (sw *TNotifier) Draw() {
	c := make(chan Line)
	go GetTime(c)
	app := sw.App
	T := app.Widget
	T.Padding.Left = 30
	T.Reset()
	T.App.DrawMode()
	for {
		ret := <-c
		if ret.Content == "end" {
			break
		}
		T.AddLine(ret)
	}
	go func() {
		time.Sleep(5 * time.Second)
		log.Println(os.Remove("/tmp/shadow.lock"))
		os.Exit(0)
	}()
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
			if t.Event == sdl.WINDOWEVENT_FOCUS_LOST {
				// return 0
			}
		case *sdl.QuitEvent:
			return 0
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
			key := sdl.GetScancodeName(t.Keysym.Scancode)
			if key == "Escape" || key == "CapsLock" {
				return 0
			}
		}
	}
	return 1
}
