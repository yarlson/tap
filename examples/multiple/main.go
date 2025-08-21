package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/yarlson/tap/pkg/core"
	"github.com/yarlson/tap/pkg/prompts"
	"github.com/yarlson/tap/pkg/terminal"
)

func getProjectLabel(projectType string) string {
	labels := map[string]string{
		"web":     "Web Application",
		"mobile":  "Mobile App",
		"desktop": "Desktop Application",
		"api":     "API/Backend Service",
		"game":    "Game Development",
		"data":    "Data Science/ML",
	}
	if label, exists := labels[projectType]; exists {
		return label
	}
	return projectType
}

func main() {
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("init terminal: %v\r\n", err)
		return
	}
	defer term.Close()

	// Intro using messages primitive
	prompts.Intro("🚀 Multiple Prompts Example", prompts.MessageOptions{Output: term.Writer})

	// First prompt: Get user's name
	nameRes := prompts.Text(prompts.TextOptions{
		Message:     "What's your name?",
		Placeholder: "Enter your name...",
		Input:       term.Reader,
		Output:      term.Writer,
	})

	if core.IsCancel(nameRes) {
		prompts.Cancel("Operation canceled.", prompts.MessageOptions{Output: term.Writer})
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
		prompts.Cancel("Operation canceled.", prompts.MessageOptions{Output: term.Writer})
		return
	}

	language, ok := langRes.(string)
	if !ok {
		fmt.Printf("Unexpected language result: %#v\r\n", langRes)
		return
	}

	// Third prompt: Select project type
	projectTypes := []prompts.SelectOption[string]{
		{Value: "web", Label: "Web Application", Hint: "Frontend and backend web development"},
		{Value: "mobile", Label: "Mobile App", Hint: "iOS and Android applications"},
		{Value: "desktop", Label: "Desktop Application", Hint: "Cross-platform desktop software"},
		{Value: "api", Label: "API/Backend Service", Hint: "REST APIs and microservices"},
		{Value: "game", Label: "Game Development", Hint: "Video games and interactive media"},
		{Value: "data", Label: "Data Science/ML", Hint: "Analytics, machine learning, AI"},
	}

	projectRes := prompts.Select(prompts.SelectOptions[string]{
		Message: fmt.Sprintf("What type of %s projects do you work on?", language),
		Options: projectTypes,
		Input:   term.Reader,
		Output:  term.Writer,
	})

	if core.IsCancel(projectRes) {
		prompts.Cancel("Operation canceled.", prompts.MessageOptions{Output: term.Writer})
		return
	}

	projectType, ok := projectRes.(string)
	if !ok {
		fmt.Printf("Unexpected project type result: %#v\r\n", projectRes)
		return
	}

	// Fourth prompt: Get years of experience
	expRes := prompts.Text(prompts.TextOptions{
		Message:      "How many years of experience do you have with " + language + "?",
		Placeholder:  "Enter number of years...",
		DefaultValue: "1",
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(expRes) {
		prompts.Cancel("Operation canceled.", prompts.MessageOptions{Output: term.Writer})
		return
	}

	experience, ok := expRes.(string)
	if !ok {
		fmt.Printf("Unexpected experience result: %#v\r\n", expRes)
		return
	}

	// Fifth prompt: Confirm if they want to see a summary
	confirmRes := prompts.Confirm(prompts.ConfirmOptions{
		Message:      "Would you like to see a summary of your information?",
		InitialValue: true,
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(confirmRes) {
		prompts.Cancel("Operation canceled.", prompts.MessageOptions{Output: term.Writer})
		return
	}

	confirmed, ok := confirmRes.(bool)
	if !ok {
		fmt.Printf("Unexpected confirmation result: %#v\r\n", confirmRes)
		return
	}

	var detailed bool
	if !confirmed {
		prompts.Outro(fmt.Sprintf("No problem! Thanks for trying the example, %s! 👋", name), prompts.MessageOptions{Output: term.Writer})
		return
	}

	// Sixth prompt: If confirmed, ask for final message preference
	styleRes := prompts.Confirm(prompts.ConfirmOptions{
		Message:      "Display summary in detailed format?",
		Active:       "Detailed",
		Inactive:     "Brief",
		InitialValue: false,
		Input:        term.Reader,
		Output:       term.Writer,
	})

	if core.IsCancel(styleRes) {
		prompts.Cancel("Operation canceled.", prompts.MessageOptions{Output: term.Writer})
		return
	}

	detailed, ok = styleRes.(bool)
	if !ok {
		fmt.Printf("Unexpected style result: %#v\r\n", styleRes)
		return
	}

	// Show progress bar while generating the summary
	profileProgress := prompts.NewProgress(prompts.ProgressOptions{
		Style:  "heavy",
		Max:    100,
		Size:   45,
		Output: term.Writer,
	})

	profileProgress.Start("Analyzing your responses...")
	time.Sleep(1200 * time.Millisecond)

	profileProgress.Advance(25, "Processing your preferences...")
	time.Sleep(800 * time.Millisecond)

	profileProgress.Advance(30, "Generating insights...")
	time.Sleep(800 * time.Millisecond)

	profileProgress.Advance(25, "Formatting summary...")
	time.Sleep(800 * time.Millisecond)

	profileProgress.Advance(20, "Finalizing report...")
	time.Sleep(600 * time.Millisecond)

	profileProgress.Stop("Profile summary ready! 📋", 0)

	// Display the summary after the progress completes
	fmt.Print("\r\n" + strings.Repeat("=", 50) + "\r\n")
	fmt.Printf("📋 PROFILE SUMMARY\r\n")
	fmt.Print(strings.Repeat("=", 50) + "\r\n")

	if detailed {
		fmt.Printf("👤 Name: %s\r\n", name)
		fmt.Printf("💻 Favorite Language: %s\r\n", language)
		fmt.Printf("🚀 Project Type: %s\r\n", getProjectLabel(projectType))
		fmt.Printf("📈 Experience Level: %s years\r\n", experience)
		fmt.Printf("\r\n🎯 Profile Analysis:\r\n")
		if experience == "0" || experience == "1" {
			fmt.Printf("   You're just getting started with %s - keep learning!\r\n", language)
		} else {
			fmt.Printf("   Great! You have solid experience with %s.\r\n", language)
		}
		fmt.Printf("   %s development is a great choice!\r\n", getProjectLabel(projectType))
	} else {
		fmt.Printf("%s • %s • %s • %s years experience\r\n", name, language, getProjectLabel(projectType), experience)
	}

	prompts.Outro("Thanks for trying out the Tap prompts library! 🎉", prompts.MessageOptions{Output: term.Writer})
}
