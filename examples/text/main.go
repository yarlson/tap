package main

import (
	"context"
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	res := tap.Text(context.Background(), tap.TextOptions{
		Message:      "Enter text:",
		InitialValue: "initial",
		DefaultValue: "anon",
		Placeholder:  "Type something...",
	})

	fmt.Printf("Result: %s\r\n", res)
}
