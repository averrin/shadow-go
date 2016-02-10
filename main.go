package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_ttf"
)

var SELECTED int
var CLIENTS []Client
var w int
var h int
var fontSize int
var X *xgbutil.XUtil
var font *ttf.Font
var bold *ttf.Font

func DrawText(parent *sdl.Surface, text string, rect *sdl.Rect, color sdl.Color, font *ttf.Font) {
	log.Println(text)
	message, err := font.RenderUTF8_Blended(text, color)
	if err != nil {
		log.Fatal(err)
	}
	defer message.Free()
	srcRect := sdl.Rect{}
	message.GetClipRect(&srcRect)
	message.Blit(&srcRect, parent, rect)
}

func Draw(renderer *sdl.Renderer, surface *sdl.Surface) {
	log.Println(CLIENTS)
	rect := sdl.Rect{0, 0, int32(w), int32(h)}
	surface.FillRect(&rect, 0xff252525)
	f := font
	for i, client := range CLIENTS {
		r := sdl.Rect{10, int32(10 + (i * (fontSize + 6))), int32(w), int32(h)}
		if SELECTED == i {
			f = bold
		} else {
			f = font
		}
		if SELECTED != i {
			DrawText(surface, fmt.Sprintf("  %d [%d] %s", i, client.Desktop, client.Name), &r, sdl.Color{200, 200, 200, 1}, f)
		} else {
			DrawText(surface, fmt.Sprintf("| %d [%d] %s", i, client.Desktop, client.Name), &r, sdl.Color{255, 255, 255, 1}, f)
		}
	}

	// tx, err := renderer.CreateTextureFromSurface(message)
	// log.Println(tx, err)
	renderer.Clear()
	renderer.Present()
}

func run() int {
	CLIENTS = GetClients()
	sdl.Init(sdl.INIT_EVERYTHING)
	ttf.Init()
	w = 500
	fontSize = 14
	h = (fontSize + 10) * len(CLIENTS)

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	log.Println(dir)
	font, _ = ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Regular.ttf"), fontSize)
	bold, _ = ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize)
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

	Draw(renderer, surface)

	sdl.Delay(5)
	window.UpdateSurface()

	var event sdl.Event
	running := true
	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyUpEvent:
				fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%s\tmodifiers:%d\tstate:%d\trepeat:%d\n",
					t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat)
				key := sdl.GetScancodeName(t.Keysym.Scancode)
				if (key == "N" && t.Keysym.Mod == 64) || key == "Down" {
					if SELECTED < len(CLIENTS)-1 {
						SELECTED++
					} else {
						SELECTED = 0
					}
					Draw(renderer, surface)
					window.UpdateSurface()
				}
				if (key == "P" && t.Keysym.Mod == 64) || key == "Up" {
					if SELECTED > 0 {
						SELECTED--
					} else {
						SELECTED = len(CLIENTS) - 1
					}
					Draw(renderer, surface)
					window.UpdateSurface()
				}
				if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
					wid := CLIENTS[SELECTED].WID
					ewmh.ActiveWindowReq(X, wid)
					// log.Println(ewmh.ActiveWindowSet(X, wid))
					running = false
				}
				if key == "Escape" || key == "CapsLock" {
					running = false
				}
			}
		}
	}
	return 0
}

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

func main() {
	SELECTED = 0
	os.Exit(run())
}
