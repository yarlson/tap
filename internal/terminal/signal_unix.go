//go:build !windows

package terminal

import (
	"os"
	"os/signal"
	"syscall"
)

// setupResizeSignal sets up SIGWINCH handling for terminal resize on Unix systems.
func setupResizeSignal() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)

	return sigChan
}

// setupTermSignal sets up SIGTERM handling for clean shutdown on Unix systems.
func setupTermSignal() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	return sigChan
}
