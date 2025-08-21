package main

import (
	"fmt"

	"github.com/yarlson/tap/pkg/core"
	"github.com/yarlson/tap/pkg/prompts"
	"github.com/yarlson/tap/pkg/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("init terminal: %v\r\n", err)
		return
	}
	defer term.Close()

	res := prompts.Text(prompts.TextOptions{
		Message:      "Enter text:",
		InitialValue: "initial",
		DefaultValue: "anon",
		Placeholder:  "Type something...",
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(res) {
		fmt.Printf("Canceled\r\n")
		return
	}
	if s, ok := res.(string); ok {
		fmt.Printf("Result: %s\r\n", s)
	} else {
		fmt.Printf("Unexpected result: %#v\r\n", res)
	}
}
