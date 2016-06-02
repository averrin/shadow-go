package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
)

type TasksCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *TasksCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{}
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
	_, ok := Cmd.Mapping[strings.ToLower(line)]
	return ok
}

func (Cmd *TasksCommand) Exec(line string) int {
	return Cmd.Mapping[strings.ToLower(line)](line)
}

func (Cmd *TasksCommand) GetText(line string) string {
	return fmt.Sprintf("Switch to %s", line)
}

func (Cmd *TasksCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	for c := range Cmd.Mapping {
		if strings.HasPrefix(c, line) {
			s = append(s, AutocompleteItem{Cmd, c})
		}
	}
	return s
}
