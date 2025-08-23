package main

import (
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	res := tap.Confirm(tap.ConfirmOptions{
		Message:      "Proceed?",
		InitialValue: true,
	})
	fmt.Printf("Confirmed: %v\r\n", res)
}
