# Go-TTY Terminal Module Implementation Plan

> **For Claude:** Use `${SUPERPOWERS_SKILLS_ROOT}/skills/collaboration/executing-plans/SKILL.md` to implement this plan task-by-task.

**Goal:** Replace the current `internal/terminal` module (using `github.com/noojuno/keyboard`) with a simpler channel-based implementation using `github.com/mattn/go-tty`.

**Architecture:** Create a minimal terminal module that wraps go-tty and provides a channel-based key stream. The module handles raw TTY I/O, escape sequence parsing, and signal cleanup. The existing prompt.go will be refactored to consume keys from a channel instead of an event-based callback system.

**Tech Stack:**
- `github.com/mattn/go-tty` for raw terminal I/O
- Channel-based concurrency for key streaming
- ANSI escape sequences for cursor control

---

## Task 1: Create New Terminal Module Structure

**Files:**
- Create: `internal/terminal/terminal.go`
- Create: `internal/terminal/key.go`
- Create: `internal/terminal/ansi.go`

**Step 1: Write key.go with Key type definition**

Create `internal/terminal/key.go`:

```go
package terminal

// Key represents a parsed keyboard input event
type Key struct {
	Name string // "up", "down", "left", "right", "return", "escape", "backspace", "delete", "space", "tab", or lowercase letter
	Rune rune   // The actual character (0 for special keys)
	Ctrl bool   // True if Ctrl modifier was pressed
}
```

**Step 2: Write ansi.go with ANSI escape sequences**

Create `internal/terminal/ansi.go`:

```go
package terminal

// ANSI escape sequences for terminal control
const (
	CursorHide = "\x1b[?25l"
	CursorShow = "\x1b[?25h"
	ClearLine  = "\r\x1b[K"
	CursorUp   = "\x1b[A"
	EraseDown  = "\x1b[J"
	SaveCursor = "\x1b[s"
	RestCursor = "\x1b[u"
)

// MoveUp returns ANSI sequence to move cursor up n lines
func MoveUp(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += CursorUp
	}
	return result
}
```

**Step 3: Write terminal.go with basic structure**

Create `internal/terminal/terminal.go`:

```go
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
```

**Step 4: Commit**

```bash
git add internal/terminal/
git commit -m "feat: create new terminal module with go-tty and channel-based key input"
```

---

## Task 2: Update Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Remove old keyboard dependency and add go-tty**

```bash
go get github.com/mattn/go-tty@latest
go mod edit -droprequire=github.com/noojuno/keyboard
go mod tidy
```

Expected: go.mod now has `github.com/mattn/go-tty` and no longer has `github.com/noojuno/keyboard`

**Step 2: Verify go.mod changes**

```bash
grep -E "(go-tty|keyboard)" go.mod
```

Expected output should show `github.com/mattn/go-tty` and NOT show `github.com/noojuno/keyboard`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: replace keyboard dependency with go-tty"
```

---

## Task 3: Create Adapter Interfaces for Backward Compatibility

**Files:**
- Modify: `internal/terminal/terminal.go`

**Step 1: Add event-based adapter for Reader**

Add to `internal/terminal/terminal.go` after the existing Reader struct:

```go
// On registers a callback for key events (compatibility adapter)
func (r *Reader) On(event string, handler func(string, Key)) {
	if event != "keypress" {
		return
	}

	// Spawn goroutine to convert channel reads to callbacks
	go func() {
		for key := range r.keys {
			char := ""
			if key.Rune != 0 {
				char = string(key.Rune)
			}
			handler(char, key)
		}
	}()
}
```

**Step 2: Add event emitter for Writer**

Add to `internal/terminal/terminal.go` after Writer struct:

```go
type resizeHandler struct {
	handlers []func()
	mu       sync.Mutex
}

var globalResizeHandler = &resizeHandler{}

