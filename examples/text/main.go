package main

import (
	"fmt"

	core "github.com/yarlson/glack/pkg/core"
)

func main() {
	res := core.Text(core.TextOptions{
		Message:      "Enter text:",
		DefaultValue: "anon",
		Input:        nil,
		Output:       nil,
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
