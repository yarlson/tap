package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yarlson/tap/pkg/core"
	"github.com/yarlson/tap/pkg/terminal"
)

func main() {
	// Initialize terminal
	t, err := terminal.New()
	if err != nil {
		fmt.Printf("Failed to initialize terminal: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()

	fmt.Println("Select Example - Unstyled Core Implementation")
	fmt.Println("Use arrow keys (or hjkl) to navigate, Enter to select, Ctrl+C to cancel")
	fmt.Println()

	// Create options for different colors
	colors := []core.SelectOption[string]{
		{Value: "red", Label: "Red", Hint: "The color of passion"},
		{Value: "blue", Label: "Blue", Hint: "The color of the sky"},
		{Value: "green", Label: "Green", Hint: "The color of nature"},
		{Value: "yellow", Label: "Yellow", Hint: "The color of sunshine"},
		{Value: "purple", Label: "Purple", Hint: "The color of royalty"},
	}

	// Run the select prompt
	result := core.Select(core.SelectOptions[string]{
		Message: "Choose your favorite color:",
		Options: colors,
		Input:   t.Reader,
		Output:  t.Writer,
	})

	// Handle the result
	if core.IsCancel(result) {
		fmt.Println("Selection cancelled.")
		os.Exit(1)
	}

	selectedColor, ok := result.(string)
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("\nYou selected: %s\n", selectedColor)

	// Example with initial value
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Second example with initial value:")

	initialValue := "blue"
	result2 := core.Select(core.SelectOptions[string]{
		Message:      "Choose again (starting with blue):",
		Options:      colors,
		InitialValue: &initialValue,
		Input:        t.Reader,
		Output:       t.Writer,
	})

	if core.IsCancel(result2) {
		fmt.Println("Selection cancelled.")
		os.Exit(1)
	}

	selectedColor2, ok := result2.(string)
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("\nYou selected: %s\n", selectedColor2)

	// Example with numbers
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Third example with numbers:")

	numbers := []core.SelectOption[int]{
		{Value: 1, Label: "One"},
		{Value: 2, Label: "Two"},
		{Value: 3, Label: "Three"},
		{Value: 42, Label: "Forty-two", Hint: "The answer to everything"},
	}

	result3 := core.Select(core.SelectOptions[int]{
		Message: "Pick a number:",
		Options: numbers,
		Input:   t.Reader,
		Output:  t.Writer,
	})

	if core.IsCancel(result3) {
		fmt.Println("Selection cancelled.")
		os.Exit(1)
	}

	selectedNumber, ok := result3.(int)
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("\nYou selected: %d\n", selectedNumber)
}
