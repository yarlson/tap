package main

import (
	"github.com/yarlson/tap"
)

func main() {
	// Section 1: Success flow
	tap.Intro("Messages Example")

	// Message without hint
	tap.Message("Task completed successfully")

	// Message with hint
	tap.Message("Files processed", tap.MessageOptions{
		Hint: "42 files in 1.2s",
	})

	// Another message with hint
	tap.Message("Database migrated", tap.MessageOptions{
		Hint: "Applied 3 migrations",
	})

	// Outro with hint
	tap.Outro("All done!", tap.MessageOptions{
		Hint: "Run 'app --help' for more options",
	})

	// Section 2: Intro with hint and Cancel
	tap.Intro("Deploy to Production", tap.MessageOptions{
		Hint: "v2.1.0 -> us-east-1",
	})

	tap.Message("Building application", tap.MessageOptions{
		Hint: "Compiled in 12.4s",
	})

	// Cancel with hint
	tap.Cancel("Deployment aborted", tap.MessageOptions{
		Hint: "No changes were made",
	})
}
