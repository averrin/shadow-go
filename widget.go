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

// Colors
var (
	ACCENT = "accent"
	GREEN  = "green"
)

// Geometry is widget size
type Geometry struct {
	Width  int32
	Height int32
}

// Padding is paddings
type Padding struct {
	Top   int32
	Left  int32
	Right int32
}

// Line is line of text
type Line struct {
	Content string
	Rules   []HighlightRule
}

// Cursor is cursor cords
type Cursor struct {
	Row    int
	Column int
	Hidden bool
}

// TextWidget is text canvas
type TextWidget struct {
	App        *Application
	Renderer   *sdl.Renderer
	Surface    *sdl.Surface
	FontSize   int
	Fonts      map[string]*ttf.Font
	Colors     map[string]sdl.Color
	BG         uint32
	Content    []Line
	LineHeight int
	Geometry
	Padding
	Cursor
}

// WidgetSettings is init settings struct
type WidgetSettings struct {
	FontSize int
	Geometry
	Padding
}

// NewTextWidget is constructor
func NewTextWidget(app *Application, renderer *sdl.Renderer, surface *sdl.Surface, settings WidgetSettings) *TextWidget {
	widget := new(TextWidget)
	widget.App = app
	widget.Renderer = renderer
	widget.Surface = surface
	widget.Fonts = make(map[string]*ttf.Font)
	widget.Colors = make(map[string]sdl.Color)
	widget.Content = make([]Line, 0)
	widget.Cursor = Cursor{0, 0, false}
	widget.FontSize = settings.FontSize

	widget.Colors["foreground"] = sdl.Color{
		R: 200,
		G: 200,
		B: 200,
		A: 1,
	}
	widget.Colors["highlight"] = sdl.Color{
		R: 255,
		G: 255,
		B: 255,
		A: 1,
	}
	widget.Colors[ACCENT] = sdl.Color{
		R: 129,
		G: 162,
		B: 190,
		A: 1,
	}
	widget.Colors["gray"] = sdl.Color{
		R: 130,
		G: 130,
		B: 130,
		A: 1,
	}
	widget.Colors["orange"] = sdl.Color{
		R: 240,
		G: 198,
		B: 116,
		A: 1,
	}
	widget.Colors["red"] = sdl.Color{
		R: 215,
		G: 46,
		B: 46,
		A: 1,
	}
	widget.Colors[GREEN] = sdl.Color{
		R: 110,
		G: 173,
		B: 110,
		A: 1,
	}

	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	dir := filepath.Join(cwd, "fonts")
	// font, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Regular.ttf"), fontSize)
	// bold, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize)
	// header, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize+4)
	// bigger, _ := ttf.OpenFont(path.Join(dir, "FantasqueSansMono-Bold.ttf"), fontSize+2)
	font, _ := ttf.OpenFont(path.Join(dir, "Fantasque Sans Mono Regular Nerd Font Plus Font Awesome Plus Octicons Plus Pomicons.ttf"), widget.FontSize)
	bold, _ := ttf.OpenFont(path.Join(dir, "Fantasque Sans Mono Bold Nerd Font Plus Font Awesome Plus Octicons Plus Pomicons.ttf"), widget.FontSize)
	header, _ := ttf.OpenFont(path.Join(dir, "Fantasque Sans Mono Bold Nerd Font Plus Font Awesome Plus Octicons Plus Pomicons.ttf"), widget.FontSize+4)
	bigger, _ := ttf.OpenFont(path.Join(dir, "Fantasque Sans Mono Bold Nerd Font Plus Font Awesome Plus Octicons Plus Pomicons.ttf"), widget.FontSize+2)
	widget.Fonts["default"] = font
	widget.Fonts["bold"] = bold
	widget.Fonts["header"] = header
	widget.Fonts["bigger"] = bigger

	widget.BG = 0xff242424
	widget.LineHeight = widget.FontSize + 6
	widget.Geometry = settings.Geometry
	widget.Padding = settings.Padding
	return widget
}

// SetContent is
func (T *TextWidget) SetContent(content []Line) {
	T.Content = content
	T.Update()
}

