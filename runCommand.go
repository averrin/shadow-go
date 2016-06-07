package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strings"
)

type RunCommand struct {
	Mapping map[string]func(string) int
	cmd     []string
}

func (Cmd *RunCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"!":  execCommand,
		"!!": execCommandShell,
	}
	pathes := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range pathes {
		fi, _ := ioutil.ReadDir(path)
		for _, info := range fi {
			line := info.Name()
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
	return Line{fmt.Sprintf("Run commands: %s", line[2:]), []HighlightRule{HighlightRule{14, -1, "foreground", "bold"}}}
}

func (Cmd *RunCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	for _, cmd := range Cmd.cmd {
		if strings.HasPrefix(cmd, line) {
			s = append(s, AutocompleteItem{Cmd, cmd})
		}
	}
	l := math.Min(float64(len(s)), float64(12))
	return s[:int(l)]
}

func execCommandShell(cmd string) int {
	cmd = "konsole -e " + cmd
	tokens := strings.Split(cmd, " ")
	c := exec.Command(tokens[0], tokens[1:]...)
	err := c.Start()
	if err != nil {
		return 1
	}
	return 0
}
