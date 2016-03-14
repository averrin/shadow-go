package main

import (
	"fmt"
	"io/ioutil"
	// "log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/renstrom/fuzzysearch/fuzzy"
	"github.com/veandco/go-sdl2/sdl"
)

type Runner struct {
	App      *Application
	Items    []string
	Suggests []string
	Selected int
}

func (R *Runner) SetApp(app *Application) {
	R.App = app
}

func getExec() []string {
	var ret []string
	pathes := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range pathes {
		fi, _ := ioutil.ReadDir(path)
		for n := range fi {
			ret = append(ret, fi[n].Name())
		}
	}
	return ret
}

func (R *Runner) Init() {
	app := R.App
	window := R.App.Window
	fontSize = 14
	w := 500
	h := (fontSize + 10) * 13
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
	R.Items = getExec()
	R.Suggests = R.Items[:13]
}

func (R *Runner) Draw() {
	c := make(chan Line)
	go GetTime(c)
	app := R.App
	T := app.Widget
	T.Padding.Left = 30
	T.Reset()
	T.App.DrawMode()
	r := sdl.Rect{T.Padding.Left - 14, T.Padding.Top - 1, 6, int32(T.LineHeight)}
	T.DrawColoredText(">", &r, "accent", "bold", []HighlightRule{})
	T.AddLine(Line{"", []HighlightRule{}})
	T.MoveCursor(0, 0)
	for _, e := range R.Items[:13] {
		T.AddLine(Line{e, []HighlightRule{}})
	}
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}

func (R *Runner) Run() int {
	app := R.App
	T := app.Widget
	// window := R.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				// R.Draw()
			}
			if t.Event == sdl.WINDOWEVENT_FOCUS_LOST {
				// return 0
			}
		case *sdl.QuitEvent:
			return 0
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard\ttype:%d\tname:%s\tmodifiers:%d\tstate:%d\trepeat:%d\tsym: %c\n",
				t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat, t.Keysym.Sym)
			key := sdl.GetScancodeName(t.Keysym.Scancode)
			if key == "Escape" || key == "CapsLock" {
				return 0
			}
			if (key == "H" && t.Keysym.Mod == 64) || key == "Backspace" {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.removeString(1)
				R.Update()
				return 1
			}
			if key == "Delete" {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.removeStringForward(1)
				R.Update()
				return 1
			}
			if key == "C" && t.Keysym.Mod == 64 {
				line := T.Content[0]
				line.Content = line.Content[:2]
				T.SetLine(0, line)
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.MoveCursor(0, 0)
				return 1
			}
			if key == "V" && t.Keysym.Mod == 64 {
				s, _ := sdl.GetClipboardText()
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.addString(s)
				R.Update()
				return 1
			}
			if key == "W" && t.Keysym.Mod == 64 {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.removeWord()
				R.Update()
				return 1
			}
			if key == "Left" {
				T.MoveCursorLeft()
				R.Update()
				return 1
			}
			if key == "Right" {
				T.MoveCursorRight()
				R.Update()
				return 1
			}
			if key == "Down" || (key == "N" && t.Keysym.Mod == 64) {
				R.Next()
				return 1
			}
			if key == "Up" || (key == "P" && t.Keysym.Mod == 64) {
				R.Prev()
				return 1
			}
			if key == "Tab" {
				R.Autocomplete()
				return 1
			}
			if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
				ret := Exec(T.Content[0].Content)
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
					if char == "-" {
						char = "_"
					}
				}
				T.addString(char)
				R.Update()
				return 1
			}
		}
	}
	return 1
}

func (R *Runner) Next() {
	T := R.App.Widget
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "default", "default"}})
	R.Selected++
	if R.Selected == len(R.Suggests) {
		R.Selected = 0
	}
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
	T.SetLine(0, Line{R.Suggests[R.Selected], []HighlightRule{HighlightRule{0, -1, "green", "default"}}})
	T.MoveCursor(0, len(R.Suggests[R.Selected]))
}

func (R *Runner) Prev() {
	T := R.App.Widget
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "default", "default"}})
	if R.Selected == 0 {
		R.Selected = len(R.Suggests)
	}
	R.Selected--
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
	T.SetLine(0, Line{R.Suggests[R.Selected], []HighlightRule{HighlightRule{0, -1, "green", "default"}}})
	T.MoveCursor(0, len(R.Suggests[R.Selected]))
}

func (R *Runner) Autocomplete() {
	T := R.App.Widget
	tokens := strings.Split(T.Content[0].Content, " ")
	line := tokens[0]
	items := fuzzy.RankFindFold(line, R.Items)
	sort.Sort(items)
	line = items[0].Target
	if len(tokens) > 1 {
		line += " " + strings.Join(tokens[1:], " ")
	}
	T.SetLine(0, Line{line, []HighlightRule{HighlightRule{0, len(line), "green", "default"}}})
	T.MoveCursor(0, len(line))
	R.Update()
}

func (R *Runner) Update() {
	R.Selected = 0
	T := R.App.Widget
	line := strings.Split(T.Content[0].Content, " ")[0]
	items := fuzzy.RankFindFold(line, R.Items)
	end := 13
	if end > len(items) {
		end = len(items)
	}
	if len(items) > 0 {
		sort.Sort(items)
		R.Suggests = func() []string {
			r := []string{}
			for _, s := range items[:end] {
				r = append(r, s.Target)
			}
			return r
		}()
		if line == R.Suggests[0] {
			T.SetRules(0, []HighlightRule{HighlightRule{0, len(line), "green", "default"}})
		} else {
			if len(R.Suggests) == 1 {
				T.SetRules(0, []HighlightRule{HighlightRule{0, len(line), "accent", "default"}})
			} else {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
			}
		}
	}
	newContent := T.Content[:1]
	for _, item := range R.Suggests {
		newContent = append(newContent, Line{item, []HighlightRule{}})
	}
	T.SetContent(newContent)
	if len(R.Suggests) == 1 {
		first := R.Suggests[0]
		if line != first && line == first[:len(line)] && len(strings.Split(T.Content[0].Content, " ")) == 1 {
			suggest := first[len(line):]
			ll, _, _ := T.Fonts["default"].SizeUTF8(line)
			sl, _, _ := T.Fonts["default"].SizeUTF8(suggest)
			r := sdl.Rect{T.Padding.Left + int32(ll), T.Padding.Top, int32(sl), int32(T.LineHeight)}
			T.DrawColoredText(suggest, &r, "gray", "default", []HighlightRule{})
			T.Show()
		}
	}
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
}

func Exec(cmd string) int {
	tokens := strings.Split(cmd, " ")
	c := exec.Command(tokens[0], tokens[1:]...)
	err := c.Start()
	if err != nil {
		return 1
	}
	return 0
}