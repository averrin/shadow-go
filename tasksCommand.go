package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
)

type TasksCommand struct {
	Mapping map[string]func(string) int
	Clients []Client
}

func (Cmd *TasksCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"switch ": func(string) int { return 0 },
		"x ":      func(string) int { return 0 },
	}
	clients := GetClients()
	for _, client := range clients {
		Cmd.Mapping[strings.ToLower(client.Class)] = Cmd.task
	}
	Cmd.Clients = clients
}

func (Cmd *TasksCommand) task(line string) int {
	// wid := client.WID
	tokens := strings.Split(strings.ToLower(line), " ")
	name := tokens[len(tokens)-1]
	var client Client
	for _, client = range Cmd.Clients {
		if strings.ToLower(client.Class) == name {
			break
		}
	}
	if strings.ToLower(client.Class) != name {
		return 1
	}
	if strings.HasPrefix(line, "x ") {
		ewmh.CloseWindow(X, client.WID)
		return 0
	}
	ewmh.ActiveWindowReq(X, client.WID)
	return 0
}

func (Cmd *TasksCommand) Test(line string) bool {
	if strings.HasPrefix(line, "switch ") {
		line = strings.Split(line, " ")[1]
	}
	if strings.HasPrefix(line, "x ") {
		line = strings.Split(line, " ")[1]
	}
	_, ok := Cmd.Mapping[strings.ToLower(line)]
	return ok
}

func (Cmd *TasksCommand) Exec(line string) int {
	var task string
	if strings.HasPrefix(line, "switch ") {
		task = strings.Split(line, " ")[1]
	} else if strings.HasPrefix(line, "x ") {
		task = strings.Split(line, " ")[1]
	} else {
		task = line
	}
	log.Println(task, "||", line)
	return Cmd.Mapping[strings.ToLower(task)](line)
}

func (Cmd *TasksCommand) GetText(line string) Line {
	if strings.TrimSpace(line) != "switch" && strings.TrimSpace(line) != "x x" {
		if strings.HasPrefix(line, "x ") {
			return Line{fmt.Sprintf("Close window %s", line[2:]), []HighlightRule{HighlightRule{13, -1, "foreground", "bold"}}}
		}
		return Line{fmt.Sprintf("Switch to %s", line), []HighlightRule{HighlightRule{10, -1, "foreground", "bold"}}}
	}
	if strings.TrimSpace(line) == "switch" {
		return Line{fmt.Sprintf("Switch task..."), []HighlightRule{}}
	}
	return Line{fmt.Sprintf("Kill task..."), []HighlightRule{}}
}

func (Cmd *TasksCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	var q string
	if strings.HasPrefix(line, "switch ") {
		q = strings.Split(line, " ")[1]
	}
	if strings.HasPrefix(line, "x ") {
		q = strings.Split(line, " ")[1]
	}
	if q == "" {
		q = line
	}
	for c := range Cmd.Mapping {
		if (c == "switch " || c == "x ") && strings.TrimSpace(q) == "" {
			continue
		}
		if strings.HasPrefix(c, q) {
			if strings.HasPrefix(line, "x ") {
				s = append(s, AutocompleteItem{Cmd, "x " + c})
			} else {
				s = append(s, AutocompleteItem{Cmd, c})
			}
		}
	}
	log.Println(line, q, s)
	return s
}
