package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yarlson/tap"
)

func main() {
	fmt.Println("Styled MultiSelect Example")
	fmt.Println("Use arrow keys (or hjkl) to navigate, Space to toggle, Enter to submit, Ctrl+C to cancel")
	fmt.Println()

	// Example 1: Choose multiple languages
	langs := []tap.SelectOption[string]{
		{Value: "go", Label: "Go", Hint: "Fast, simple, great for CLIs"},
		{Value: "rust", Label: "Rust", Hint: "Safety and performance"},
		{Value: "python", Label: "Python", Hint: "Great ecosystem and DX"},
		{Value: "js", Label: "JavaScript", Hint: "Ubiquitous on the web"},
	}

	selected := tap.MultiSelect[string](context.Background(), tap.MultiSelectOptions[string]{
		Message: "Which languages are you using this year?",
		Options: langs,
	})
	fmt.Printf("\nYou selected: %v\n", selected)

	// Example 2: With initial values
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Frameworks (pre-selected) Example:")

	frameworks := []tap.SelectOption[string]{
		{Value: "react", Label: "React"},
		{Value: "vue", Label: "Vue"},
		{Value: "svelte", Label: "Svelte"},
		{Value: "angular", Label: "Angular"},
	}
	initial := []string{"react", "svelte"}

	selected2 := tap.MultiSelect[string](context.Background(), tap.MultiSelectOptions[string]{
		Message:       "Select the frontend frameworks you know:",
		Options:       frameworks,
		InitialValues: initial,
	})
	fmt.Printf("\nYou chose: %v\n", selected2)

	fmt.Println("\nAll examples completed successfully! 🎉")
}
