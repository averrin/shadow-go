package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

//ColorCommand show hex color
type ColorCommand struct {
	Mapping map[string]func(string) int
}

//Init is
func (Cmd *ColorCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"#": func(string) int { return 0 },
	}
}

//Test is
func (Cmd *ColorCommand) Test(line string) bool {
	_, ok := Cmd.Mapping[line]
	return ok
}

//Exec is
func (Cmd *ColorCommand) Exec(line string) int {
	return Cmd.Mapping[line](line)
}

//GetText is
func (Cmd *ColorCommand) GetText(line string) Line {
	// line[0] = strings.ToUpper(line[0])
	color, err := hex.DecodeString(line[1:])
	if err == nil && len(color) == 3 {
		return Line{fmt.Sprintf("%v: \u2588\u2588\u2588\u2588\u2588\u2588", line), []HighlightRule{HighlightRule{0, -1, line, "default"}}}
	}
	if strings.TrimSpace(line) != "#" {
		return Line{fmt.Sprintf("Wrong color"), []HighlightRule{HighlightRule{0, -1, "red", "default"}}}
	}
	return Line{fmt.Sprintf("Type hex color"), []HighlightRule{}}
}

//GetSuggests is
func (Cmd *ColorCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	if strings.HasPrefix(line, "#") {
		s = append(s, AutocompleteItem{Cmd, line})
	}
	return s
}
