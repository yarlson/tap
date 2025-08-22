package main

import (
    "fmt"
    "time"

    "github.com/yarlson/tap/tap"
)

func main() {
	fmt.Println("Spinner Examples")
	fmt.Println()

	// Example 1: Default dots indicator
    spin := tap.NewSpinner(tap.SpinnerOptions{})
	spin.Start("Connecting")
	time.Sleep(2 * time.Second)
	spin.Stop("Connected", 0)
	fmt.Println()

	// Example 2: Timer indicator
    timerSpin := tap.NewSpinner(tap.SpinnerOptions{Indicator: "timer"})
	timerSpin.Start("Fetching data")
	time.Sleep(1500 * time.Millisecond)
	timerSpin.Stop("Done", 0)
	fmt.Println()

	// Example 3: Custom frames and delay
    custom := tap.NewSpinner(tap.SpinnerOptions{Frames: []string{"-", "\\", "|", "/"}, Delay: 100 * time.Millisecond})
	custom.Start("Working")
	time.Sleep(1200 * time.Millisecond)
	custom.Stop("Complete", 0)
}
