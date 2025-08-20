package main

import (
	"fmt"

	"github.com/yarlson/glack/pkg/core"
	"github.com/yarlson/glack/pkg/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("init terminal: %v\r\n", err)
		return
	}
	defer term.Close()

	res := core.Confirm(core.ConfirmOptions{
		Message:      "Proceed? (y/n)",
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
