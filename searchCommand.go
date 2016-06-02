package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type SearchCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *SearchCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"g":  searchInGoogle,
		"w":  searchInWiki,
		"gh": searchInGH,
		"m":  searchInMusic,
	}
}

func (Cmd *SearchCommand) getPrefix(line string) string {
	for p := range Cmd.Mapping {
		if strings.HasPrefix(line, p+" ") {
			return p
		}
	}
	return ""
}

func (Cmd *SearchCommand) Test(line string) bool {
	return Cmd.getPrefix(line) != ""
}

func (Cmd *SearchCommand) Exec(line string) int {
	p := Cmd.getPrefix(line)
	return Cmd.Mapping[p](line[len(p)+1:])
}

func (Cmd *SearchCommand) GetText(line string) string {
	p := Cmd.getPrefix(line)
	mapping := map[string]string{
		"g":  "Google",
		"w":  "Wiki",
		"gh": "GitHub",
		"m":  "Google Music",
	}
	place := mapping[p]
	return fmt.Sprintf("Search in %s: %s", place, line[len(p)+1:])
}

func searchInGoogle(q string) int {
	url := fmt.Sprintf("https://www.google.com/search?q=%s", q)
	return openURL(url)
}

func searchInWiki(q string) int {
	url := fmt.Sprintf("https://en.wikipedia.org/wiki/Special:Search/%s", q)
	return openURL(url)
}

func searchInGH(q string) int {
	url := fmt.Sprintf("https://github.com/search?utf8=✓&q=%s", q)
	return openURL(url)
}

func searchInMusic(q string) int {
	url := fmt.Sprintf("https://play.google.com/music/listen#/sr/%s", q)
	return openURL(url)
}

func openURL(url string) int {
	c := exec.Command("xdg-open", url)
	err := c.Start()
	if err != nil {
		return 1
	}
	return 0
}
