package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_ttf"
)

type Geometry struct {
	Width  int32
	Height int32
}

type Padding struct {
	Top   int32
	Left  int32
	Right int32
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

func NewTextWidget(renderer *sdl.Renderer, surface *sdl.Surface) *TextWidget {
	widget := new(TextWidget)
	widget.Renderer = renderer
	widget.Surface = surface
	widget.Fonts = make(map[string]*ttf.Font)
	widget.Colors = make(map[string]sdl.Color)

	widget.Colors["foreground"] = sdl.Color{200, 200, 200, 1}
	widget.Colors["highlight"] = sdl.Color{255, 255, 255, 1}
	widget.Colors["accent"] = sdl.Color{35, 157, 200, 1}
	widget.Colors["gray"] = sdl.Color{100, 100, 100, 1}
	widget.Colors["orange"] = sdl.Color{242, 155, 23, 1}
	widget.Colors["red"] = sdl.Color{215, 46, 46, 1}
	widget.Colors["green"] = sdl.Color{110, 173, 110, 1}

	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dir := filepath.Join(cwd, "fonts")
	font, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Regular.ttf"), fontSize)
	bold, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize)
	widget.Fonts["default"] = font
	widget.Fonts["bold"] = bold

	widget.BG = 0xff202020
	widget.LineHeight = fontSize + 6
	widget.Padding = Padding{10, 10, 10}
	return widget
}

func (T *TextWidget) StripLine(line string, fontname string) string {
	w := T.Geometry.Width
	lw, _, _ := T.Fonts[fontname].SizeUTF8(line)
	for int32(lw) > (int32(w) - T.Padding.Left*2) {
		line = strings.TrimRight(line[:len(line)-4], " -") + "…"
		lw, _, _ = T.Fonts[fontname].SizeUTF8(line)
	}
	return line
}

func (T *TextWidget) Clear() {
	w := T.Geometry.Width
	h := T.Geometry.Height
	rect := sdl.Rect{0, 0, int32(w), int32(h)}
	T.Surface.FillRect(&rect, T.BG)
}

func (T *TextWidget) DrawText(text string, rect *sdl.Rect, colorName string, fontName string) {
	if strings.TrimSpace(text) == "" {
		return
	}
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
			// log.Println(text, rules[i].Len)
			token = text[:rules[i].Len]
			// log.Println(token)
			T.DrawText(token, rect, rules[i].Color, fontName)
			tw, _, _ = T.Fonts[fontName].SizeUTF8(token)
			rect = &sdl.Rect{rect.X + int32(tw), rect.Y, rect.W - int32(tw), rect.H}
			text = text[rules[i].Len:]
			// log.Println(text)
		}
		if len(token) > 0 {
			T.DrawText(text, rect, colorName, fontName)
		}
	}
}
