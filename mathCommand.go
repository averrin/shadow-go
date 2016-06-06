package main

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"log"
	"strings"
)

type MathCommand struct {
	Mapping map[string]func(string) int
}

func (Cmd *MathCommand) Init() {
	Cmd.Mapping = map[string]func(string) int{
		"=": func(string) int { return 0 },
	}
}

func (Cmd *MathCommand) Test(line string) bool {
	_, ok := Cmd.Mapping[line]
	return ok
}

func (Cmd *MathCommand) Exec(line string) int {
	return Cmd.Mapping[line](line)
}

func (Cmd *MathCommand) GetText(line string) Line {
	// line[0] = strings.ToUpper(line[0])
	expr := calc(line[1:])
	log.Println(expr)
	if expr != nil {
		return Line{fmt.Sprintf("= %v", expr), []HighlightRule{HighlightRule{2, len(fmt.Sprintf("%v", expr)), "foreground", "bold"}}}
	}
	if strings.TrimSpace(line) != "=" {
		return Line{fmt.Sprintf("Wrong expression"), []HighlightRule{HighlightRule{0, -1, "red", "default"}}}
	} else {
		return Line{fmt.Sprintf("Type expression to calc"), []HighlightRule{}}
	}
}

func (Cmd *MathCommand) GetSuggests(line string) []AutocompleteItem {
	s := []AutocompleteItem{}
	if strings.HasPrefix(line, "=") {
		s = append(s, AutocompleteItem{Cmd, line})
	}
	return s
}

func calc(expr string) constant.Value {
	fs := token.NewFileSet()
	// tr, _ := parser.ParseExpr(expr)
	tv, err := types.Eval(fs, nil, token.NoPos, expr)
	log.Println(tv.Value, err)
	if err == nil {
		return tv.Value
	}
	ZERO := new(constant.Value)
	return *ZERO
}
