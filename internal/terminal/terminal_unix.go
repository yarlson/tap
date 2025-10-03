//go:build !windows

package terminal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// setupSignalReraising sets up signal handling that cleans up terminal state
// and then re-raises the signal to allow the hosting application to handle it.
func setupSignalReraising(sigChan chan os.Signal, cleanupOnce *sync.Once, doCleanup func()) {
	go func() {
		sig := <-sigChan

		cleanupOnce.Do(doCleanup)
		// stop notifications and restore default behavior for this signal
		signal.Stop(sigChan)
		signal.Reset(sig)
		// best-effort: re-send the signal to this process to allow default handling
		if s, ok := sig.(syscall.Signal); ok {
			_ = syscall.Kill(os.Getpid(), s)
		}
	}()
}

// setupResizeHandler sets up terminal resize notifications.
func setupResizeHandler(writer *Writer) {
	resizeChan := make(chan os.Signal, 1)
	signal.Notify(resizeChan, syscall.SIGWINCH)

	go func() {
		for range resizeChan {
			writer.Emit("resize")
		}
	}()
}
