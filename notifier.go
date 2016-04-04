package main

import (
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

//TEXT is notification text
var TEXT *string

//TITLE is notificaton title
var TITLE *string

//Notifier mode
type Notifier struct {
	App   *Application
	Alias string
}

//SetApp interface method
func (N *Notifier) SetApp(app *Application) {
	N.App = app
	// tn.Alias = "\uf017"
}

//GetAlias interface method
func (N *Notifier) GetAlias() string {
	return N.Alias
}

//Init interface method
func (N *Notifier) Init() WidgetSettings {
	app := N.App
	window := N.App.Window
	fontSize = 14
	w := 200
	h := (fontSize + 10) * 5
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	window.SetPosition(1720, 0)
	app.Window = window
	return WidgetSettings{fontSize, Geometry{int32(w), int32(h)}, Padding{10, 30, 10}}
}

//Draw interface method
func (N *Notifier) Draw() {
	// c := make(chan Line)
	// go GetTime(c)
	app := N.App
	T := app.Widget
	T.Reset()
	T.Cursor.Hidden = true
	T.SetContent([]Line{
		Line{*TITLE, []HighlightRule{HighlightRule{0, -1, "highlight", "header"}}},
		Line{"", []HighlightRule{}},
		Line{*TEXT, []HighlightRule{}},
	})
	T.App.DrawMode()
	// for {
	// 	ret := <-c
	// 	if ret.Content == "end" {
	// 		break
	// 	}
	// 	T.AddLine(ret)
	// }
	go func() {
		time.Sleep(5 * time.Second)
		log.Println(os.Remove("/tmp/shadow.lock"))
		os.Exit(0)
	}()
}

//DispatchEvents interface method
func (N *Notifier) DispatchEvents(event sdl.Event) int {
	return 1
}

//DispatchKeys interface method
func (N *Notifier) DispatchKeys(event *sdl.KeyDownEvent) int {
	return 1
}
