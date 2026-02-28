package terminal

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/mattn/go-tty"
)

// ANSI escape sequences for terminal control.
const (
	CursorHide = "\x1b[?25l"
	CursorShow = "\x1b[?25h"
	ClearLine  = "\r\x1b[K"
	CursorUp   = "\x1b[A"
	EraseDown  = "\x1b[J"
	SaveCursor = "\x1b[s"
	RestCursor = "\x1b[u"
)

// MoveUp returns ANSI sequence to move cursor up n lines.
func MoveUp(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += CursorUp
	}

	return result
}

// Key represents a parsed keyboard input event.
type Key struct {
	Name  string // "up", "down", "left", "right", "return", "escape", "backspace", "delete", "space", "tab", or lowercase letter
	Rune  rune   // The actual character (0 for special keys)
	Ctrl  bool   // True if Ctrl modifier was pressed
	Shift bool   // True if Shift modifier was pressed
}

// Terminal manages terminal I/O operations with channel-based key input.
type Terminal struct {
	tty       *tty.TTY
	keys      chan Key
	done      chan struct{}
	closeOnce sync.Once
	Reader    *Reader
	Writer    *Writer
}

// Reader provides read-only access to the key channel.
type Reader struct {
	keys   <-chan Key
	cancel chan struct{} // Cancel channel to stop current consumer
	mu     sync.Mutex    // Protects cancel channel
}

// Writer wraps stdout.
type Writer struct{}

// Singleton terminal management to prevent multiple terminals competing for input.
// When multiple prompts run sequentially, they should share a single TTY reader
// to avoid the race condition where old readKeys goroutines steal keypresses.
var (
	globalTerminal *Terminal
	terminalMu     sync.Mutex
)

// New creates a new terminal instance and starts key reading.
// Uses a singleton pattern: the first call creates the TTY and readKeys goroutine,
// subsequent calls reuse the same terminal but create new Reader/Writer wrappers.
func New() (*Terminal, error) {
	terminalMu.Lock()
	defer terminalMu.Unlock()

	// If we have an existing terminal, return a new wrapper that shares the same
	// keys channel. The single readKeys goroutine continues running.
	if globalTerminal != nil && globalTerminal.tty != nil {
		// Return a wrapper with the same keys channel - no new goroutine!
		term := &Terminal{
			tty:       globalTerminal.tty,
			keys:      globalTerminal.keys,
			done:      globalTerminal.done,
			Reader:    globalTerminal.Reader,
			Writer:    globalTerminal.Writer,
			closeOnce: sync.Once{}, // Fresh once for this wrapper
		}

		return term, nil
	}

	// First terminal - create new TTY
	t, err := tty.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open tty: %w", err)
	}

	keysChan := make(chan Key, 10)
	doneChan := make(chan struct{})

	term := &Terminal{
		tty:    t,
		keys:   keysChan,
		done:   doneChan,
		Reader: &Reader{keys: keysChan},
		Writer: &Writer{},
	}

	globalTerminal = term

	// Enable kitty keyboard protocol to receive modifier information for keys like Shift+Enter.
	// This requests the terminal to send distinct escape sequences for modified keys.
	// Kitty: CSI > 4 m
	// xterm: CSI > 4 m (newer versions)
	fmt.Print("\x1b[>4m") // Request kitty/xterm extended keyboard mode

	// Set up signal handling for clean shutdown
	sigChan := setupTermSignal()

	go func() {
		<-sigChan
		fmt.Print(CursorShow, "\n")
		os.Exit(1)
	}()

	// Start key reading goroutine
	go term.readKeys()

	// Give the readKeys goroutine a chance to start
	runtime.Gosched()

	return term, nil
}

// readKeys continuously reads from TTY and sends parsed keys to channel.
func (t *Terminal) readKeys() {
	defer func() {
		close(t.keys)
	}()

	for {
		select {
		case <-t.done:
			return
		default:
		}

		r, err := t.tty.ReadRune()
		if err != nil {
			continue
		}

		key := t.parseKey(r)
		select {
		case t.keys <- key:
		case <-t.done:
			return
		}
	}
}

// parseKey converts a rune to a Key struct, handling escape sequences.
func (t *Terminal) parseKey(r rune) Key {
	switch r {
	case 27: // ESC
		// Try to read the next runes for escape sequence
		n1, err := t.tty.ReadRune()
		if err != nil {
			return Key{Name: "escape"}
		}

		if n1 == '[' {
			return t.parseCSI()
		}

		return Key{Name: "escape"}
	case 13: // Enter
		return Key{Name: "return"}
	case 127, 8: // Backspace
		return Key{Name: "backspace"}
	case 9: // Tab
		return Key{Name: "tab"}
	case 32: // Space
		return Key{Name: "space", Rune: ' '}
	case 3: // Ctrl+C
		return Key{Name: "c", Rune: 'c', Ctrl: true}
	default:
		if r >= 32 && r <= 126 {
			return Key{Name: string(r), Rune: r}
		}

		return Key{Name: "", Rune: r}
	}
}

