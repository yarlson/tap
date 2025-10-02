package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yarlson/tap"
)

func suggest(list []string) func(string) []string {
	return func(input string) []string {
		if input == "" {
			return list
		}

		low := strings.ToLower(input)

		var out []string

		for _, s := range list {
			if strings.Contains(strings.ToLower(s), low) {
				out = append(out, s)
			}
		}

		return out
	}
}

func main() {
	tap.Intro("🔎 Autocomplete Example")

	langs := []string{
		"Go", "Golang", "Python", "Rust", "Java", "JavaScript", "TypeScript", "Ruby", "Kotlin", "Swift",
	}

	res := tap.Autocomplete(context.Background(), tap.AutocompleteOptions{
		Message:     "Search language:",
		Placeholder: "Start typing...",
		Suggest:     suggest(langs),
		MaxResults:  6,
	})

	tap.Message(fmt.Sprintf("You selected: %s", res))
	tap.Outro("Thanks for trying autocomplete! ✨")
}
