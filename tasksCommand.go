package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
)

type TasksCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *TasksCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"switch ": func(string) int { return 0 },
	}
	clients := GetClients()
	for n := range clients {
		c := clients[n]
		Cmd.Mapping[strings.ToLower(c.Class)] = func(string) int {
			wid := c.WID
			ewmh.ActiveWindowReq(X, wid)
			return 0
		}
	}
}

func (Cmd *TasksCommand) Test(line string) bool {
	if strings.HasPrefix(line, "switch ") {
		line = strings.Split(line, " ")[1]
	}
	_, ok := Cmd.Mapping[strings.ToLower(line)]
	return ok
}

func (Cmd *TasksCommand) Exec(line string) int {
	if strings.HasPrefix(line, "switch ") {
		line = strings.Split(line, " ")[1]
	}
	return Cmd.Mapping[strings.ToLower(line)](line)
}

func (Cmd *TasksCommand) GetText(line string) Line {
	if strings.TrimSpace(line) != "switch" {
		return Line{fmt.Sprintf("Switch to %s", line), []HighlightRule{HighlightRule{10, len(line), "foreground", "bold"}}}
	}
	return Line{fmt.Sprintf("Switch task..."), []HighlightRule{}}
}

func (Cmd *TasksCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	if strings.HasPrefix(line, "switch ") {
		line = strings.Split(line, " ")[1]
	}
	log.Println(line)
	for c := range Cmd.Mapping {
		if c == "switch " && strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(c, line) {
			s = append(s, AutocompleteItem{Cmd, c})
		}
	}
	return s
}