// SetLine is
func (T *TextWidget) SetLine(index int, line Line) {
	w := T.Geometry.Width
	// h := T.Geometry.Height
	r := sdl.Rect{
		X: T.Padding.Left,
		Y: T.Padding.Top + int32(index*T.LineHeight),
		W: int32(w),
		H: int32(T.LineHeight),
	}
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

// AddLine is
func (T *TextWidget) AddLine(line Line) {
	w := T.Geometry.Width
	h := T.Geometry.Height
	r := sdl.Rect{
		X: T.Padding.Left,
		Y: T.Padding.Top + int32(len(T.Content)*T.LineHeight),
		W: int32(w),
		H: int32(h),
	}
	T.Content = append(T.Content, line)
	T.DrawColoredText(line.Content,
		&r, "foreground", "default",
		line.Rules,
	)
	// T.Renderer.Clear()
	// T.Renderer.Present()
	// sdl.Delay(1)
	T.App.Window.UpdateSurface()
}

// ClearContent is
func (T *TextWidget) ClearContent() {
	T.Content = []Line{}
}

// Reset is
func (T *TextWidget) Reset() {
	T.ClearContent()
	T.Clear()
}

// Update is
func (T *TextWidget) Update() {
	w := T.Geometry.Width
	// h := T.Geometry.Height
	var r sdl.Rect
	T.Clear()
	for i, line := range T.Content {
		r = sdl.Rect{
			X: T.Padding.Left,
			Y: T.Padding.Top + int32(i*T.LineHeight),
			W: int32(w),
			H: int32(T.LineHeight),
		}
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

// Show is
func (T *TextWidget) Show() {
	T.Renderer.Present()
	T.App.Window.UpdateSurface()
}

// StripLine is
func (T *TextWidget) StripLine(line string, fontname string) string {
	w := T.Geometry.Width
	lw, _, _ := T.Fonts[fontname].SizeUTF8(line)
	for int32(lw) > (int32(w) - T.Padding.Left*2) {
		line = strings.TrimRight(line[:len(line)-4], " -") + "…"
		lw, _, _ = T.Fonts[fontname].SizeUTF8(line)
	}
	return line
}

// FullClear is
func (T *TextWidget) FullClear() {
	w := T.Geometry.Width
	h := T.Geometry.Height
	rect := sdl.Rect{
		X: 0,
		Y: 0,
		W: int32(w),
		H: int32(h),
	}
	T.Surface.FillRect(&rect, T.BG)
}

// Clear is
func (T *TextWidget) Clear() {
	w := T.Geometry.Width - T.Padding.Left - T.Padding.Right + 3
	h := T.Geometry.Height - T.Padding.Top
	rect := sdl.Rect{
		X: T.Padding.Left,
		Y: T.Padding.Top,
		W: int32(w),
		H: int32(h),
	}
	T.Surface.FillRect(&rect, T.BG)
}

// DrawText is
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

// HighlightRule is highlighting rule
type HighlightRule struct {
	Start int
	Len   int
	Color string
	Font  string
}

// DrawColoredText is
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
				rect = &sdl.Rect{
					X: rect.X + int32(tw),
					Y: rect.Y,
					W: rect.W - int32(tw),
					H: rect.H,
				}
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
			rect = &sdl.Rect{
				X: rect.X + int32(tw),
				Y: rect.Y,
				W: rect.W - int32(tw),
				H: rect.H,
			}
			text = text[l:]
			// log.Println(text)
		}
		if len(token) > 0 {
			T.DrawText(text, rect, colorName, fontName)
		}
	}
}

// MoveCursor is
func (T *TextWidget) MoveCursor(r int, c int) (int, int) {
	T.Cursor.Row = r
	line := T.Content[T.Cursor.Row]
	if c >= 0 && c <= len(line.Content) {
		T.Cursor.Column = c
	}
	T.drawCursor()
	return T.Cursor.Row, T.Cursor.Column
}

// MoveCursorLeft is
func (T *TextWidget) MoveCursorLeft() (int, int) {
	T.MoveCursor(T.Cursor.Row, T.Cursor.Column-1)
	return T.Cursor.Row, T.Cursor.Column
}

// MoveCursorRight is
func (T *TextWidget) MoveCursorRight() (int, int) {
	T.MoveCursor(T.Cursor.Row, T.Cursor.Column+1)
	return T.Cursor.Row, T.Cursor.Column
}

// MoveCursorUp is
func (T *TextWidget) MoveCursorUp() (int, int) {
	T.MoveCursor(T.Cursor.Row-1, T.Cursor.Column)
	return T.Cursor.Row, T.Cursor.Column
}

// MoveCursorDown is
func (T *TextWidget) MoveCursorDown() (int, int) {
	T.MoveCursor(T.Cursor.Row+1, T.Cursor.Column)
	return T.Cursor.Row, T.Cursor.Column
}

// addString is
func (T *TextWidget) addString(s string) (int, int) {
	line := T.Content[T.Cursor.Row]
	i := T.Cursor.Column
	line.Content = line.Content[:i] + s + line.Content[i:]
	T.SetLine(0, line)
	T.MoveCursor(T.Cursor.Row, T.Cursor.Column+len(s))
	return T.Cursor.Row, T.Cursor.Column
}

// removeString is
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

// removeStringForward is
func (T *TextWidget) removeStringForward(n int) (int, int) {
	line := T.Content[T.Cursor.Row]
	i := T.Cursor.Column
	log.Println(i+n, i, len(line.Content))
	if i+n > len(line.Content) {
		n = len(line.Content) - i
	}
	if i > 0 {
		line.Content = line.Content[:i] + line.Content[i+n:]
		T.SetLine(0, line)
		T.MoveCursor(T.Cursor.Row, T.Cursor.Column)
	}
	return T.Cursor.Row, T.Cursor.Column
}

// removeWord is
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

// drawCursor is
func (T *TextWidget) drawCursor() {
	if T.Cursor.Hidden {
		return
	}
	index := T.Cursor.Row
	var lw int
	if T.Cursor.Column > 0 {
		line := T.Content[index].Content[:T.Cursor.Column-1]
		lw, _, _ = T.Fonts["default"].SizeUTF8(line)
	} else {
		lw = -5
	}
	rect := sdl.Rect{
		X: T.Padding.Left - 3,
		Y: T.Padding.Top,
		W: 3,
		H: int32(T.LineHeight),
	}
	T.Surface.FillRect(&rect, T.BG)
	r := sdl.Rect{
		X: T.Padding.Left + int32(lw) + 8,
		Y: T.Padding.Top + int32(index*T.LineHeight),
		W: int32(8),
		H: int32(T.LineHeight),
	}
	T.DrawColoredText("_", &r, "orange", "default", []HighlightRule{})
	T.Renderer.Present()
	T.App.Window.UpdateSurface()
}

// SetRules is
func (T *TextWidget) SetRules(index int, rules []HighlightRule) {
	line := T.Content[index]
	line.Rules = rules
	T.SetLine(index, line)
	T.drawCursor()
}
