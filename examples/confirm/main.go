package main

import (
	"fmt"

	core "github.com/yarlson/glack/pkg/core"
)

func main() {
	res := core.Confirm(core.ConfirmOptions{
		Message:      "Proceed? (y/n)",
		InitialValue: true,
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
