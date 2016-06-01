package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	// "log"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strings"

	"github.com/renstrom/fuzzysearch/fuzzy"
	"github.com/veandco/go-sdl2/sdl"
)

//Runner mode
type Runner struct {
	App      *Application
	Alias    string
	History  []string
	Items    []string
	Suggests []string
	Selected int
}

//SetApp interface method
func (R *Runner) SetApp(app *Application) {
	R.App = app
	R.Alias = "\uf120"
}

//GetAlias interface method
func (R *Runner) GetAlias() string {
	return R.Alias
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func reverse(arr []string) []string {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func (R *Runner) getExec() []string {
	ret := reverse(R.History)
	pathes := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range pathes {
		fi, _ := ioutil.ReadDir(path)
		for n := range fi {
			line := fi[n].Name()
			if !stringInSlice(line, ret) {
				ret = append(ret, line)
			}
		}
	}
	return ret
}

//Init interface method
func (R *Runner) Init() WidgetSettings {
	R.Selected = -1
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
	R.History = readExec()
	R.Items = R.getExec()
	R.Suggests = R.Items[:12]
	return WidgetSettings{fontSize, Geometry{int32(w), int32(h)}, Padding{10, 30, 10}}
}

//Draw interface method
func (R *Runner) Draw() {
	c := make(chan Line)
	go GetTime(c)
	app := R.App
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
	for _, e := range R.Items[:12] {
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

// DispatchKeys is main mode loop
func (R *Runner) DispatchKeys(t *sdl.KeyDownEvent) int {
	app := R.App
	T := app.Widget
	fmt.Printf("[%d ms] Keyboard\ttype:%d\tname:%s\tmodifiers:%d\tstate:%d\trepeat:%d\tsym: %c\n",
		t.Timestamp, t.Type, sdl.GetScancodeName(t.Keysym.Scancode), t.Keysym.Mod, t.State, t.Repeat, t.Keysym.Sym)
	key := sdl.GetScancodeName(t.Keysym.Scancode)
	if t.Keysym.Sym == sdl.K_ESCAPE || t.Keysym.Sym == sdl.K_CAPSLOCK {
		return 0
	}
	if (key == "H" && t.Keysym.Mod == 64) || key == "Backspace" {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.removeString(1)
		R.update()
		return 1
	}
	if key == "Delete" {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.removeStringForward(1)
		R.update()
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
		R.update()
		return 1
	}
	if key == "W" && t.Keysym.Mod == 64 {
		T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
		T.removeWord()
		R.update()
		return 1
	}
	if key == "Left" {
		T.MoveCursorLeft()
		R.update()
		return 1
	}
	if key == "Right" {
		T.MoveCursorRight()
		R.update()
		return 1
	}
	if key == "Down" || (key == "N" && t.Keysym.Mod == 64) {
		R.next()
		return 1
	}
	if key == "Up" || (key == "P" && t.Keysym.Mod == 64) {
		R.prev()
		return 1
	}
	if key == "Tab" {
		R.autocomplete()
		return 1
	}
	if (key == "J" && t.Keysym.Mod == 64) || key == "Return" {
		ret := execCommand(T.Content[0].Content)
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
		R.update()
		return 1
	}
	return 1
}

//DispatchEvents interface method
func (R *Runner) DispatchEvents(event sdl.Event) int {
	return 1
}

func (R *Runner) next() {
	T := R.App.Widget
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "default", "default"}})
	R.Selected++
	if R.Selected == len(R.Suggests) {
		R.Selected = 0
	}
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
	T.ChangeLine(0, Line{R.Suggests[R.Selected], []HighlightRule{HighlightRule{0, -1, GREEN, "default"}}})
	T.MoveCursor(0, len(R.Suggests[R.Selected]))
}

func (R *Runner) prev() {
	T := R.App.Widget
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "default", "default"}})
	if R.Selected == 0 {
		R.Selected = len(R.Suggests)
	}
	R.Selected--
	T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
	T.ChangeLine(0, Line{R.Suggests[R.Selected], []HighlightRule{HighlightRule{0, -1, GREEN, "default"}}})
	T.MoveCursor(0, len(R.Suggests[R.Selected]))
}

func (R *Runner) autocomplete() {
	T := R.App.Widget
	tokens := strings.Split(T.Content[0].Content, " ")
	line := tokens[0]
	items := fuzzy.RankFindFold(line, R.Items)
	if len(items) > 0 {
		sort.Sort(items)
		line = items[0].Target
		if len(tokens) > 1 {
			line += " " + strings.Join(tokens[1:], " ")
		}
		T.ChangeLine(0, Line{line, []HighlightRule{HighlightRule{0, len(line), GREEN, "default"}}})
		T.MoveCursor(0, len(line))
		R.update()
	}
}

func (R *Runner) update() {
	R.Selected = -1
	T := R.App.Widget
	line := strings.Split(T.Content[0].Content, " ")[0]
	items := fuzzy.RankFindFold(line, R.Items)
	end := 12
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
			T.SetRules(0, []HighlightRule{HighlightRule{0, len(line), GREEN, "default"}})
		} else {
			if len(R.Suggests) == 1 {
				T.SetRules(0, []HighlightRule{HighlightRule{0, len(line), ACCENT, "default"}})
			} else {
				T.SetRules(0, []HighlightRule{HighlightRule{0, -1, "foreground", "default"}})
			}
		}
	}
	tmp := make([]Line, len(T.Content))
	copy(tmp, T.Content)
	newContent := tmp[:1]
	for _, item := range R.Suggests {
		newContent = append(newContent, Line{item, []HighlightRule{}})
	}
	if len(items) > len(R.Suggests) {
		newContent = append(newContent, Line{fmt.Sprintf("%d more…", len(items)-12), []HighlightRule{HighlightRule{0, -1, "gray", "default"}}})
	}
	T.SetContent(newContent)
	// if len(R.Suggests) == 1 {
	first := R.Suggests[0]
	if line != first && len(first) > len(line) && line == first[:len(line)] && len(strings.Split(T.Content[0].Content, " ")) == 1 {
		suggest := first[len(line):]
		ll, _, _ := T.Fonts["default"].SizeUTF8(line)
		sl, _, _ := T.Fonts["default"].SizeUTF8(suggest)
		r := sdl.Rect{
			X: T.Padding.Left + int32(ll),
			Y: T.Padding.Top,
			W: int32(sl),
			H: int32(T.LineHeight),
		}
		T.DrawColoredText(suggest, &r, "gray", "default", []HighlightRule{})
		T.Show()
	}
	// }
	// T.SetRules(R.Selected+1, []HighlightRule{HighlightRule{0, -1, "highlight", "bold"}})
}

func execCommand(cmd string) int {
	tokens := strings.Split(cmd, " ")
	c := exec.Command(tokens[0], tokens[1:]...)
	err := c.Start()
	if err != nil {
		return 1
	}
	saveExec(cmd)
	return 0
}

func readExec() []string {
	ret := []string{}
	usr, _ := user.Current()
	filename := path.Join(usr.HomeDir, ".shadow_history")
	if _, err := os.Stat(filename); err == nil {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				ret = append(ret, line)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	return ret
}

func saveExec(text string) {
	lines := readExec()
	if stringInSlice(text, lines) {
		return
	}
	usr, err := user.Current()
	filename := path.Join(usr.HomeDir, ".shadow_history")
	if _, err = os.Stat(filename); err != nil {
		os.Create(filename)
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	if _, err = f.WriteString(text + "\n"); err != nil {
		panic(err)
	}
}
