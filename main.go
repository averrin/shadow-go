package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_ttf"
)

var fontSize int

// X is display manager
var X *xgbutil.XUtil

// SHADOW is WID of current opened shadow window
var SHADOW xproto.Window

// DrawMode is mode indicator drawing method
func (app *Application) DrawMode() {
	T := app.Widget
	w := T.Geometry.Width
	h := T.Geometry.Height
	m := app.Mode
	if app.Modes[app.Mode].GetAlias() != "" {
		m = app.Modes[app.Mode].GetAlias()
	}
	line := fmt.Sprintf("[%s]", m)
	lw, _, _ := T.Fonts["default"].SizeUTF8(line)
	r := sdl.Rect{
		X: int32(w) - T.Padding.Right - int32(lw),
		Y: int32(h-4) - int32(T.LineHeight),
		W: int32(lw),
		H: int32(T.LineHeight),
	}
	T.DrawColoredText(line,
		&r, "highlight", "default",
		[]HighlightRule{
			HighlightRule{1, len(m), ACCENT, "default"},
		},
	)
	T.Renderer.Present()
	T.App.Window.UpdateSurface()
}

// Application is main class
type Application struct {
	Mode   string
	Widget *TextWidget
	Modes  map[string]Mode
	Window *sdl.Window
}

// Mode is mode logic interface
type Mode interface {
	Init() WidgetSettings
	Draw()
	DispatchEvents(sdl.Event) int
	DispatchKeys(*sdl.KeyDownEvent) int
	SetApp(*Application)
	GetAlias() string
}

func (app *Application) run() int {
	sdl.Init(sdl.INIT_EVERYTHING)
	ttf.Init()

	settings := app.Modes[app.Mode].Init()
	defer app.Window.Destroy()
	renderer, err := sdl.CreateRenderer(app.Window, -1, sdl.RENDERER_ACCELERATED)
	surface, err := app.Window.GetSurface()
	if err != nil {
		panic(err)
	}
	app.Widget = NewTextWidget(app, renderer, surface, settings)
	app.Widget.FullClear()
	app.Modes[app.Mode].Draw()

	running := true
	for running {
		var event sdl.Event
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			ret := 1
			switch t := event.(type) {
			case *sdl.QuitEvent:
				ret = 0
			case *sdl.KeyDownEvent:
				fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
					t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
				key := sdl.GetScancodeName(t.Keysym.Scancode)
				//TODO: make mode switching more robust
				if t.Keysym.Sym == sdl.K_ESCAPE || t.Keysym.Sym == sdl.K_CAPSLOCK {
					ret = 0
				} else if key == "S" && t.Keysym.Mod == 64 && app.Mode != "tasks" {
					app.Mode = "tasks"
					app.Window.Destroy()
					app.run()
					ret = 0
				} else if key == "R" && t.Keysym.Mod == 64 && app.Mode != "runner" {
					app.Mode = "runner"
					app.Window.Destroy()
					app.run()
					ret = 0
				} else {
					ret = app.Modes[app.Mode].DispatchKeys(t)
				}
			default:
				ret = app.Modes[app.Mode].DispatchEvents(event)
			}
			if ret == 0 {
				running = false
			}
		}
	}
	os.Remove("/tmp/shadow.lock")
	return 0
}

func newApplication(mode string) *Application {
	app := new(Application)
	app.Mode = mode
	app.Modes = map[string]Mode{
		"tasks":  new(Switcher),
		"time":   new(TNotifier),
		"runner": new(Runner),
		"notify": new(Notifier),
	}
	for _, mode := range app.Modes {
		mode.SetApp(app)
	}
	return app
}

func main() {
	mode := flag.String("mode", "tasks", "shadow mode")
	TITLE = flag.String("title", "Notify", "Notification title")
	TEXT = flag.String("text", "Notify", "Notification text")
	flag.Parse()
	lockPath := path.Join("/tmp", "shadow.lock")
	if fi, _ := os.Stat(lockPath); fi != nil {
		log.Println(fi)
		GetClients()
		ewmh.ActiveWindowReq(X, SHADOW)
	} else {
		app := newApplication(*mode)
		os.Create(lockPath)
		os.Exit(app.run())
	}
}
