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
	App   *Application
	Items []string
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

func (R *Runner) Autocomplete() {
	T := R.App.Widget
	line := strings.Split(T.Content[0].Content, " ")[0]
	items := fuzzy.RankFindFold(line, R.Items)
	sort.Sort(items)
	line = items[0].Target
	T.SetLine(0, Line{line, []HighlightRule{HighlightRule{0, len(line), "green", "default"}}})
	T.MoveCursor(0, len(line))
	R.Update()
}

func (R *Runner) Update() {
	T := R.App.Widget
	line := strings.Split(T.Content[0].Content, " ")[0]
	items := fuzzy.RankFindFold(line, R.Items)
	sort.Sort(items)
	end := 13
	if end > len(items) {
		end = len(items)
	}
	if len(items) == 1 {
		if line != items[0].Target {
			// T.Content[0].Content = items[0].Target
			// T.MoveCursor(0, len(items[0].Target))
			T.SetRules(0, []HighlightRule{HighlightRule{0, len(line), "accent", "default"}})
		} else {
			T.SetRules(0, []HighlightRule{HighlightRule{0, len(line), "green", "default"}})
		}
	} else {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
	}
	newContent := T.Content[:1]
	for _, item := range items[:end] {
		newContent = append(newContent, Line{item.Target, []HighlightRule{}})
	}
	T.SetContent(newContent)
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
