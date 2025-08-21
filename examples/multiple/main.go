package main

import (
	"fmt"
	"strings"

	"github.com/yarlson/tap/pkg/core"
	"github.com/yarlson/tap/pkg/prompts"
	"github.com/yarlson/tap/pkg/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("init terminal: %v\r\n", err)
		return
	}
	defer term.Close()

	// Welcome message
	fmt.Printf("🚀 Welcome to the Multiple Prompts Example!\r\n")
	fmt.Printf("This demonstrates a sequence of prompts building on each other.\r\n\r\n")

	// First prompt: Get user's name
	nameRes := prompts.Text(prompts.TextOptions{
		Message:     "What's your name?",
		Placeholder: "Enter your name...",
		Input:       term.Reader,
		Output:      term.Writer,
	})

	if core.IsCancel(nameRes) {
		fmt.Printf("Operation canceled.\r\n")
		return
	}

	name, ok := nameRes.(string)
	if !ok {
		fmt.Printf("Unexpected name result: %#v\r\n", nameRes)
		return
	}

	// Second prompt: Get user's favorite programming language
	langRes := prompts.Text(prompts.TextOptions{
		Message:      fmt.Sprintf("Hi %s! What's your favorite programming language?", name),
		Placeholder:  "e.g., Go, Python, JavaScript...",
		DefaultValue: "Go",
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(langRes) {
		fmt.Printf("Operation canceled.\r\n")
		return
	}

	language, ok := langRes.(string)
	if !ok {
		fmt.Printf("Unexpected language result: %#v\r\n", langRes)
		return
	}

	// Third prompt: Get years of experience
	expRes := prompts.Text(prompts.TextOptions{
		Message:      "How many years of experience do you have with " + language + "?",
		Placeholder:  "Enter number of years...",
		DefaultValue: "1",
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(expRes) {
		fmt.Printf("Operation canceled.\r\n")
		return
	}

	experience, ok := expRes.(string)
	if !ok {
		fmt.Printf("Unexpected experience result: %#v\r\n", expRes)
		return
	}

	// Fourth prompt: Confirm if they want to see a summary
	confirmRes := prompts.Confirm(prompts.ConfirmOptions{
		Message:      "Would you like to see a summary of your information?",
		InitialValue: true,
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(confirmRes) {
		fmt.Printf("Operation canceled.\r\n")
		return
	}

	confirmed, ok := confirmRes.(bool)
	if !ok {
		fmt.Printf("Unexpected confirmation result: %#v\r\n", confirmRes)
		return
	}

	var detailed bool
	if !confirmed {
		fmt.Printf("\r\nNo problem! Thanks for trying the example, %s! 👋\r\n", name)
		return
	}

	// Final prompt: If confirmed, ask for final message preference
	styleRes := prompts.Confirm(prompts.ConfirmOptions{
		Message:      "Display summary in detailed format?",
		Active:       "Detailed",
		Inactive:     "Brief",
		InitialValue: false,
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(styleRes) {
		fmt.Printf("Operation canceled.\r\n")
		return
	}

	detailed, ok = styleRes.(bool)
	if !ok {
		fmt.Printf("Unexpected style result: %#v\r\n", styleRes)
		return
	}

	// Display the summary
	fmt.Print("\r\n" + strings.Repeat("=", 50) + "\r\n")
	fmt.Printf("📋 PROFILE SUMMARY\r\n")
	fmt.Print(strings.Repeat("=", 50) + "\r\n")

	if detailed {
		fmt.Printf("👤 Name: %s\r\n", name)
		fmt.Printf("💻 Favorite Language: %s\r\n", language)
		fmt.Printf("📈 Experience Level: %s years\r\n", experience)
		fmt.Printf("\r\n🎯 Profile Analysis:\r\n")
		if experience == "0" || experience == "1" {
			fmt.Printf("   You're just getting started with %s - keep learning!\r\n", language)
		} else {
			fmt.Printf("   Great! You have solid experience with %s.\r\n", language)
		}
	} else {
		fmt.Printf("%s • %s • %s years experience\r\n", name, language, experience)
	}

	fmt.Print(strings.Repeat("=", 50) + "\r\n")
	fmt.Printf("Thanks for trying out the Tap prompts library! 🎉\r\n")
}
