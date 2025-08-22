package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/yarlson/tap"
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
	tap.Intro("🚀 Multiple Prompts Example")

	// First prompt: Get user's name
	nameRes := tap.Text(tap.TextOptions{
		Message:     "What's your name?",
		Placeholder: "Enter your name...",
	})

	if tap.IsCancel(nameRes) {
		tap.Cancel("Operation canceled.")
		return
	}

	name, ok := nameRes.(string)
	if !ok {
		fmt.Printf("Unexpected name result: %#v\r\n", nameRes)
		return
	}

	// Second prompt: Get user's favorite programming language
	langRes := tap.Text(tap.TextOptions{
		Message:      fmt.Sprintf("Hi %s! What's your favorite programming language?", name),
		Placeholder:  "e.g., Go, Python, JavaScript...",
		DefaultValue: "Go",
	})

	if tap.IsCancel(langRes) {
		tap.Cancel("Operation canceled.")
		return
	}

	language, ok := langRes.(string)
	if !ok {
		fmt.Printf("Unexpected language result: %#v\r\n", langRes)
		return
	}

	// Third prompt: Select project type
	projectTypes := []tap.SelectOption[string]{
		{Value: "web", Label: "Web Application", Hint: "Frontend and backend web development"},
		{Value: "mobile", Label: "Mobile App", Hint: "iOS and Android applications"},
		{Value: "desktop", Label: "Desktop Application", Hint: "Cross-platform desktop software"},
		{Value: "api", Label: "API/Backend Service", Hint: "REST APIs and microservices"},
		{Value: "game", Label: "Game Development", Hint: "Video games and interactive media"},
		{Value: "data", Label: "Data Science/ML", Hint: "Analytics, machine learning, AI"},
	}

	projectRes := tap.Select(tap.SelectOptions[string]{
		Message: fmt.Sprintf("What type of %s projects do you work on?", language),
		Options: projectTypes,
	})

	if tap.IsCancel(projectRes) {
		tap.Cancel("Operation canceled.")
		return
	}

	projectType, ok := projectRes.(string)
	if !ok {
		fmt.Printf("Unexpected project type result: %#v\r\n", projectRes)
		return
	}

	// Fourth prompt: Get years of experience
	expRes := tap.Text(tap.TextOptions{
		Message:      "How many years of experience do you have with " + language + "?",
		Placeholder:  "Enter number of years...",
		DefaultValue: "1",
	})

	if tap.IsCancel(expRes) {
		tap.Cancel("Operation canceled.")
		return
	}

	experience, ok := expRes.(string)
	if !ok {
		fmt.Printf("Unexpected experience result: %#v\r\n", expRes)
		return
	}

	// Fifth prompt: Confirm if they want to see a summary
	confirmRes := tap.Confirm(tap.ConfirmOptions{
		Message:      "Would you like to see a summary of your information?",
		InitialValue: true,
	})

	if tap.IsCancel(confirmRes) {
		tap.Cancel("Operation canceled.")
		return
	}

	confirmed, ok := confirmRes.(bool)
	if !ok {
		fmt.Printf("Unexpected confirmation result: %#v\r\n", confirmRes)
		return
	}

	var detailed bool
	if !confirmed {
		tap.Outro(fmt.Sprintf("No problem! Thanks for trying the example, %s! 👋", name))
		return
	}

	// Sixth prompt: If confirmed, ask for final message preference
	styleRes := tap.Confirm(tap.ConfirmOptions{
		Message:      "Display summary in detailed format?",
		Active:       "Detailed",
		Inactive:     "Brief",
		InitialValue: false,
	})

	if tap.IsCancel(styleRes) {
		tap.Cancel("Operation canceled.")
		return
	}

	detailed, ok = styleRes.(bool)
	if !ok {
		fmt.Printf("Unexpected style result: %#v\r\n", styleRes)
		return
	}

	// Show progress bar while generating the summary
	profileProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "heavy",
		Max:   100,
		Size:  45,
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

	// Display the summary after the progress completes using a box
	tap.Box(
		func() string {
			if detailed {
				var b strings.Builder
				_, _ = fmt.Fprintf(&b, "👤 Name: %s\n", name)
				_, _ = fmt.Fprintf(&b, "💻 Favorite Language: %s\n", language)
				_, _ = fmt.Fprintf(&b, "🚀 Project Type: %s\n", getProjectLabel(projectType))
				_, _ = fmt.Fprintf(&b, "📈 Experience Level: %s years\n", experience)
				_, _ = fmt.Fprintf(&b, "\n🎯 Profile Analysis:\n")
				if experience == "0" || experience == "1" {
					_, _ = fmt.Fprintf(&b, "   You're just getting started with %s - keep learning!\n", language)
				} else {
					_, _ = fmt.Fprintf(&b, "   Great! You have solid experience with %s.\n", language)
				}
				_, _ = fmt.Fprintf(&b, "   %s development is a great choice!", getProjectLabel(projectType))
				return b.String()
			}
			return fmt.Sprintf("%s • %s • %s • %s years experience", name, language, getProjectLabel(projectType), experience)
		}(),
		"📋 PROFILE SUMMARY",
		tap.BoxOptions{
			Columns:        80,
			WidthFraction:  1.0,
			TitlePadding:   1,
			ContentPadding: 1,
			Rounded:        true,
			IncludePrefix:  true,
			FormatBorder:   tap.GrayBorder,
		},
	)

	tap.Outro("Thanks for trying out the Tap prompts library! 🎉")
}
