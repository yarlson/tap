//go:build windows

package terminal

import (
	"os"
	"os/signal"
	"sync"
)

// setupSignalReraising sets up signal handling that cleans up terminal state.
// On Windows, we don't re-raise the signal as syscall.Kill is not available.
func setupSignalReraising(sigChan chan os.Signal, cleanupOnce *sync.Once, doCleanup func()) {
	go func() {
		sig := <-sigChan

		cleanupOnce.Do(doCleanup)
		// stop notifications and restore default behavior for this signal
		signal.Stop(sigChan)
		signal.Reset(sig)
		// On Windows, we cannot re-raise the signal using syscall.Kill,
		// so we just clean up and let the signal handler exit normally.
	}()
}

// setupResizeHandler sets up terminal resize notifications.
// On Windows, SIGWINCH is not available, so this is a no-op.
func setupResizeHandler(writer *Writer) {
	// Windows does not support SIGWINCH for terminal resize notifications.
	// Terminal resize handling would need to be implemented differently on Windows,
	// potentially using Windows Console API if needed in the future.
}
