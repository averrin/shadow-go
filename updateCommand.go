package main

import (
	"fmt"
	"strings"
)

type UpdateCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *UpdateCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"update": func(string) int {
			return execCommandShell("sudo apt full-upgrade --fix-missing -f -y --allow-downgrades")
		},
	}
}

func (Cmd *UpdateCommand) Test(line string) bool {
	_, ok := Cmd.Mapping[line]
	return ok
}

func (Cmd *UpdateCommand) Exec(line string) int {
	return Cmd.Mapping[line](line)
}

func (Cmd *UpdateCommand) GetText(line string) Line {
	// line[0] = strings.ToUpper(line[0])
	return Line{fmt.Sprintf("Update system"), []HighlightRule{}}
}

func (Cmd *UpdateCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	for c := range Cmd.Mapping {
		if strings.HasPrefix(c, line) {
			s = append(s, AutocompleteItem{Cmd, c})
		}
	}
	return s
}
