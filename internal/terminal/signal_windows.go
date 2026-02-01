//go:build windows

package terminal

import (
	"os"
	"os/signal"
)

// setupResizeSignal returns a channel for resize events on Windows.
// Windows does not have SIGWINCH, so resize detection is handled differently.
// The channel is returned but will not receive resize signals.
// Applications can poll terminal size if resize detection is needed.
func setupResizeSignal() chan os.Signal {
	// Return an unbuffered channel that will never receive events.
	// Windows handles terminal resize differently than Unix systems.
	return make(chan os.Signal, 1)
}

// setupTermSignal sets up interrupt handling for clean shutdown on Windows.
func setupTermSignal() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	return sigChan
}
