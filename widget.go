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

type Line struct {
	Content string
	Rules   []HighlightRule
}

type Cursor struct {
	Row    int
	Column int
}

type TextWidget struct {
	App        *Application
	Renderer   *sdl.Renderer
	Surface    *sdl.Surface
	Fonts      map[string]*ttf.Font
	Colors     map[string]sdl.Color
	BG         uint32
	Content    []Line
	LineHeight int
	Geometry
	Padding
	Cursor
}

func NewTextWidget(app *Application, renderer *sdl.Renderer, surface *sdl.Surface) *TextWidget {
	widget := new(TextWidget)
	widget.App = app
	widget.Renderer = renderer
	widget.Surface = surface
	widget.Fonts = make(map[string]*ttf.Font)
	widget.Colors = make(map[string]sdl.Color)
	widget.Content = make([]Line, 0)
	widget.Cursor = Cursor{0, 0}

	widget.Colors["foreground"] = sdl.Color{200, 200, 200, 1}
	widget.Colors["highlight"] = sdl.Color{255, 255, 255, 1}
	widget.Colors["accent"] = sdl.Color{129, 162, 190, 1}
	widget.Colors["gray"] = sdl.Color{100, 100, 100, 1}
	widget.Colors["orange"] = sdl.Color{240, 198, 116, 1}
	widget.Colors["red"] = sdl.Color{215, 46, 46, 1}
	widget.Colors["green"] = sdl.Color{110, 173, 110, 1}

	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dir := filepath.Join(cwd, "fonts")
	font, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Regular.ttf"), fontSize)
	bold, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize)
	header, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize+4)
	bigger, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize+2)
	widget.Fonts["default"] = font
	widget.Fonts["bold"] = bold
	widget.Fonts["header"] = header
	widget.Fonts["bigger"] = bigger

	widget.BG = 0xff242424
	widget.LineHeight = fontSize + 6
	widget.Padding = Padding{10, 10, 10}
	return widget
}

func (T *TextWidget) SetContent(content []Line) {
	T.Content = content
	T.Update()
}

func (T *TextWidget) SetLine(index int, line Line) {
	w := T.Geometry.Width
	// h := T.Geometry.Height
	r := sdl.Rect{T.Padding.Left, T.Padding.Top + int32(index*T.LineHeight), int32(w), int32(T.LineHeight)}
	T.Content[index] = line
	T.Surface.FillRect(&r, T.BG)
	T.DrawColoredText(line.Content,
		&r, "foreground", "default",
		line.Rules,
	)
	// T.Renderer.Clear()
	T.Renderer.Present()
	// sdl.Delay(1)
	T.App.Window.UpdateSurface()
}

func (T *TextWidget) AddLine(line Line) {
	w := T.Geometry.Width
	h := T.Geometry.Height
	r := sdl.Rect{T.Padding.Left, T.Padding.Top + int32(len(T.Content)*T.LineHeight), int32(w), int32(h)}
	T.Content = append(T.Content, line)
	T.DrawColoredText(line.Content,
		&r, "foreground", "default",
		line.Rules,
	)
	// T.Renderer.Clear()
	T.Renderer.Present()
	// sdl.Delay(1)
	T.App.Window.UpdateSurface()
}

func (T *TextWidget) ClearContent() {
	T.Content = []Line{}
}

func (T *TextWidget) Reset() {
	T.ClearContent()
	T.Clear()
}

func (T *TextWidget) Update() {
	w := T.Geometry.Width
	// h := T.Geometry.Height
	var r sdl.Rect
	T.Clear()
	for i, line := range T.Content {
		r = sdl.Rect{T.Padding.Left, T.Padding.Top + int32(i*T.LineHeight), int32(w), int32(T.LineHeight)}
		T.DrawColoredText(line.Content,
			&r, "foreground", "default",
			line.Rules,
		)
	}

	T.App.DrawMode()
	T.drawCursor()
	T.Renderer.Clear()
	T.Renderer.Present()
	sdl.Delay(5)
	T.App.Window.UpdateSurface()
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
	if fontName != "default" {
		_, h, _ := T.Fonts["default"].SizeUTF8("A")
		_, h2, _ := font.SizeUTF8("A")
		rect.Y -= int32((h2 - h) / 2)
	}
	message.Blit(&srcRect, T.Surface, rect)
}

type HighlightRule struct {
	Start int
	Len   int
	Color string
	Font  string
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
			l := rules[i].Len
			if l > len(text) || l == -1 {
				l = len(text)
			}
			token = text[:l]
			// log.Println(token)
			T.DrawText(token, rect, rules[i].Color, rules[i].Font)
			tw, _, _ = T.Fonts[fontName].SizeUTF8(token)
			rect = &sdl.Rect{rect.X + int32(tw), rect.Y, rect.W - int32(tw), rect.H}
			text = text[l:]
			// log.Println(text)
		}
		if len(token) > 0 {
			T.DrawText(text, rect, colorName, fontName)
		}
	}
}

func (T *TextWidget) MoveCursor(r int, c int) (int, int) {
	T.Cursor.Row = r
	T.Cursor.Column = c
	T.drawCursor()
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) MoveCursorLeft() (int, int) {
	T.MoveCursor(T.Cursor.Row, T.Cursor.Column-1)
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) MoveCursorRight() (int, int) {
	T.MoveCursor(T.Cursor.Row, T.Cursor.Column+1)
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) MoveCursorUp() (int, int) {
	T.MoveCursor(T.Cursor.Row-1, T.Cursor.Column)
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) MoveCursorDown() (int, int) {
	T.MoveCursor(T.Cursor.Row+1, T.Cursor.Column)
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) addString(s string) (int, int) {
	line := T.Content[T.Cursor.Row]
	i := T.Cursor.Column
	line.Content = line.Content[:i] + s + line.Content[i:]
	T.SetLine(0, line)
	T.MoveCursor(T.Cursor.Row, T.Cursor.Column+len(s))
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) removeString(n int) (int, int) {
	if T.Cursor.Column > 0 {
		line := T.Content[T.Cursor.Row]
		i := T.Cursor.Column
		line.Content = line.Content[:i-n] + line.Content[i:]
		T.SetLine(0, line)
		T.MoveCursor(T.Cursor.Row, T.Cursor.Column-n)
	}
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) removeWord() (int, int) {
	log.Println(T.Cursor.Column)
	if T.Cursor.Column > 0 {
		index := T.Cursor.Row
		line := T.Content[index].Content[:T.Cursor.Column-1]
		n := strings.LastIndexAny(line, " -;") + 1
		return T.removeString(T.Cursor.Column - n)
	}
	return T.Cursor.Row, T.Cursor.Column
}

func (T *TextWidget) drawCursor() {
	index := T.Cursor.Row
	var lw int
	if T.Cursor.Column > 0 {
		line := T.Content[index].Content[:T.Cursor.Column-1]
		lw, _, _ = T.Fonts["default"].SizeUTF8(line)
	} else {
		lw = -6
	}
	r := sdl.Rect{T.Padding.Left + int32(lw) + 6, T.Padding.Top + int32(index*T.LineHeight), int32(5), int32(T.LineHeight)}
	T.DrawColoredText("|", &r, "accent", "default", []HighlightRule{})
	T.Renderer.Present()
	T.App.Window.UpdateSurface()
}

func (T *TextWidget) SetRules(index int, rules []HighlightRule) {
	line := T.Content[index]
	line.Rules = rules
	T.SetLine(index, line)
}