// parseCSI parses a CSI (Control Sequence Introducer) sequence after ESC[.
// Handles: arrow keys, delete, kitty keyboard protocol, xterm modifyOtherKeys.
// Supports both semicolon (;) and colon (:) as parameter separators for compatibility.
func (t *Terminal) parseCSI() Key {
	// Collect numeric parameters and terminator
	var params []int
	current := 0
	hasDigit := false

	for {
		ch, err := t.tty.ReadRune()
		if err != nil {
			return Key{Name: "escape"}
		}

		switch {
		case ch >= '0' && ch <= '9':
			current = current*10 + int(ch-'0')
			hasDigit = true

		case ch == ';' || ch == ':':
			// Both ; and : are valid parameter separators
			params = append(params, current)
			current = 0
			hasDigit = false

		default:
			// Terminator character reached
			if hasDigit {
				params = append(params, current)
			}

			return t.resolveCSI(params, ch)
		}
	}
}

// resolveCSI maps collected CSI parameters and terminator to a Key.
func (t *Terminal) resolveCSI(params []int, terminator rune) Key {
	switch terminator {
	case 'A':
		return Key{Name: "up"}
	case 'B':
		return Key{Name: "down"}
	case 'C':
		return Key{Name: "right"}
	case 'D':
		return Key{Name: "left"}

	case '~':
		if len(params) == 0 {
			return Key{Name: "escape"}
		}

		// ESC[3~ → Delete
		if params[0] == 3 {
			return Key{Name: "delete"}
		}

		// xterm modifyOtherKeys: ESC[27;modifier;keycode~
		if params[0] == 27 && len(params) == 3 {
			return t.resolveModifiedKey(params[2], params[1])
		}

		return Key{Name: "escape"}

	case 'u':
		// Kitty keyboard protocol: ESC[keycode;modifiersu (or ESC[keycodeu for unmodified)
		if len(params) == 1 {
			// Single parameter: keycode without explicit modifier
			// Only handle Enter (13); others fall through
			if params[0] == 13 {
				return Key{Name: "return"}
			}
		} else if len(params) >= 2 {
			return t.resolveModifiedKey(params[0], params[1])
		}

		return Key{Name: "escape"}
	}

	return Key{Name: "escape"}
}

// resolveModifiedKey maps a keycode + modifier bitmask to a Key.
func (t *Terminal) resolveModifiedKey(keycode, modifier int) Key {
	// CSI modifier encoding: modifier value = 1 + bitmask
	// modifier=1 means no modifiers (bitmask=0)
	// modifier=2 means shift only (bitmask=1, bit 0 set)
	// modifier=3 means shift+other (bitmask=2, etc.)
	shift := false
	if modifier >= 2 {
		shift = ((modifier - 1) & 0x01) != 0
	}

	if keycode == 13 {
		return Key{Name: "return", Shift: shift}
	}

	return Key{Name: "escape"}
}

// Keys returns the read-only key channel.
func (t *Terminal) Keys() <-chan Key {
	return t.keys
}

// Write implements io.Writer.
func (t *Terminal) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

// On registers a callback for key events (compatibility adapter)
// If a previous handler was registered, its consumer goroutine is stopped first.
func (r *Reader) On(event string, handler func(string, Key)) {
	if event != "keypress" {
		return
	}

	r.mu.Lock()
	// Cancel any existing consumer goroutine
	if r.cancel != nil {
		close(r.cancel)
	}
	// Create new cancel channel for this consumer
	cancel := make(chan struct{})
	r.cancel = cancel
	r.mu.Unlock()

	// Spawn goroutine to convert channel reads to callbacks
	go func() {
		for {
			select {
			case <-cancel:
				return
			case key, ok := <-r.keys:
				if !ok {
					return
				}

				// Check if we're still the active consumer before calling handler
				select {
				case <-cancel:
					return
				default:
				}

				char := ""
				if key.Rune != 0 {
					char = string(key.Rune)
				}

				handler(char, key)
			}
		}
	}()
}

// Reader methods.
func (r *Reader) Read(_ []byte) (int, error) {
	return 0, nil
}

type resizeHandler struct {
	handlers []func()
	mu       sync.Mutex
}

var globalResizeHandler = &resizeHandler{}

// On registers a callback for terminal events.
func (w *Writer) On(event string, handler func()) {
	if event != "resize" {
		return
	}

	globalResizeHandler.mu.Lock()
	defer globalResizeHandler.mu.Unlock()

	if len(globalResizeHandler.handlers) == 0 {
		// First handler - set up signal
		sigChan := setupResizeSignal()

		go func() {
			for range sigChan {
				globalResizeHandler.mu.Lock()
				handlers := append([]func(){}, globalResizeHandler.handlers...)
				globalResizeHandler.mu.Unlock()

				for _, h := range handlers {
					h()
				}
			}
		}()
	}

	globalResizeHandler.handlers = append(globalResizeHandler.handlers, handler)
}

// Emit triggers an event (no-op for compatibility).
func (w *Writer) Emit(_ string) {}

// Writer methods.
func (w *Writer) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}
