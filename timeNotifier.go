package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

//TNotifier mode
type TNotifier struct {
	App   *Application
	Alias string
}

//SetApp interface method
func (tn *TNotifier) SetApp(app *Application) {
	tn.App = app
	tn.Alias = "\uf017"
}

//GetAlias interface method
func (tn *TNotifier) GetAlias() string {
	return tn.Alias
}

//Init interface method
func (tn *TNotifier) Init() WidgetSettings {
	app := tn.App
	window := tn.App.Window
	fontSize = 14
	w := 500
	h := (fontSize + 10) * 13
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
	return WidgetSettings{fontSize, Geometry{int32(w), int32(h)}, Padding{10, 30, 10}}
}

//Draw interface method
func (tn *TNotifier) Draw() {
	c := make(chan Line)
	go GetTime(c)
	app := tn.App
	T := app.Widget
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

//Run interface method
func (tn *TNotifier) Run() int {
	// app := tn.App
	// window := tn.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				// tn.Draw()
			}
			if t.Event == sdl.WINDOWEVENT_FOCUS_LOST {
				// return 0
			}
		case *sdl.QuitEvent:
			return 0
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
			// key := sdl.GetScancodeName(t.Keysym.Scancode)
			if t.Keysym.Sym == sdl.K_ESCAPE || t.Keysym.Sym == sdl.K_CAPSLOCK {
				return 0
			}
		}
	}
	return 1
}
