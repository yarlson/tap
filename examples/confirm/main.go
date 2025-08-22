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

	if tap.IsCancel(res) {
		fmt.Printf("Canceled\r\n")
		return
	}

	if b, ok := res.(bool); ok {
		fmt.Printf("Confirmed: %v\r\n", b)
	} else {
		fmt.Printf("Unexpected result: %#v\r\n", res)
	}
}
