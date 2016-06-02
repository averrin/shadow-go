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
	return fmt.Sprintf("%s session", strings.ToTitle(line))
}
