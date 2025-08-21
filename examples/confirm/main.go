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

	res := prompts.Confirm(prompts.ConfirmOptions{
		Message:      "Proceed?",
		InitialValue: true,
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(res) {
		fmt.Printf("Canceled\r\n")
		return
	}

	if b, ok := res.(bool); ok {
		fmt.Printf("Confirmed: %v\r\n", b)
	} else {
		fmt.Printf("Unexpected result: %#v\r\n", res)
	}
}
