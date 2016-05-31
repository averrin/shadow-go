package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/veandco/go-sdl2/sdl"
)

var icons map[string][]string

//Client struct
type Client struct {
	WID     xproto.Window
	Name    string
	Desktop uint
	Active  bool
	Class   string
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
		class, _ := icccm.WmClassGet(X, wid)
		clients = append(clients, Client{
			wid, name, desk, wid == a, class.Class,
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
	h := (fontSize + 12) * len(sw.Clients)
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window

	icons = map[string][]string{}
	icons["vivaldi-snapshot"] = []string{"\uf27d", "red"}
	icons["Skype"] = []string{"\uf17e", ACCENT}
	icons["konsole"] = []string{"\uf120", "default"}
	icons["Thunderbird"] = []string{"\uf0e0", "default"}
	icons["Atom"] = []string{"\ue7ba", "default"}
	icons["yakyak"] = []string{"\uf086", GREEN}

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
	// return Line{client.Name, []HighlightRule{}}
	app := sw.App
	T := app.Widget
	var line string
	var n string
	var d string
	var ret Line
	if i < 10 {
		n = fmt.Sprintf("%d", i)
	} else {
		n = " "
	}
	t := client.Name
	icon, hasIcon := icons[client.Class]
	var rules []HighlightRule
	if !focused {
		if hasIcon {
			rules = []HighlightRule{
				HighlightRule{4, 3, icon[1], "default"},
			}
			line = fmt.Sprintf("  %s %s  %s", n, icon[0], t)
		} else {
			rules = []HighlightRule{
				HighlightRule{5, 1, "orange", "default"},
			}
			d = strconv.Itoa(int(client.Desktop))
			line = fmt.Sprintf("  %s [%s] %s", n, d, t)
		}
		line = T.StripLine(line, "default")
		ret = Line{
			line,
			rules,
		}
	} else {
		if hasIcon {
			line = fmt.Sprintf("  %s %s  %s", n, icon[0], t)
		} else {
			d = strconv.Itoa(int(client.Desktop))
			line = fmt.Sprintf("| %s [%s] %s", n, d, t)
		}
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

//DispatchKeys interface method
func (sw *Switcher) DispatchKeys(t *sdl.KeyDownEvent) int {
	app := sw.App
	T := app.Widget
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
	if strings.Index("0123456789", key) > -1 {
		i, err := strconv.Atoi(key)
		if err == nil && len(sw.Clients) > i {
			sw.Selected = i
			if t.Keysym.Mod == 64 {
				wid := sw.Clients[sw.Selected].WID
				ewmh.ActiveWindowReq(X, wid)
				return 0
			}
			sw.Draw()
			return 1
		}
	}
	if t.Keysym.Sym == sdl.K_ESCAPE || t.Keysym.Sym == sdl.K_CAPSLOCK {
		return 0
	}
	return 1
}

//DispatchEvents interface method
func (sw *Switcher) DispatchEvents(event sdl.Event) int {
	switch t := event.(type) {
	case *sdl.WindowEvent:
		if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
			sw.Clients = GetClients()
			sw.Draw()
		}
		if t.Event == sdl.WINDOWEVENT_FOCUS_LOST {
			ewmh.ActiveWindowReq(X, SHADOW)
		}
	}
	return 1
}
