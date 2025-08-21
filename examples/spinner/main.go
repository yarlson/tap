package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yarlson/tap/prompts"
	"github.com/yarlson/tap/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("Failed to initialize terminal: %v\n", err)
		os.Exit(1)
	}
	defer term.Close()

	fmt.Println("Spinner Examples")
	fmt.Println()

	// Example 1: Default dots indicator
	spin := prompts.NewSpinner(prompts.SpinnerOptions{Output: term.Writer})
	spin.Start("Connecting")
	time.Sleep(2 * time.Second)
	spin.Stop("Connected", 0)
	fmt.Println()

	// Example 2: Timer indicator
	timerSpin := prompts.NewSpinner(prompts.SpinnerOptions{Output: term.Writer, Indicator: "timer"})
	timerSpin.Start("Fetching data")
	time.Sleep(1500 * time.Millisecond)
	timerSpin.Stop("Done", 0)
	fmt.Println()

	// Example 3: Custom frames and delay
	custom := prompts.NewSpinner(prompts.SpinnerOptions{Output: term.Writer, Frames: []string{"-", "\\", "|", "/"}, Delay: 100 * time.Millisecond})
	custom.Start("Working")
	time.Sleep(1200 * time.Millisecond)
	custom.Stop("Complete", 0)
}
