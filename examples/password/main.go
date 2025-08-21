package main

import (
	"fmt"

	"github.com/yarlson/tap/core"
	"github.com/yarlson/tap/prompts"
	"github.com/yarlson/tap/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("init terminal: %v\r\n", err)
		return
	}
	defer term.Close()

	res := prompts.Password(prompts.PasswordOptions{
		Message:      "Enter password:",
		DefaultValue: "",
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(res) {
		fmt.Printf("Canceled\r\n")
		return
	}
	if s, ok := res.(string); ok {
		fmt.Printf("Password length: %d\r\n", len(s))
	} else {
		fmt.Printf("Unexpected result: %#v\r\n", res)
	}
}
