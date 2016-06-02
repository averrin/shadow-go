package main

import (
	"fmt"
	"strings"
)

type RunCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *RunCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"!": execCommand,
	}
}

func (Cmd *RunCommand) getPrefix(line string) string {
	for p := range Cmd.Mapping {
		if strings.HasPrefix(line, p+" ") {
			return p
		}
	}
	return ""
}

func (Cmd *RunCommand) Test(line string) bool {
	return Cmd.getPrefix(line) != ""
}

func (Cmd *RunCommand) Exec(line string) int {
	p := Cmd.getPrefix(line)
	return Cmd.Mapping[p](line[len(p)+1:])
}

func (Cmd *RunCommand) GetText(line string) string {
	return fmt.Sprintf("Run commands: %s", line[2:])
}

func (Cmd *RunCommand) GetSuggests(line string) []string {
	return []string{}
}
