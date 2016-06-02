package main

import (
	"log"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

//Command interface
type Command interface {
	Init()
	Test(string) bool
	GetText(string) string
	Exec(string) int
	GetSuggests(string) []string
}

//Ultra mode
type Ultra struct {
	App      *Application
	Alias    string
	History  []string
	Items    []Command
	Suggests []string
	Selected int
}

//SetApp interface method
func (U *Ultra) SetApp(app *Application) {
	U.App = app
	U.Alias = "\ue721"
}

//GetAlias interface method
func (U *Ultra) GetAlias() string {
	return U.Alias
}

//Init interface method
func (U *Ultra) Init() WidgetSettings {
	U.Selected = -1
	app := U.App
	window := U.App.Window
	fontSize = 14
	w := 500
	h := (fontSize + 10) * 13
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
	// U.History = readExec()
	U.Items = []Command{
		new(SearchCommand),
		new(RunCommand),
		new(SessionCommand),
		new(TasksCommand),
	}
	for n := range U.Items {
		U.Items[n].Init()
	}
	// U.Suggests = U.Items[:12]

	return WidgetSettings{fontSize, Geometry{int32(w), int32(h)}, Padding{10, 30, 10}}
}

//Draw interface method
func (U *Ultra) Draw() {
	app := U.App
	T := app.Widget
	T.Reset()
	T.App.DrawMode()
	r := sdl.Rect{
		X: T.Padding.Left - 14,
		Y: T.Padding.Top,
		W: 6,
		H: int32(T.LineHeight),
	}
	T.DrawColoredText("\uf054", &r, ACCENT, "bold", []HighlightRule{})
	T.AddLine(Line{"", []HighlightRule{}})
	T.MoveCursor(0, 0)
	T.ChangeLine(1, Line{"Type anything...", []HighlightRule{HighlightRule{0, -1, "gray", "default"}}})
	// b := math.Min(float64(len(U.Items)), float64(12))
	// for _, e := range U.Items[:int(b)] {
	// 	T.AddLine(Line{e.GetText(""), []HighlightRule{}})
	// }
}

//DispatchEvents interface method
func (U *Ultra) DispatchEvents(event sdl.Event) int {
	return 1
}

//DispatchKeys interface method
func (U *Ultra) DispatchKeys(t *sdl.KeyDownEvent) int {
	app := U.App
	T := app.Widget
	// fmt.Printf("[%d ms] Keyboard\ttype:%d\tname:%s\tmodifiers:%d\tstate:%d\trepeat:%d\tsym: %c\n",
	// 	t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat, t.Keysym.Sym)
	key := sdl.GetScancodeName(t.Keysym.Scancode)
	if t.Keysym.Sym == sdl.K_ESCAPE || t.Keysym.Sym == sdl.K_CAPSLOCK {
		return 0
	}
	if (key == "H" && t.Keysym.Mod == 64) || key == "Backspace" {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.removeString(1)
		U.update()
		return 1
	}
	if key == "Delete" {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.removeStringForward(1)
		U.update()
		return 1
	}
	if key == "C" && t.Keysym.Mod == 64 {
		line := T.Content[0]
		T.MoveCursor(0, 0)
		line.Content = ""
		T.ChangeLine(0, line)
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		return 1
	}
	if key == "V" && t.Keysym.Mod == 64 {
		s, _ := sdl.GetClipboardText()
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.addString(s)
		U.update()
		return 1
	}
	if key == "W" && t.Keysym.Mod == 64 {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.removeWord()
		U.update()
		return 1
	}
	if key == "Left" {
		T.MoveCursorLeft()
		U.update()
		return 1
	}
	if key == "Right" {
		T.MoveCursorRight()
		U.update()
		return 1
	}
	// if key == "Down" || (key == "N" && t.Keysym.Mod == 64) {
	// 	R.next()
	// 	return 1
	// }
	// if key == "Up" || (key == "P" && t.Keysym.Mod == 64) {
	// 	R.prev()
	// 	return 1
	// }
	if key == "Tab" {
		U.autocomplete()
		return 1
	}
	if (key == "J" && t.Keysym.Mod == 64) || t.Keysym.Sym == sdl.K_RETURN {
		ret := U.execInput(T.Content[0].Content)
		if ret != 0 {
			T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "red", "default"}})
			T.drawCursor()
		}
		return ret
	}
	symbols := map[string]string{
		"-": "_",
		"=": "+",
		";": ";",
		"1": "!",
		"2": "@",
		"3": "#",
		"4": "$",
	}
	if isASCII(string(t.Keysym.Sym)) && t.Keysym.Mod <= 1 {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		char := string(t.Keysym.Sym)
		if t.Keysym.Mod == 1 {
			char = strings.ToUpper(char)
			sub, ok := symbols[char]
			if ok {
				char = sub
			}
		}
		T.addString(char)
		U.update()
		return 1
	}
	return 1
}

func (U *Ultra) autocomplete() {
	app := U.App
	T := app.Widget
	T.ChangeLine(0, Line{U.Suggests[0], []HighlightRule{HighlightRule{0, -1, GREEN, "default"}}})
	U.update()
}

func (U *Ultra) update() {
	app := U.App
	T := app.Widget
	line := T.Content[0].Content
	log.Println(line)
	U.Suggests = []string{}
	for n := range U.Items {
		if U.Items[n].Test(line) {
			T.ChangeLine(1, Line{U.Items[n].GetText(line), []HighlightRule{}})
			return
		}
		if line != "" {
			s := U.Items[n].GetSuggests(line)
			for i := range s {
				U.Suggests = append(U.Suggests, s[i])
			}
		}
	}
	if line != "" {
		T.ChangeLine(1, Line{"No results... Confirm to search in Google.", []HighlightRule{HighlightRule{0, -1, "gray", "default"}}})
		for i := range U.Suggests {
			l := U.Suggests[i]
			log.Println(l)
			T.ChangeLine(1+i, Line{l, []HighlightRule{}})
		}
	} else {
		T.ChangeLine(1, Line{"Type anything...", []HighlightRule{HighlightRule{0, -1, "gray", "default"}}})
	}
}

func (U *Ultra) execInput(line string) int {
	log.Println(line)
	for n := range U.Items {
		if U.Items[n].Test(line) {
			return U.Items[n].Exec(line)
		}
	}
	if line != "" {
		return U.execInput("g " + line)
	}
	return 1
}
