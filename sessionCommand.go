package main

import (
	"fmt"
	"strings"
)

type SessionCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *SessionCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"lock": func(string) int {
			return execCommand("qdbus org.freedesktop.ScreenSaver /ScreenSaver Lock")
		},
	}
}

func (Cmd *SessionCommand) Test(line string) bool {
	_, ok := Cmd.Mapping[line]
	return ok
}

func (Cmd *SessionCommand) Exec(line string) int {
	return Cmd.Mapping[line](line)
}

func (Cmd *SessionCommand) GetText(line string) string {
	// line[0] = strings.ToUpper(line[0])
	return fmt.Sprintf("%s session", line)
}

func (Cmd *SessionCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	for c := range Cmd.Mapping {
		if strings.HasPrefix(c, line) {
			s = append(s, AutocompleteItem{Cmd, c})
		}
	}
	return s
}
