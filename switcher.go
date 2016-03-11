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
	T := app.Widget
	T.Reset()
	for i, client := range CLIENTS {
		T.AddLine(sw.GetLine(i, client, SELECTED == i))
	}
	T.App.DrawMode()
}

func (sw *Switcher) GetLine(i int, client Client, focused bool) Line {
	app := sw.App
	T := app.Widget
	var line string
	var n string
	var ret Line
	if i < 10 {
		n = fmt.Sprintf("%d", i)
	} else {
		n = " "
	}
	if !focused {
		line = fmt.Sprintf("  %s [%d] %s", n, client.Desktop, client.Name)
		line = T.StripLine(line, "default")
		ret = Line{
			line,
			[]HighlightRule{
				HighlightRule{5, 1, "orange", "default"},
			},
		}
	} else {
		line = fmt.Sprintf("| %s [%d] %s", n, client.Desktop, client.Name)
		line = T.StripLine(line, "bold")
		ret = Line{line,
			[]HighlightRule{
				HighlightRule{0, 1, "accent", "default"},
				HighlightRule{1, len(line) - 2, "highlight", "bold"},
			},
		}
	}
	return ret
}

func (sw *Switcher) Run() int {
	app := sw.App
	T := app.Widget
	// window := sw.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				CLIENTS = GetClients()
				sw.Draw()
			}
			if t.Event == sdl.WINDOWEVENT_FOCUS_LOST {
				ewmh.ActiveWindowReq(X, SHADOW)
			}
		case *sdl.QuitEvent:
			return 0
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
				t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
			key := sdl.GetScancodeName(t.Keysym.Scancode)
			if (key == "N" && t.Keysym.Mod == 64) || key == "Down" {
				T.SetLine(SELECTED, sw.GetLine(SELECTED, CLIENTS[SELECTED], false))
				if SELECTED < len(CLIENTS)-1 {
					SELECTED++
				} else {
					SELECTED = 0
				}
				T.SetLine(SELECTED, sw.GetLine(SELECTED, CLIENTS[SELECTED], true))
			}
			if (key == "P" && t.Keysym.Mod == 64) || key == "Up" {
				T.SetLine(SELECTED, sw.GetLine(SELECTED, CLIENTS[SELECTED], false))
				if SELECTED > 0 {
					SELECTED--
				} else {
					SELECTED = len(CLIENTS) - 1
				}
				T.SetLine(SELECTED, sw.GetLine(SELECTED, CLIENTS[SELECTED], true))
			}
			if key == "X" && t.Keysym.Mod == 64 {
				wid := CLIENTS[SELECTED].WID
				ewmh.CloseWindow(X, wid)
				sdl.Delay(500)
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
