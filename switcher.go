package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/veandco/go-sdl2/sdl"
)

//Client struct
type Client struct {
	WID     xproto.Window
	Name    string
	Desktop uint
	Active  bool
}

//GetClients get windows list
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

//Switcher mode
type Switcher struct {
	App      *Application
	Selected int
	Clients  []Client
	Alias    string
}

//SetApp interface method
func (sw *Switcher) SetApp(app *Application) {
	sw.App = app
	sw.Alias = "\uf248"
}

//GetAlias interface method
func (sw *Switcher) GetAlias() string {
	return sw.Alias
}

//Init interface method
func (sw *Switcher) Init() WidgetSettings {
	sw.Selected = 0
	app := sw.App
	window := sw.App.Window
	sw.Clients = GetClients()
	fontSize = 14
	w := 500
	h := (fontSize + 10) * len(sw.Clients)
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
	return WidgetSettings{fontSize, Geometry{int32(w), int32(h)}, Padding{10, 10, 10}}
}

//Draw interface method
func (sw *Switcher) Draw() {
	app := sw.App
	T := app.Widget
	T.Reset()
	for i, client := range sw.Clients {
		T.AddLine(sw.getLine(i, client, sw.Selected == i))
	}
	T.App.DrawMode()
}

func (sw *Switcher) getLine(i int, client Client, focused bool) Line {
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
				HighlightRule{0, 1, ACCENT, "default"},
				HighlightRule{1, len(line) - 2, "highlight", "bold"},
			},
		}
	}
	return ret
}

//Run interface method
func (sw *Switcher) Run() int {
	app := sw.App
	T := app.Widget
	// window := sw.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				sw.Clients = GetClients()
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
				T.SetLine(sw.Selected, sw.getLine(sw.Selected, sw.Clients[sw.Selected], false))
				if sw.Selected < len(sw.Clients)-1 {
					sw.Selected++
				} else {
					sw.Selected = 0
				}
				T.SetLine(sw.Selected, sw.getLine(sw.Selected, sw.Clients[sw.Selected], true))
				return 1
			}
			if (key == "P" && t.Keysym.Mod == 64) || key == "Up" {
				T.SetLine(sw.Selected, sw.getLine(sw.Selected, sw.Clients[sw.Selected], false))
				if sw.Selected > 0 {
					sw.Selected--
				} else {
					sw.Selected = len(sw.Clients) - 1
				}
				T.SetLine(sw.Selected, sw.getLine(sw.Selected, sw.Clients[sw.Selected], true))
				return 1
			}
			if key == "X" && t.Keysym.Mod == 64 {
				wid := sw.Clients[sw.Selected].WID
				ewmh.CloseWindow(X, wid)
				sdl.Delay(500)
				sw.Clients = GetClients()
				sw.Draw()
				return 1
			}
			if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
				wid := sw.Clients[sw.Selected].WID
				ewmh.ActiveWindowReq(X, wid)
				return 0
			}
			if t.Keysym.Sym == sdl.K_ESCAPE || t.Keysym.Sym == sdl.K_CAPSLOCK {
				return 0
			}
		}
	}
	return 1
}
