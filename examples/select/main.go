package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yarlson/tap/core"
	"github.com/yarlson/tap/prompts"
	"github.com/yarlson/tap/terminal"
)

func main() {
	// Initialize terminal
	t, err := terminal.New()
	if err != nil {
		fmt.Printf("Failed to initialize terminal: %v\n", err)
		os.Exit(1)
	}
	defer t.Close()

	fmt.Println("Styled Select Example")
	fmt.Println("Use arrow keys (or hjkl) to navigate, Enter to select, Ctrl+C to cancel")
	fmt.Println()

	// Example 1: Color selection with hints
	colors := []prompts.SelectOption[string]{
		{Value: "red", Label: "Red", Hint: "The color of passion and energy"},
		{Value: "blue", Label: "Blue", Hint: "The color of calm and trust"},
		{Value: "green", Label: "Green", Hint: "The color of nature and growth"},
		{Value: "yellow", Label: "Yellow", Hint: "The color of happiness and optimism"},
		{Value: "purple", Label: "Purple", Hint: "The color of creativity and mystery"},
	}

	result := prompts.Select(prompts.SelectOptions[string]{
		Message: "What's your favorite color?",
		Options: colors,
		Input:   t.Reader,
		Output:  t.Writer,
	})

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

	// Example 2: Framework selection with initial value
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Framework Selection Example:")

	frameworks := []prompts.SelectOption[string]{
		{Value: "react", Label: "React", Hint: "A JavaScript library for building user interfaces"},
		{Value: "vue", Label: "Vue.js", Hint: "The Progressive JavaScript Framework"},
		{Value: "angular", Label: "Angular", Hint: "Platform for building mobile and desktop web apps"},
		{Value: "svelte", Label: "Svelte", Hint: "Cybernetically enhanced web apps"},
		{Value: "solid", Label: "SolidJS", Hint: "Simple and performant reactivity"},
	}

	initialValue := "react"
	result2 := prompts.Select(prompts.SelectOptions[string]{
		Message:      "Which frontend framework do you prefer?",
		Options:      frameworks,
		InitialValue: &initialValue,
		Input:        t.Reader,
		Output:       t.Writer,
	})

	if core.IsCancel(result2) {
		fmt.Println("Selection cancelled.")
		os.Exit(1)
	}

	selectedFramework, ok := result2.(string)
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("\nYou chose: %s\n", selectedFramework)

	// Example 3: Priority levels (numeric values)
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Priority Selection Example:")

	priorities := []prompts.SelectOption[int]{
		{Value: 1, Label: "Low Priority", Hint: "Can be done when time permits"},
		{Value: 2, Label: "Medium Priority", Hint: "Should be completed this week"},
		{Value: 3, Label: "High Priority", Hint: "Needs attention today"},
		{Value: 4, Label: "Critical", Hint: "Drop everything and do this now"},
	}

	result3 := prompts.Select(prompts.SelectOptions[int]{
		Message: "What's the priority level for this task?",
		Options: priorities,
		Input:   t.Reader,
		Output:  t.Writer,
	})

	if core.IsCancel(result3) {
		fmt.Println("Selection cancelled.")
		os.Exit(1)
	}

	selectedPriority, ok := result3.(int)
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("\nSelected priority level: %d\n", selectedPriority)

	// Example 4: Simple options without labels (uses values as labels)
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Simple Options Example:")

	environments := []prompts.SelectOption[string]{
		{Value: "development"},
		{Value: "staging"},
		{Value: "production"},
	}

	result4 := prompts.Select(prompts.SelectOptions[string]{
		Message: "Which environment to deploy to?",
		Options: environments,
		Input:   t.Reader,
		Output:  t.Writer,
	})

	if core.IsCancel(result4) {
		fmt.Println("Selection cancelled.")
		os.Exit(1)
	}

	selectedEnv, ok := result4.(string)
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("\nDeploying to: %s\n", selectedEnv)
	fmt.Println("\nAll examples completed successfully! 🎉")
}
