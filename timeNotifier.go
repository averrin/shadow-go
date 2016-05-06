package main

import (
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
		if app.Mode == "time" {
			log.Println(os.Remove("/tmp/shadow.lock"))
			os.Exit(0)
		}
	}()
}

//DispatchEvents interface method
func (tn *TNotifier) DispatchEvents(event sdl.Event) int {
	return 1
}

//DispatchKeys interface method
func (tn *TNotifier) DispatchKeys(event *sdl.KeyDownEvent) int {
	return 1
}
