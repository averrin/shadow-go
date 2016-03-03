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
var fontSize int
var X *xgbutil.XUtil
var font *ttf.Font
var bold *ttf.Font
var SHADOW xproto.Window

type Geometry struct {
	Width  int32
	Height int32
}

type Padding struct {
	Top  int32
	Left int32
}

type TextWidget struct {
	Renderer   *sdl.Renderer
	Surface    *sdl.Surface
	Fonts      map[string]*ttf.Font
	Colors     map[string]sdl.Color
	BG         uint32
	LineHeight int
	Geometry
	Padding
}

var TW *TextWidget

func NewTextWidget(renderer *sdl.Renderer, surface *sdl.Surface) *TextWidget {
	widget := new(TextWidget)
	widget.Renderer = renderer
	widget.Surface = surface
	widget.Fonts = make(map[string]*ttf.Font)
	widget.Colors = make(map[string]sdl.Color)

	widget.Colors["foreground"] = sdl.Color{220, 220, 220, 1}
	widget.Colors["highlight"] = sdl.Color{255, 255, 255, 1}
	widget.Colors["accent"] = sdl.Color{35, 157, 200, 1}
	widget.Colors["gray"] = sdl.Color{100, 100, 100, 1}
	widget.Colors["orange"] = sdl.Color{242, 155, 23, 1}

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	font, _ = ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Regular.ttf"), fontSize)
	bold, _ = ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize)
	widget.Fonts["default"] = font
	widget.Fonts["bold"] = bold

	widget.BG = 0xff252525
	widget.LineHeight = fontSize + 6
	widget.Padding = Padding{10, 10}
	return widget
}

func (T *TextWidget) DrawText(text string, rect *sdl.Rect, colorName string, fontName string) {
	// log.Println(text)
	font, ok := T.Fonts[fontName]
	if !ok {
		font = T.Fonts["default"]
	}
	color, ok := T.Colors[colorName]
	if !ok {
		color = T.Colors["foreground"]
	}
	message, err := font.RenderUTF8_Blended(text, color)
	if err != nil {
		log.Fatal(err)
	}
	defer message.Free()
	srcRect := sdl.Rect{}
	message.GetClipRect(&srcRect)
	message.Blit(&srcRect, T.Surface, rect)
}

type HighlightRule struct {
	Start int
	Len   int
	Color string
}

func (T *TextWidget) DrawColoredText(text string, rect *sdl.Rect, colorName string, fontName string, rules []HighlightRule) {
	if len(rules) == 0 {
		T.DrawText(text, rect, colorName, fontName)
	} else {
		var token string
		for i := range rules {
			token = text[:rules[i].Start]
			// log.Println(token)
			var tw int
			if len(token) > 0 {
				T.DrawText(token, rect, colorName, fontName)
				tw, _, _ = T.Fonts[fontName].SizeUTF8(token)
				rect = &sdl.Rect{rect.X + int32(tw), rect.Y, rect.W - int32(tw), rect.H}
			}
			text = text[rules[i].Start:]
			token = text[:rules[i].Len]
			// log.Println(token)
			T.DrawText(token, rect, rules[i].Color, fontName)
			tw, _, _ = T.Fonts[fontName].SizeUTF8(token)
			rect = &sdl.Rect{rect.X + int32(tw), rect.Y, rect.W - int32(tw), rect.H}
			text = text[rules[i].Len:]
		}
		if len(token) > 0 {
			T.DrawText(text, rect, colorName, fontName)
		}
	}
}

func (T *TextWidget) Draw() {
	// log.Println(CLIENTS)
	w := T.Geometry.Width
	h := T.Geometry.Height
	rect := sdl.Rect{0, 0, int32(w), int32(h)}
	T.Surface.FillRect(&rect, T.BG)
	var r sdl.Rect
	for i, client := range CLIENTS {
		if SELECTED != i {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w), int32(h)}
			T.DrawColoredText(fmt.Sprintf("  %d [%d] %s", i, client.Desktop, client.Name),
				&r, "foreground", "default",
				[]HighlightRule{
					HighlightRule{5, 1, "orange"},
				},
			)
		} else {
			r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w) - T.Padding.Left, int32(h)}
			T.DrawColoredText(fmt.Sprintf("| %d [%d] %s", i, client.Desktop, client.Name),
				&r, "highlight", "bold",
				[]HighlightRule{
					HighlightRule{0, 1, "accent"},
				},
			)
		}
	}

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

func main() {
	SELECTED = 0
	lockPath := path.Join("/tmp", "shadow.lock")
	if fi, _ := os.Stat(lockPath); fi != nil {
		log.Println(fi)
		GetClients()
		ewmh.ActiveWindowReq(X, SHADOW)
		// file, err := os.Open(initPath)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer file.Close()
	} else {
		os.Create(lockPath)
		// conn, err := dbus.SessionBus()
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		// 	os.Exit(1)
		// }
		// var s []string
		// log.Println(conn.Object("org.kde.konsole", "/Sessions/4").Call(
		// 	"org.kde.konsole.Session.title", 2, "1").Store(&s))
		//
		// fmt.Println(s)
		os.Exit(run())
	}
}