// On registers a callback for terminal events
func (w *Writer) On(event string, handler func()) {
	if event != "resize" {
		return
	}

	globalResizeHandler.mu.Lock()
	defer globalResizeHandler.mu.Unlock()

	if len(globalResizeHandler.handlers) == 0 {
		// First handler - set up signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGWINCH)
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

// Emit triggers an event (no-op for compatibility)
func (w *Writer) Emit(event string) {}
```

**Step 3: Update Terminal to return Reader and Writer with adapters**

Modify the `Terminal` struct in `internal/terminal/terminal.go`:

```go
// Terminal manages terminal I/O operations with channel-based key input
type Terminal struct {
	tty       *tty.TTY
	keys      chan Key
	done      chan struct{}
	closeOnce sync.Once
	Reader    *Reader
	Writer    *Writer
}
```

Then update `New()` function to initialize Reader and Writer:

```go
func New() (*Terminal, error) {
	t, err := tty.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open tty: %w", err)
	}

	keysChan := make(chan Key, 10)

	term := &Terminal{
		tty:    t,
		keys:   keysChan,
		done:   make(chan struct{}),
		Reader: &Reader{keys: keysChan},
		Writer: &Writer{},
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
```

**Step 4: Commit**

```bash
git add internal/terminal/terminal.go
git commit -m "feat: add event-based adapter for backward compatibility"
```

---

## Task 4: Update Types and Remove Old Terminal Files

**Files:**
- Modify: `types.go:92`
- Delete: `internal/terminal/terminal_unix.go`
- Delete: `internal/terminal/terminal_windows.go`

**Step 1: Update Key type alias in types.go**

In `types.go` at line 92, the line currently reads:
```go
type Key = terminal.Key
```

This should remain the same since our new terminal module also exports a `Key` type with compatible fields.

Verify the Key type matches by checking if `terminal.Key` has `Name`, `Ctrl` fields (Rune is new but shouldn't break existing code).

**Step 2: Remove old platform-specific terminal files**

```bash
rm /Users/yaroslavk/home/tap/internal/terminal/terminal_unix.go
rm /Users/yaroslavk/home/tap/internal/terminal/terminal_windows.go
```

Expected: Files removed successfully

**Step 3: Verify no broken imports**

```bash
go build ./...
```

Expected: Build should succeed with no import errors

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor: remove old platform-specific terminal files"
```

---

## Task 5: Update ANSI Constants in types.go

**Files:**
- Modify: `types.go:117-123`

**Step 1: Update ANSI constants to use terminal package**

Replace lines 117-123 in `types.go`:

Old:
```go
const (
	CursorHide = "\x1b[?25l"
	CursorShow = "\x1b[?25h"
	EraseLine  = "\x1b[K"
	CursorUp   = "\x1b[A"
	EraseDown  = "\x1b[J"
)
```

New:
```go
const (
	CursorHide = terminal.CursorHide
	CursorShow = terminal.CursorShow
	EraseLine  = terminal.ClearLine
	CursorUp   = terminal.CursorUp
	EraseDown  = terminal.EraseDown
)
```

**Step 2: Verify build**

```bash
go build ./...
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add types.go
git commit -m "refactor: use ANSI constants from terminal package"
```

---

## Task 6: Test the Implementation

**Files:**
- Create: `internal/terminal/terminal_test.go`

**Step 1: Write test for terminal creation and cleanup**

Create `internal/terminal/terminal_test.go`:

```go
package terminal

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer term.Close()

	if term.tty == nil {
		t.Error("tty should not be nil")
	}

	if term.Reader == nil {
		t.Error("Reader should not be nil")
	}

	if term.Writer == nil {
		t.Error("Writer should not be nil")
	}
}

func TestKeys(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer term.Close()

	keys := term.Keys()
	if keys == nil {
		t.Error("Keys() should return non-nil channel")
	}
}

func TestClose(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	term.Close()

	// Verify keys channel is closed
	select {
	case _, ok := <-term.Keys():
		if ok {
			t.Error("Keys channel should be closed after Close()")
		}
	case <-time.After(100 * time.Millisecond):
		// Channel might not be closed immediately, give it time
	}

	// Calling Close() again should not panic
	term.Close()
}

func TestParseKey(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer term.Close()

	tests := []struct {
		name     string
		input    rune
		expected Key
	}{
		{"enter", 13, Key{Name: "return", Rune: 0}},
		{"backspace", 127, Key{Name: "backspace", Rune: 0}},
		{"tab", 9, Key{Name: "tab", Rune: 0}},
		{"space", 32, Key{Name: "space", Rune: ' '}},
		{"ctrl-c", 3, Key{Name: "c", Rune: 'c', Ctrl: true}},
		{"letter a", 'a', Key{Name: "a", Rune: 'a'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := term.parseKey(tt.input)
			if result.Name != tt.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tt.expected.Name)
			}
			if result.Rune != tt.expected.Rune {
				t.Errorf("Rune: got %q, want %q", result.Rune, tt.expected.Rune)
			}
			if result.Ctrl != tt.expected.Ctrl {
				t.Errorf("Ctrl: got %v, want %v", result.Ctrl, tt.expected.Ctrl)
			}
		})
	}
}

func TestMoveUp(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, ""},
		{1, "\x1b[A"},
		{3, "\x1b[A\x1b[A\x1b[A"},
	}

	for _, tt := range tests {
		result := MoveUp(tt.n)
		if result != tt.expected {
			t.Errorf("MoveUp(%d): got %q, want %q", tt.n, result, tt.expected)
		}
	}
}
```

**Step 2: Run tests**

```bash
go test ./internal/terminal -v
```

Expected: All tests pass. Note: Some tests may be skipped in CI if no TTY available.

**Step 3: Commit**

```bash
git add internal/terminal/terminal_test.go
git commit -m "test: add tests for terminal module"
```

---

## Task 7: Manual Integration Test

**Files:**
- Create: `cmd/terminal-test/main.go` (temporary test binary)

**Step 1: Write manual test program**

Create `cmd/terminal-test/main.go`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/yarlson/tap/internal/terminal"
)

func main() {
	term, err := terminal.New()
	if err != nil {
		log.Fatal(err)
	}
	defer term.Close()

	fmt.Print(terminal.CursorHide)
	defer fmt.Print(terminal.CursorShow)

	fmt.Println("Terminal test - press keys (Ctrl+C or ESC to exit):")

	for key := range term.Keys() {
		if key.Name == "escape" || (key.Ctrl && key.Name == "c") {
			break
		}

		fmt.Printf("%sKey: name=%q rune=%q ctrl=%v\n",
			terminal.ClearLine,
			key.Name,
			key.Rune,
			key.Ctrl,
		)
	}

	fmt.Println("\nTest complete!")
}
```

**Step 2: Run manual test**

```bash
go run cmd/terminal-test/main.go
```

Expected:
- Program starts and hides cursor
- User can type keys and see them printed
- Arrow keys show as "up", "down", "left", "right"
- ESC or Ctrl+C exits cleanly and shows cursor

**Step 3: Remove test program**

```bash
rm -rf cmd/terminal-test
```

**Step 4: Commit**

```bash
git add -A
git commit -m "test: verify terminal module works interactively"
```

---

## Task 8: Build and Verify Full Integration

**Files:**
- None (verification step)

**Step 1: Build the entire project**

```bash
go build ./...
```

Expected: No compilation errors

**Step 2: Run all tests**

```bash
go test ./...
```

Expected: All tests pass (some may skip if no TTY)

**Step 3: Check for any remaining references to old keyboard package**

```bash
grep -r "noojuno/keyboard" . --exclude-dir=.git --exclude-dir=vendor
```

Expected: No matches found

**Step 4: Verify go.mod is clean**

```bash
cat go.mod
```

Expected: Should see `github.com/mattn/go-tty` and NOT see `github.com/noojuno/keyboard`

---

## Completion

All tasks complete! The terminal module has been successfully replaced with a cleaner go-tty-based implementation using channel-based key streaming.

**Summary of changes:**
1. ✅ Created new `internal/terminal` module with go-tty
2. ✅ Updated dependencies (removed keyboard, added go-tty)
3. ✅ Maintained backward compatibility with event-based adapter
4. ✅ Removed old platform-specific terminal files
5. ✅ Updated ANSI constants
6. ✅ Added comprehensive tests
7. ✅ Verified integration

**Note:** The existing `prompt.go` continues to work because we provided event-based adapters (`Reader.On()` and `Writer.On()`). In the future, prompt.go can be refactored to consume from `term.Keys()` channel directly for better performance and simpler code.
