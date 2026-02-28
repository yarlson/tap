package main

import (
	"context"
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	tap.Intro("Textarea Example")

	res := tap.Textarea(context.Background(), tap.TextareaOptions{
		Message:     "Enter your commit message (Shift+Enter for new line):",
		Placeholder: "Type something...",
		Validate: func(s string) error {
			if len(s) < 10 {
				return fmt.Errorf("at least 10 characters required")
			}
			return nil
		},
	})

	tap.Message(fmt.Sprintf("You entered: %s", res))
	tap.Outro("Done!")
}
