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

    if tap.IsCancel(res) {
		fmt.Printf("Canceled\r\n")
		return
	}
	if s, ok := res.(string); ok {
		fmt.Printf("Result: %s\r\n", s)
	} else {
		fmt.Printf("Unexpected result: %#v\r\n", res)
	}
}
