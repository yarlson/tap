package main

import (
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	res := tap.Password(tap.PasswordOptions{
		Message:      "Enter password:",
		DefaultValue: "",
		Validate: func(s string) error {
			if len(s) < 6 {
				return fmt.Errorf("password must be at least 6 characters")
			}
			return nil
		},
	})
	fmt.Printf("Password length: %d\r\n", len(res))
}
