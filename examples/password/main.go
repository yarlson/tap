package main

import (
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	res := tap.Password(tap.PasswordOptions{
		Message:      "Enter password:",
		DefaultValue: "",
	})

	if tap.IsCancel(res) {
		fmt.Printf("Canceled\r\n")
		return
	}
	if s, ok := res.(string); ok {
		fmt.Printf("Password length: %d\r\n", len(s))
	} else {
		fmt.Printf("Unexpected result: %#v\r\n", res)
	}
}
