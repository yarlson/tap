package main

import (
	"fmt"
	"time"

	"github.com/yarlson/tap"
)

func main() {
	tap.Intro("Styled Progress Bar Example")
	// Example 1: File download simulation (heavy style)
	tap.Message("Example 1: File Download Simulation (Heavy Style)")

	downloadProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "heavy",
		Max:   100,
		Size:  40,
	})

	downloadProgress.Start("Downloading file...")

	// Simulate download progress
	for i := 0; i <= 100; i += 10 {
		time.Sleep(200 * time.Millisecond)

		msg := fmt.Sprintf("Downloading file... %d%%", i)
		downloadProgress.Advance(10, msg)
	}

	downloadProgress.Stop("Download complete!", 0)

	// Example 2: Data processing (block style)
	tap.Message("Example 2: Data Processing (Block Style)")

	processProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "block",
		Max:   50,
		Size:  30,
	})

	processProgress.Start("Processing data...")

	// Simulate data processing with different batch sizes
	steps := []int{5, 8, 12, 7, 10, 8}
	messages := []string{
		"Loading dataset...",
		"Preprocessing...",
		"Training model...",
		"Validating...",
		"Optimizing...",
		"Finalizing...",
	}

	for i, step := range steps {
		time.Sleep(300 * time.Millisecond)

		if i < len(messages) {
			processProgress.Advance(step, messages[i])
		} else {
			processProgress.Advance(step, "Processing...")
		}
	}

	processProgress.Stop("Processing complete!", 0)

	// Example 3: Installation progress (light style)
	tap.Message("Example 3: Package Installation (Light Style)")

	installProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "light",
		Max:   20,
		Size:  50,
	})

	installProgress.Start("Installing packages...")

	packages := []string{
		"Installing core dependencies...",
		"Installing development tools...",
		"Installing optional packages...",
		"Configuring environment...",
		"Running post-install scripts...",
	}

	for _, pkg := range packages {
		time.Sleep(400 * time.Millisecond)
		installProgress.Advance(4, pkg)
	}

	installProgress.Stop("Installation successful!", 0)

	// Example 4: Task with message updates (no progress advancement)
	tap.Message("Example 4: Task with Message Updates")

	statusProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "heavy",
		Max:   10,
		Size:  25,
	})

	statusProgress.Start("Initializing...")
	time.Sleep(500 * time.Millisecond)

	// Update messages without advancing progress
	statusProgress.Message("Connecting to server...")
	time.Sleep(500 * time.Millisecond)

	statusProgress.Message("Authenticating...")
	time.Sleep(500 * time.Millisecond)

	statusProgress.Message("Loading configuration...")
	time.Sleep(500 * time.Millisecond)

	// Now start making actual progress
	statusProgress.Advance(3, "Syncing data...")
	time.Sleep(400 * time.Millisecond)

	statusProgress.Advance(4, "Processing updates...")
	time.Sleep(400 * time.Millisecond)

	statusProgress.Advance(3, "Cleaning up...")
	time.Sleep(400 * time.Millisecond)

	statusProgress.Stop("Task completed successfully!", 0)

	// Example 5: Demonstrate cancellation
	tap.Message("Example 5: Cancelled Task")

	cancelProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "heavy",
		Max:   100,
		Size:  35,
	})

	cancelProgress.Start("Running long task...")

	// Simulate some progress then cancel
	for i := 0; i < 30; i += 10 {
		time.Sleep(200 * time.Millisecond)
		cancelProgress.Advance(10, fmt.Sprintf("Processing step %d...", i/10+1))
	}

	// Simulate cancellation
	cancelProgress.Stop("Task was cancelled by user", 1)

	// Example 6: Error scenario
	tap.Message("Example 6: Task with Error")

	errorProgress := tap.NewProgress(tap.ProgressOptions{
		Style: "block",
		Max:   10,
		Size:  20,
	})

	errorProgress.Start("Attempting risky operation...")
	time.Sleep(300 * time.Millisecond)

	errorProgress.Advance(3, "Step 1 completed...")
	time.Sleep(300 * time.Millisecond)

	errorProgress.Advance(2, "Step 2 completed...")
	time.Sleep(300 * time.Millisecond)

	// Simulate error
	errorProgress.Stop("Operation failed with error", 2)

	tap.Outro("All progress bar examples completed successfully! 🎉")
}
