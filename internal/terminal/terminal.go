package terminal

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/yarlson/tap/internal/prompts"

	"github.com/eiannone/keyboard"
	"golang.org/x/term"
)

// Reader provides terminal input functionality.
type Reader struct {
	mu        sync.Mutex
	listeners map[string][]func(string, prompts.Key)
}

// Writer provides terminal output functionality.
type Writer struct {
	mu        sync.Mutex
	listeners map[string][]func()
}

// Terminal manages terminal I/O operations.
type Terminal struct {
	Reader        *Reader
	Writer        *Writer
	cleanup       func()
	originalFd    int
	originalState *term.State
}

// New creates a new terminal instance with keyboard input and output handling.
func New() (*Terminal, error) {
	// Save original terminal state for restoration
	fd := int(os.Stdin.Fd())
	originalState, err := term.GetState(fd)
	if err != nil {
		// If we can't get terminal state, continue anyway - might not be a TTY
		originalState = nil
	}

	if err := keyboard.Open(); err != nil {
		return nil, err
	}

	reader := &Reader{listeners: make(map[string][]func(string, prompts.Key))}
	writer := &Writer{listeners: make(map[string][]func())}

	stop := make(chan struct{})

	// Set up signal handling to ensure terminal is always restored
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var cleanupOnce sync.Once
	doCleanup := func() {
		close(stop)
		_ = keyboard.Close()

		// Restore original terminal state if we saved it
		if originalState != nil {
			_ = term.Restore(fd, originalState)
		}

		signal.Stop(sigChan)
	}

	// Keyboard input handling goroutine
	go func() {
		start := time.Now()
		var escPending bool
		var escStarted time.Time
		var escPrefix rune // 0, '[' or 'O'
		var escBuf []rune
		// Window to assemble ESC-based sequences (keep small to reduce latency)
		const escWindow = 10 * time.Millisecond
		var escTimer *time.Timer
		stopEscTimer := func() {
			if escTimer != nil {
				escTimer.Stop()
				escTimer = nil
			}
		}
		for {
			select {
			case <-stop:
				return
			default:
			}

			r, key, err := keyboard.GetKey()
			if err != nil {
				continue
			}

			char := string(r)
			name := ""
			ctrl := false

			// If we are assembling an escape sequence from a previous ESC
			if escPending {
				// If the library already decoded an arrow, use it immediately
				if key == keyboard.KeyArrowUp || key == keyboard.KeyArrowDown || key == keyboard.KeyArrowLeft || key == keyboard.KeyArrowRight {
					switch key {
					case keyboard.KeyArrowUp:
						name = "up"
					case keyboard.KeyArrowDown:
						name = "down"
					case keyboard.KeyArrowLeft:
						name = "left"
					case keyboard.KeyArrowRight:
						name = "right"
					}
					escPending = false
					escPrefix = 0
					escBuf = nil
					k := prompts.Key{Name: name, Ctrl: false}
					reader.emit("", k)
					continue
				}
				// First follow-up may be '[' or 'O'
				if escPrefix == 0 && (r == '[' || r == 'O') {
					escPrefix = r
					escBuf = append(escBuf, r)
					continue
				}
				// If we have a prefix, map final byte
				if escPrefix != 0 && (r == 'A' || r == 'B' || r == 'C' || r == 'D') {
					switch r {
					case 'A':
						name = "up"
					case 'B':
						name = "down"
					case 'C':
						name = "right"
					case 'D':
						name = "left"
					}
					// Clear pending and emit arrow
					escPending = false
					escPrefix = 0
					escBuf = nil
					char = ""
					stopEscTimer()
					k := prompts.Key{Name: name, Ctrl: false}
					reader.emit(char, k)
					continue
				}
				// Timeout or unrelated key: if we saw a prefix, swallow; if not, emit escape
				if time.Since(escStarted) >= escWindow || (escPrefix == 0 && r != '[' && r != 'O') {
					if escPrefix == 0 && len(escBuf) == 0 {
						// Plain ESC
						escPending = false
						kEsc := prompts.Key{Name: "escape", Ctrl: false}
						stopEscTimer()
						reader.emit("", kEsc)
						// Fall through to process current event below
					} else {
						// Incomplete CSI/SS3 sequence: treat as a horizontal move to avoid cancel
						escPending = false
						escDir := "right"
						kMv := prompts.Key{Name: escDir, Ctrl: false}
						stopEscTimer()
						reader.emit("", kMv)
					}
					escPrefix = 0
					escBuf = nil
				} else {
					// Continue waiting for completion
					if r != 0 {
						escBuf = append(escBuf, r)
					}
					continue
				}
			}

			switch key {
			case keyboard.KeyArrowUp:
				name = "up"
			case keyboard.KeyArrowDown:
				name = "down"
			case keyboard.KeyArrowLeft:
				name = "left"
			case keyboard.KeyArrowRight:
				name = "right"
			case keyboard.KeyEnter:
				name = "return"
			case keyboard.KeyBackspace, keyboard.KeyBackspace2:
				name = "backspace"
			case keyboard.KeyEsc:
				// Some terminals emit a stray ESC on startup; ignore it within a short window
				if time.Since(start) < 100*time.Millisecond {
					continue
				}
				// Begin ESC sequence collection; do not emit yet
				escPending = true
				escStarted = time.Now()
				escPrefix = 0
				escBuf = nil
				// In some terminals, the ESC event carries '[' already
				if r == '[' || r == 'O' {
					escPrefix = r
					escBuf = append(escBuf, r)
				}
				// Arm timer to emit fallback without needing another key event
				stopEscTimer()
				escTimer = time.AfterFunc(escWindow, func() {
					// Timer callback runs concurrently; emit based on current pending state
					if !escPending {
						return
					}
					if escPrefix == 0 && len(escBuf) == 0 {
						// Plain Escape
						kEsc := prompts.Key{Name: "escape", Ctrl: false}
						reader.emit("", kEsc)
					} else {
						// Incomplete sequence -> treat as right
						kMv := prompts.Key{Name: "right", Ctrl: false}
						reader.emit("", kMv)
					}
					escPending = false
					escPrefix = 0
					escBuf = nil
					stopEscTimer()
				})
				continue
			case keyboard.KeyCtrlC:
				char = "\x03"
				name = "c"
				ctrl = true
			case keyboard.KeyDelete:
				name = "delete"
			case keyboard.KeySpace:
				char = " "
				name = "space"
			default:
				if r != 0 {
					char = string(r)
					name = strings.ToLower(string(r))
				}
			}

			k := prompts.Key{Name: name, Ctrl: ctrl}
			reader.emit(char, k)
		}
	}()

	// Signal handler for clean shutdown without forcing process exit.
	// We clean up terminal state, then restore default handling and re-raise
	// the signal so the hosting application decides the exit policy.
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

	// Terminal resize notifications
	resizeChan := make(chan os.Signal, 1)
	signal.Notify(resizeChan, syscall.SIGWINCH)
	go func() {
		for range resizeChan {
			writer.Emit("resize")
		}
	}()

	cleanup := func() {
		cleanupOnce.Do(doCleanup)
	}

	return &Terminal{
		Reader:        reader,
		Writer:        writer,
		cleanup:       cleanup,
		originalFd:    fd,
		originalState: originalState,
	}, nil
}

// Close releases terminal resources.
func (t *Terminal) Close() {
	if t.cleanup != nil {
		t.cleanup()
	}
}

func (r *Reader) Read(_ []byte) (int, error) { return 0, nil }

func (r *Reader) On(event string, handler func(string, prompts.Key)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners[event] = append(r.listeners[event], handler)
}

func (r *Reader) emit(char string, key prompts.Key) {
	r.mu.Lock()
	hs := append([]func(string, prompts.Key){}, r.listeners["keypress"]...)
	r.mu.Unlock()
	for _, h := range hs {
		h(char, key)
	}
}

func (w *Writer) Write(b []byte) (int, error) { return os.Stdout.Write(b) }

func (w *Writer) On(event string, handler func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.listeners[event] = append(w.listeners[event], handler)
}

func (w *Writer) Emit(event string) {
	w.mu.Lock()
	hs := append([]func(){}, w.listeners[event]...)
	w.mu.Unlock()
	for _, h := range hs {
		h()
	}
}
