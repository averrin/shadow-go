package main

import (
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

func (T *TextWidget) Draw() {
	// log.Println(CLIENTS)
	T.Clear()
	w := T.Geometry.Width
	h := T.Geometry.Height
	var r sdl.Rect
	var line string
	for i, client := range CLIENTS {
		if SELECTED != i {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w), int32(h)}
			// TODO: move strip line into T method
			line = fmt.Sprintf("  %d [%d] %s", i, client.Desktop, client.Name)
			lw, _, _ := T.Fonts["default"].SizeUTF8(line)
			for int32(lw) > (int32(w) - T.Padding.Left*2) {
				line = line[:len(line)-4] + "…"
				lw, _, _ = T.Fonts["default"].SizeUTF8(line)
			}
			T.DrawColoredText(line,
				&r, "foreground", "default",
				[]HighlightRule{
					HighlightRule{5, 1, "orange"},
				},
			)
		} else {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w) - T.Padding.Left, int32(h)}
			line = fmt.Sprintf("| %d [%d] %s", i, client.Desktop, client.Name)
			lw, _, _ := T.Fonts["bold"].SizeUTF8(line)
			for int32(lw) > (int32(w) - T.Padding.Left*2) {
				line = line[:len(line)-4] + "…"
				lw, _, _ = T.Fonts["default"].SizeUTF8(line)
			}
			T.DrawColoredText(line,
				&r, "highlight", "bold",
				[]HighlightRule{
					HighlightRule{0, 1, "accent"},
				},
			)
		}
	}

	// TODO: global mode
	line = "[tasks]"
	lw, _, _ := T.Fonts["default"].SizeUTF8(line)
	r = sdl.Rect{int32(w) - T.Padding.Left - int32(lw), int32(h-4) - int32(T.LineHeight), int32(lw), int32(T.LineHeight)}
	T.DrawColoredText(line,
		&r, "highlight", "default",
		[]HighlightRule{
			HighlightRule{1, 5, "accent"},
		},
	)
	T.Renderer.Clear()
	T.Renderer.Present()
}

func run() int {
	CLIENTS = GetClients()
	sdl.Init(sdl.INIT_EVERYTHING)
	ttf.Init()

	fontSize = 14
	w := 500
	h := (fontSize + 10) * len(CLIENTS)
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	TW = NewTextWidget(renderer, surface)
	TW.Geometry = Geometry{int32(w), int32(h)}
	TW.Draw()

	sdl.Delay(5)
	window.UpdateSurface()

	var event sdl.Event
	running := true
	for running {
		// TODO: move handlers to mode obj
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.WindowEvent:
				if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
					CLIENTS = GetClients()
					TW.Draw()
					window.UpdateSurface()
				}
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyUpEvent:
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
					TW.Draw()
					window.UpdateSurface()
				}
				if (key == "P" && t.Keysym.Mod == 64) || key == "Up" {
					if SELECTED > 0 {
						SELECTED--
					} else {
						SELECTED = len(CLIENTS) - 1
					}
					TW.Draw()
					window.UpdateSurface()
				}
				if key == "X" && t.Keysym.Mod == 64 {
					wid := CLIENTS[SELECTED].WID
					ewmh.CloseWindow(X, wid)
					// time.Sleep(1 * time.Second)
					sdl.Delay(1000)
					CLIENTS = GetClients()
					TW.Draw()
					window.UpdateSurface()
				}
				if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
					wid := CLIENTS[SELECTED].WID
					ewmh.ActiveWindowReq(X, wid)
					running = false
				}
				if key == "Escape" || key == "CapsLock" {
					running = false
				}
			}
		}
	}
	log.Println(os.Remove("/tmp/shadow.lock"))
	return 0
}

func main() {
	SELECTED = 0
	lockPath := path.Join("/tmp", "shadow.lock")
	if fi, _ := os.Stat(lockPath); fi != nil {
		log.Println(fi)
		GetClients()
		ewmh.ActiveWindowReq(X, SHADOW)
	} else {
		os.Create(lockPath)
		os.Exit(run())
	}
}
