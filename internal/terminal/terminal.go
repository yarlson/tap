package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mattn/go-tty"
)

// Terminal manages terminal I/O operations with channel-based key input
type Terminal struct {
	tty       *tty.TTY
	keys      chan Key
	done      chan struct{}
	closeOnce sync.Once
}

// Reader provides read-only access to the key channel
type Reader struct {
	keys <-chan Key
}

// Writer wraps stdout
type Writer struct{}

// New creates a new terminal instance and starts key reading
func New() (*Terminal, error) {
	t, err := tty.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open tty: %w", err)
	}

	term := &Terminal{
		tty:  t,
		keys: make(chan Key, 10), // Buffered to prevent blocking
		done: make(chan struct{}),
	}

	// Set up signal handling for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Print(CursorShow, "\n")
		os.Exit(1)
	}()

	// Start key reading goroutine
	go term.readKeys()

	return term, nil
}

// readKeys continuously reads from TTY and sends parsed keys to channel
func (t *Terminal) readKeys() {
	defer close(t.keys)

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

// parseKey converts a rune to a Key struct, handling escape sequences
func (t *Terminal) parseKey(r rune) Key {
	switch r {
	case 27: // ESC
		// Try to read the next runes for escape sequence
		n1, err := t.tty.ReadRune()
		if err != nil {
			return Key{Name: "escape", Rune: 0}
		}
		if n1 == '[' {
			n2, err := t.tty.ReadRune()
			if err != nil {
				return Key{Name: "escape", Rune: 0}
			}
			switch n2 {
			case 'A':
				return Key{Name: "up", Rune: 0}
			case 'B':
				return Key{Name: "down", Rune: 0}
			case 'C':
				return Key{Name: "right", Rune: 0}
			case 'D':
				return Key{Name: "left", Rune: 0}
			case '3':
				// Delete key is ESC[3~
				t.tty.ReadRune() // consume '~'
				return Key{Name: "delete", Rune: 0}
			}
		}
		return Key{Name: "escape", Rune: 0}
	case 13: // Enter
		return Key{Name: "return", Rune: 0}
	case 127, 8: // Backspace
		return Key{Name: "backspace", Rune: 0}
	case 9: // Tab
		return Key{Name: "tab", Rune: 0}
	case 32: // Space
		return Key{Name: "space", Rune: ' '}
	case 3: // Ctrl+C
		return Key{Name: "c", Rune: 'c', Ctrl: true}
	default:
		if r >= 32 && r <= 126 {
			// Printable ASCII
			return Key{Name: string(r), Rune: r}
		}
		// Unknown control character
		return Key{Name: "", Rune: r}
	}
}

// Keys returns the read-only key channel
func (t *Terminal) Keys() <-chan Key {
	return t.keys
}

// Write implements io.Writer
func (t *Terminal) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

// Close releases terminal resources
func (t *Terminal) Close() {
	t.closeOnce.Do(func() {
		close(t.done)
		if t.tty != nil {
			t.tty.Close()
		}
	})
}

// Reader methods
func (r *Reader) Read(p []byte) (int, error) {
	return 0, nil
}

// Writer methods
func (w *Writer) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}
