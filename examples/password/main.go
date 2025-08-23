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
	fmt.Printf("Password length: %d\r\n", len(res))
}
