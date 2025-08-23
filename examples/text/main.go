package main

import (
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	res := tap.Text(tap.TextOptions{
		Message:      "Enter text:",
		InitialValue: "initial",
		DefaultValue: "anon",
		Placeholder:  "Type something...",
	})

	fmt.Printf("Result: %s\r\n", res)
}
