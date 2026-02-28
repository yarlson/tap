package main

import (
	"context"
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	tap.Intro("Textarea Example")

	res := tap.Textarea(context.Background(), tap.TextareaOptions{
		Message:     "Enter your message:",
		Placeholder: "Type something...",
	})

	tap.Message(fmt.Sprintf("You entered: %s", res))
	tap.Outro("Done!")
}
