package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/veandco/go-sdl2/sdl"
)

type Client struct {
	WID     xproto.Window
	Name    string
	Desktop uint
	Active  bool
}

func GetClients() []Client {
	clients := []Client{}
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	wids, err := ewmh.ClientListGet(X)
	if err != nil {
		log.Fatal(err)
	}
	a, _ := ewmh.ActiveWindowGet(X)
	for _, wid := range wids {
		name, err := ewmh.WmNameGet(X, wid)
		if name == "Shadow" {
			SHADOW = wid
			continue
		}
		if err != nil { // not a fatal error
			log.Println(err)
			name = ""
		}
		desk, _ := ewmh.WmDesktopGet(X, wid)
		clients = append(clients, Client{
			wid, name, desk, wid == a,
		})
	}
	return clients

}

type Switcher struct {
	App *Application
}

func (sw *Switcher) SetApp(app *Application) {
	sw.App = app
}

func (sw *Switcher) Init() {
	app := sw.App
	window := sw.App.Window
	CLIENTS = GetClients()
	fontSize = 14
	w := 500
	h := (fontSize + 10) * len(CLIENTS)
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
}

func (sw *Switcher) Draw() {
	app := sw.App
	// window := sw.App.Window
	T := app.Widget
	T.Clear()
	w := T.Geometry.Width
	h := T.Geometry.Height
	var r sdl.Rect
	var line string
	var n string
	for i, client := range CLIENTS {
		if i < 10 {
			n = fmt.Sprintf("%d", i)
		} else {
			n = " "
		}
		if SELECTED != i {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w), int32(h)}
			// TODO: move strip line into T method
			line = fmt.Sprintf("  %s [%d] %s", n, client.Desktop, client.Name)
			line = T.StripLine(line, "default")
			T.DrawColoredText(line,
				&r, "foreground", "default",
				[]HighlightRule{
					HighlightRule{5, 1, "orange"},
				},
			)
		} else {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w) - T.Padding.Left, int32(h)}
			line = fmt.Sprintf("| %s [%d] %s", n, client.Desktop, client.Name)
			line = T.StripLine(line, "bold")
			T.DrawColoredText(line,
				&r, "highlight", "bold",
				[]HighlightRule{
					HighlightRule{0, 1, "accent"},
				},
			)
		}
	}

	app.DrawMode()
	T.Renderer.Clear()
	T.Renderer.Present()
	sdl.Delay(5)
	sw.App.Window.UpdateSurface()
}

func (sw *Switcher) Run() int {
	// app := sw.App
	// window := sw.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				CLIENTS = GetClients()
				sw.Draw()
			}
		case *sdl.QuitEvent:
			return 0
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
			key := sdl.GetScancodeName(t.Keysym.Scancode)
			// TODO: add cursor to TW
			if (key == "N" && t.Keysym.Mod == 64) || key == "Down" {
				if SELECTED < len(CLIENTS)-1 {
					SELECTED++
				} else {
					SELECTED = 0
				}
				sw.Draw()
			}
			if (key == "P" && t.Keysym.Mod == 64) || key == "Up" {
				if SELECTED > 0 {
					SELECTED--
				} else {
					SELECTED = len(CLIENTS) - 1
				}
				sw.Draw()
			}
			if key == "X" && t.Keysym.Mod == 64 {
				wid := CLIENTS[SELECTED].WID
				ewmh.CloseWindow(X, wid)
				// time.Sleep(1 * time.Second)
				sdl.Delay(1000)
				CLIENTS = GetClients()
				sw.Draw()
			}
			if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
				wid := CLIENTS[SELECTED].WID
				ewmh.ActiveWindowReq(X, wid)
				return 0
			}
			if key == "Escape" || key == "CapsLock" {
				return 0
			}
		}
	}
	return 1
}
