package main

import (
	"context"
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	res := tap.Confirm(context.Background(), tap.ConfirmOptions{
		Message:      "Proceed?",
		InitialValue: true,
	})
	fmt.Printf("Confirmed: %v\r\n", res)
}
