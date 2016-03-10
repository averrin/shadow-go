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

var SELECTED int
var CLIENTS []Client
var fontSize int
var X *xgbutil.XUtil

var SHADOW xproto.Window

func (app *Application) DrawMode() {
	T := app.Widget
	w := T.Geometry.Width
	h := T.Geometry.Height
	line := fmt.Sprintf("[%s]", app.Mode)
	lw, _, _ := T.Fonts["default"].SizeUTF8(line)
	r := sdl.Rect{int32(w) - T.Padding.Left - int32(lw), int32(h-4) - int32(T.LineHeight), int32(lw), int32(T.LineHeight)}
	T.DrawColoredText(line,
		&r, "highlight", "default",
		[]HighlightRule{
			HighlightRule{1, len(app.Mode), "accent"},
		},
	)
}

type Application struct {
	Mode   string
	Widget *TextWidget
	Modes  map[string]Mode
	Window *sdl.Window
}

type Mode interface {
	Init()
	Draw()
	Run() int
	SetApp(*Application)
}

func (app *Application) run() int {
	sdl.Init(sdl.INIT_EVERYTHING)
	ttf.Init()

	app.Modes[app.Mode].Init()
	defer app.Window.Destroy()
	renderer, err := sdl.CreateRenderer(app.Window, -1, sdl.RENDERER_ACCELERATED)
	surface, err := app.Window.GetSurface()
	if err != nil {
		panic(err)
	}
	app.Widget = NewTextWidget(renderer, surface)
	w, h := app.Window.GetSize()
	app.Widget.Geometry = Geometry{int32(w), int32(h)}
	app.Modes[app.Mode].Draw()

	running := true
	for running {
		ret := app.Modes[app.Mode].Run()
		if ret == 0 {
			running = false
		}
	}
	log.Println(os.Remove("/tmp/shadow.lock"))
	return 0
}

func NewApplication(mode string) *Application {
	app := new(Application)
	app.Mode = mode
	app.Modes = make(map[string]Mode)
	sw := new(Switcher)
	sw.SetApp(app)
	app.Modes["tasks"] = sw

	tn := new(TNotifier)
	tn.SetApp(app)
	app.Modes["time"] = tn
	return app
}

func main() {
	mode := flag.String("mode", "tasks", "shadow mode")
	flag.Parse()
	SELECTED = 0
	lockPath := path.Join("/tmp", "shadow.lock")
	if fi, _ := os.Stat(lockPath); fi != nil {
		log.Println(fi)
		GetClients()
		ewmh.ActiveWindowReq(X, SHADOW)
	} else {
		app := NewApplication(*mode)
		os.Create(lockPath)
		os.Exit(app.run())
	}
}
