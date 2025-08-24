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
	name := tap.Text(tap.TextOptions{
		Message:     "What's your name?",
		Placeholder: "Enter your name...",
	})

	// Second prompt: Choose programming languages (multi-select)
	langOptions := []tap.SelectOption[string]{
		{Value: "Go", Label: "Go"},
		{Value: "Python", Label: "Python"},
		{Value: "JavaScript", Label: "JavaScript"},
		{Value: "TypeScript", Label: "TypeScript"},
		{Value: "Rust", Label: "Rust"},
		{Value: "Java", Label: "Java"},
	}
	languages := tap.MultiSelect[string](tap.MultiSelectOptions[string]{
		Message: fmt.Sprintf("Hi %s! Which programming languages do you use?", name),
		Options: langOptions,
	})

	// Third prompt: Select project type
	projectTypes := []tap.SelectOption[string]{
		{Value: "web", Label: "Web Application", Hint: "Frontend and backend web development"},
		{Value: "mobile", Label: "Mobile App", Hint: "iOS and Android applications"},
		{Value: "desktop", Label: "Desktop Application", Hint: "Cross-platform desktop software"},
		{Value: "api", Label: "API/Backend Service", Hint: "REST APIs and microservices"},
		{Value: "game", Label: "Game Development", Hint: "Video games and interactive media"},
		{Value: "data", Label: "Data Science/ML", Hint: "Analytics, machine learning, AI"},
	}

	projectType := tap.Select[string](tap.SelectOptions[string]{
		Message: fmt.Sprintf("What type of projects do you work on with %s?", strings.Join(languages, ", ")),
		Options: projectTypes,
	})

	// Fourth prompt: Get years of experience
	experience := tap.Text(tap.TextOptions{
		Message:      "How many years of experience do you have with your selected languages?",
		Placeholder:  "Enter number of years...",
		DefaultValue: "1",
		Validate: func(s string) error {
			var years int
			_, err := fmt.Sscanf(s, "%d", &years)
			if err != nil || years < 0 {
				return fmt.Errorf("please enter a valid non-negative number")
			}

			return nil
		},
	})

	// Fifth prompt: Confirm if they want to see a summary
	confirmed := tap.Confirm(tap.ConfirmOptions{
		Message:      "Would you like to see a summary of your information?",
		InitialValue: true,
	})

	var detailed bool
	if !confirmed {
		tap.Outro(fmt.Sprintf("No problem! Thanks for trying the example, %s! 👋", name))
		return
	}

	// Sixth prompt: If confirmed, ask for final message preference
	detailed = tap.Confirm(tap.ConfirmOptions{
		Message:      "Display summary in detailed format?",
		Active:       "Detailed",
		Inactive:     "Brief",
		InitialValue: false,
	})

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
	border := tap.GrayBorder
	if detailed {
		border = tap.CyanBorder
	}
	tap.Box(
		func() string {
			if detailed {
				var b strings.Builder
				_, _ = fmt.Fprintf(&b, "👤 Name: %s\n", name)
				_, _ = fmt.Fprintf(&b, "💻 Languages: %s\n", strings.Join(languages, ", "))
				_, _ = fmt.Fprintf(&b, "🚀 Project Type: %s\n", getProjectLabel(projectType))
				_, _ = fmt.Fprintf(&b, "📈 Experience Level: %s years\n", experience)
				_, _ = fmt.Fprintf(&b, "\n🎯 Profile Analysis:\n")
				if experience == "0" || experience == "1" {
					_, _ = fmt.Fprintf(&b, "   You're just getting started - keep learning!\n")
				} else {
					_, _ = fmt.Fprintf(&b, "   Great! You have solid experience.\n")
				}
				_, _ = fmt.Fprintf(&b, "   %s development is a great choice!", getProjectLabel(projectType))
				return b.String()
			}
			return fmt.Sprintf("%s • %s • %s • %s years experience", name, strings.Join(languages, ", "), getProjectLabel(projectType), experience)
		}(),
		"📋 PROFILE SUMMARY",
		tap.BoxOptions{
			Columns:        80,
			WidthFraction:  1.0,
			TitlePadding:   1,
			ContentPadding: 1,
			Rounded:        true,
			IncludePrefix:  true,
			FormatBorder:   border,
		},
	)

	tap.Outro("Thanks for trying out the Tap prompts library! 🎉")
}
