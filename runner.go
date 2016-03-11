package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

func (sw *Runner) SetApp(app *Application) {
	sw.App = app
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

func (sw *Runner) Init() {
	app := sw.App
	window := sw.App.Window
	fontSize = 14
	w := 500
	h := (fontSize + 10) * 13
	window, err := sdl.CreateWindow("Shadow", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	app.Window = window
	sw.Items = getExec()
}

func (sw *Runner) Draw() {
	c := make(chan Line)
	go GetTime(c)
	app := sw.App
	T := app.Widget
	T.Padding.Left = 30
	T.Reset()
	T.App.DrawMode()
	r := sdl.Rect{T.Padding.Left - 14, T.Padding.Top - 1, 6, int32(T.LineHeight)}
	T.DrawColoredText(">", &r, "accent", "bold", []HighlightRule{})
	T.AddLine(Line{"", []HighlightRule{}})
	T.MoveCursor(0, 0)
	for _, e := range sw.Items[:13] {
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

func (sw *Runner) Run() int {
	app := sw.App
	T := app.Widget
	// window := sw.App.Window
	var event sdl.Event
	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_FOCUS_GAINED {
				// sw.Draw()
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
				sw.Update()
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
				sw.Update()
				return 1
			}
			if key == "W" && t.Keysym.Mod == 64 {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.removeWord()
				sw.Update()
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
			if isASCII(string(t.Keysym.Sym)) && t.Keysym.Mod == 0 {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
				T.addString(fmt.Sprintf("%c", t.Keysym.Sym))
				sw.Update()
				return 1
			}
		}
	}
	return 1
}

func (R *Runner) Update() {
	T := R.App.Widget
	line := T.Content[0]
	items := fuzzy.RankFindFold(line.Content, R.Items)
	sort.Sort(items)
	log.Println(items)
	end := 13
	if end > len(items) {
		end = len(items)
	}
	if len(items) == 1 {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "green", "default"}})
		if line.Content != items[0].Target {
			T.Content[0].Content = items[0].Target
			T.MoveCursor(0, len(items[0].Target))
		}
	} else {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
	}
	newContent := T.Content[:1]
	for _, item := range items[:end] {
		newContent = append(newContent, Line{item.Target, []HighlightRule{}})
	}
	// T.Reset()
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
