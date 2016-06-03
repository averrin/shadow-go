package main

import (
	"log"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

var symbols map[string]string

//Command interface
type Command interface {
	Init()
	Test(string) bool
	GetText(string) Line
	Exec(string) int
	GetSuggests(string) []AutocompleteItem
}

//AutocompleteItem is
type AutocompleteItem struct {
	Command Command
	Text    string
}

//Ultra mode
type Ultra struct {
	App      *Application
	Alias    string
	History  []string
	Items    []Command
	Suggests []AutocompleteItem
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
		new(MathCommand),
		new(ColorCommand),
	}
	for n := range U.Items {
		U.Items[n].Init()
	}

	symbols = map[string]string{
		"-": "_",
		"=": "+",
		";": ";",
		"1": "!",
		"2": "@",
		"3": "#",
		"4": "$",
		"5": "%",
		"6": "^",
		"7": "&",
		"8": "*",
		"9": "(",
		"0": ")",
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
	if key == "Down" || (key == "N" && t.Keysym.Mod == 64) {
		U.next()
		return 1
	}
	if key == "Up" || (key == "P" && t.Keysym.Mod == 64) {
		U.prev()
		return 1
	}
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

func (U *Ultra) next() {
	T := U.App.Widget
	T.SetRules(U.Selected+1, []HighlightRule{HighlightRule{0, -1, "default", "default"}})
	U.Selected++
	if U.Selected == len(U.Suggests) {
		U.Selected = 0
	}
	T.SetRules(U.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
	T.ChangeLine(0, Line{U.Suggests[U.Selected].Text, []HighlightRule{HighlightRule{0, -1, GREEN, "default"}}})
	T.MoveCursor(0, len(U.Suggests[U.Selected].Text))
}

func (U *Ultra) prev() {
	T := U.App.Widget
	T.SetRules(U.Selected+1, []HighlightRule{HighlightRule{0, -1, "default", "default"}})
	if U.Selected == 0 {
		U.Selected = len(U.Suggests)
	}
	U.Selected--
	T.SetRules(U.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
	T.ChangeLine(0, Line{U.Suggests[U.Selected].Text, []HighlightRule{HighlightRule{0, -1, GREEN, "default"}}})
	T.MoveCursor(0, len(U.Suggests[U.Selected].Text))
}

func (U *Ultra) autocomplete() {
	app := U.App
	T := app.Widget
	line := T.Content[0].Content
	i := U.Selected
	if i == -1 {
		i = 0
		U.Selected = 0
	}
	if line != U.Suggests[i].Text {
		T.ChangeLine(0, Line{U.Suggests[i].Text, []HighlightRule{HighlightRule{0, -1, GREEN, "default"}}})
		T.MoveCursor(0, len(U.Suggests[i].Text))
		U.update()
	} else {
		U.next()
	}
}

func (U *Ultra) update() {
	app := U.App
	T := app.Widget
	line := T.Content[0].Content
	log.Println(line)
	U.Suggests = []AutocompleteItem{}
	for n := range U.Items {
		if line != "" {
			s := U.Items[n].GetSuggests(line)
			for i := range s {
				U.Suggests = append(U.Suggests, s[i])
			}
		}
	}
	tmp := make([]Line, len(T.Content))
	copy(tmp, T.Content)
	newContent := tmp[:1]
	if line != "" {
		if len(U.Suggests) > 1 {
			for _, item := range U.Suggests {
				newContent = append(newContent, item.Command.GetText(item.Text))
			}
		} else {
			if len(U.Suggests) == 1 {
				newContent = append(newContent, U.Suggests[0].Command.GetText(U.Suggests[0].Text))
			} else {
				newContent = append(newContent, Line{"No results... Confirm to search in Google.", []HighlightRule{HighlightRule{0, -1, "gray", "default"}}})
			}
		}
	} else {
		newContent = append(newContent, Line{"Type anything...", []HighlightRule{HighlightRule{0, -1, "gray", "default"}}})
	}
	T.SetContent(newContent)
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
