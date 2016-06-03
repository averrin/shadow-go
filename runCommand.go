package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
)

type RunCommand struct {
	Mapping map[string]func(string) int
	cmd     []string
}

func (Cmd *RunCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"!": execCommand,
	}
	pathes := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range pathes {
		fi, _ := ioutil.ReadDir(path)
		for n := range fi {
			line := fi[n].Name()
			if !stringInSlice(line, Cmd.cmd) {
				Cmd.cmd = append(Cmd.cmd, "! "+line)
			}
		}
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

func (Cmd *RunCommand) GetText(line string) Line {
	return Line{fmt.Sprintf("Run commands: %s", line[2:]), []HighlightRule{}}
}

func (Cmd *RunCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	for c := range Cmd.cmd {
		cmd := Cmd.cmd[c]
		if strings.HasPrefix(cmd, line) {
			s = append(s, AutocompleteItem{Cmd, cmd})
		}
	}
	l := math.Min(float64(len(s)), float64(12))
	return s[:int(l)]
}
