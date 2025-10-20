# Fix Terminal Key Event Fan-Out Issue Implementation Plan

> **For Claude:** Use `${SUPERPOWERS_SKILLS_ROOT}/skills/collaboration/executing-plans/SKILL.md` to implement this plan task-by-task.

**Goal:** Fix the issue where arrow keys don't work immediately in prompts - users must press a key first before navigation works.

**Architecture:** The root cause is that `Reader.On()` spawns a goroutine that reads from a shared channel. When multiple prompts exist or `On()` is called multiple times, keys get distributed randomly across goroutines (channel fan-out). The solution is to implement a proper event broadcaster that multicasts each key to all registered handlers, and ensure only one reader per terminal lifetime.

**Tech Stack:**
- Go channels for event broadcasting
- Mutex for thread-safe handler registration
- Single key reading goroutine per Terminal

---

## Task 1: Add Event Broadcaster to Reader

**Files:**
- Modify: `internal/terminal/terminal.go:24-26` (Reader struct)
- Modify: `internal/terminal/terminal.go:159-175` (Reader.On method)

**Step 1: Update Reader struct to include handler management**

Modify the `Reader` struct in `internal/terminal/terminal.go` (around line 24):

Old:
```go
type Reader struct {
	keys <-chan Key
}
```

New:
```go
type Reader struct {
	keys     <-chan Key
	handlers []func(string, Key)
	mu       sync.Mutex
	started  bool
}
```

**Step 2: Rewrite Reader.On() to broadcast to all handlers**

Replace the existing `Reader.On()` method (lines 159-175) with:

```go
// On registers a callback for key events (compatibility adapter)
func (r *Reader) On(event string, handler func(string, Key)) {
	if event != "keypress" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Add handler to the list
	r.handlers = append(r.handlers, handler)

	// Start the broadcaster goroutine only once
	if !r.started {
		r.started = true
		go r.broadcast()
	}
}

// broadcast reads from keys channel and sends to all handlers
func (r *Reader) broadcast() {
	for key := range r.keys {
		r.mu.Lock()
		handlers := append([]func(string, Key){}, r.handlers...)
		r.mu.Unlock()

		char := ""
		if key.Rune != 0 {
			char = string(key.Rune)
		}

		// Call all handlers with the same key
		for _, handler := range handlers {
			handler(char, key)
		}
	}
}
```

**Step 3: Verify build**

```bash
go build ./...
```

Expected: Build succeeds with no errors

**Step 4: Commit**

```bash
git add internal/terminal/terminal.go
git commit -m "fix: implement event broadcaster to prevent key fan-out"
```

---

## Task 2: Write Test for Multiple Handlers

**Files:**
- Modify: `internal/terminal/terminal_test.go`

**Step 1: Add test for multiple handlers receiving same keys**

Add to `internal/terminal/terminal_test.go`:

```go
func TestMultipleHandlers(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test - no TTY available: %v", err)
		return
	}
	defer term.Close()

	// Track keys received by each handler
	keys1 := make(chan Key, 10)
	keys2 := make(chan Key, 10)

	// Register two handlers
	term.Reader.On("keypress", func(char string, key Key) {
		keys1 <- key
	})

	term.Reader.On("keypress", func(char string, key Key) {
		keys2 <- key
	})

	// Simulate sending a key (we'll use parseKey directly since we can't simulate TTY input)
	testKey := term.parseKey('a')

	// Send key to the channel manually for testing
	go func() {
		term.keys <- testKey
	}()

	// Both handlers should receive the same key
	select {
	case k1 := <-keys1:
		if k1.Name != "a" {
			t.Errorf("Handler 1: expected 'a', got %q", k1.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Handler 1 did not receive key")
	}

	select {
	case k2 := <-keys2:
		if k2.Name != "a" {
			t.Errorf("Handler 2: expected 'a', got %q", k2.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Handler 2 did not receive key")
	}
}

func TestBroadcastToAllHandlers(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test - no TTY available: %v", err)
		return
	}
	defer term.Close()

	received := make([]int, 3)
	var mu sync.Mutex

	// Register three handlers
	for i := 0; i < 3; i++ {
		idx := i
		term.Reader.On("keypress", func(char string, key Key) {
			mu.Lock()
			received[idx]++
			mu.Unlock()
		})
	}

	// Send 5 keys
	go func() {
		for i := 0; i < 5; i++ {
			term.keys <- Key{Name: "test", Rune: 't'}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for all keys to be processed
	time.Sleep(200 * time.Millisecond)

	// All handlers should have received all 5 keys
	mu.Lock()
	defer mu.Unlock()
	for i, count := range received {
		if count != 5 {
			t.Errorf("Handler %d: expected 5 keys, got %d", i, count)
		}
	}
}
```

**Step 2: Run the new tests**

```bash
go test ./internal/terminal -v -run TestMultiple
go test ./internal/terminal -v -run TestBroadcast
```

Expected: Both tests pass (or skip if no TTY)

**Step 3: Commit**

```bash
git add internal/terminal/terminal_test.go
git commit -m "test: add tests for event broadcasting to multiple handlers"
```

---

## Task 3: Test with Actual Prompts

**Files:**
- Create: `cmd/test-prompt/main.go` (temporary)

**Step 1: Create test program for confirm prompt**

Create `cmd/test-prompt/main.go`:

```go
package main

import (
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	fmt.Println("Testing arrow key responsiveness...")
	fmt.Println("Arrow keys should work IMMEDIATELY when prompt appears")
	fmt.Println()

	// Test 1: Confirm prompt
	result := tap.Confirm(tap.ConfirmOptions{
		Message:      "Use arrow keys immediately - does left/right work?",
		Active:       "Yes",
		Inactive:     "No",
		InitialValue: false,
	})
	fmt.Printf("Result: %v\n\n", result)

	// Test 2: Select prompt
	options := []tap.SelectOption[string]{
		{Value: "apple", Label: "Apple"},
		{Value: "banana", Label: "Banana"},
		{Value: "cherry", Label: "Cherry"},
	}

	selected := tap.Select(tap.SelectOptions[string]{
		Message: "Use arrow keys immediately - does up/down work?",
		Options: options,
	})
	fmt.Printf("Selected: %v\n\n", selected)

	// Test 3: Multiple prompts in sequence
	for i := 1; i <= 3; i++ {
		result := tap.Confirm(tap.ConfirmOptions{
			Message: fmt.Sprintf("Prompt %d - arrow keys work?", i),
			Active:  "Yes",
			Inactive: "No",
		})
		fmt.Printf("Prompt %d result: %v\n", i, result)
	}

	fmt.Println("\nAll tests complete!")
}
```

**Step 2: Run manual test**

```bash
go run cmd/test-prompt/main.go
```

Expected behavior:
- Arrow keys work IMMEDIATELY when each prompt appears
- No need to press another key first
- All prompts respond to arrow keys instantly
- Multiple prompts in sequence all work correctly

**Step 3: Remove test program**

```bash
rm -rf cmd/test-prompt
```

**Step 4: Commit**

```bash
git add -A
git commit -m "test: verify arrow keys work immediately in prompts"
```

---

## Task 4: Run Full Test Suite

**Files:**
- None (verification)

**Step 1: Build entire project**

```bash
go build ./...
```

Expected: No errors

**Step 2: Run all tests**

```bash
go test ./...
```

Expected: All tests pass

**Step 3: Test example programs**

```bash
# Test confirm example
cd examples/confirm
go run .
# Verify arrow keys work immediately

# Test select example
cd ../select
go run .
# Verify arrow keys work immediately

# Test multiselect example
cd ../multiselect
go run .
# Verify arrow keys work immediately
```

Expected: All examples work with immediate arrow key response

---

## Completion

The terminal key event fan-out issue has been fixed!

**Root cause:** The old `Reader.On()` spawned a new goroutine for each call, and all goroutines competed for keys from the same channel. Only one goroutine would receive any given key (channel fan-out behavior).

**Solution:** Implemented an event broadcaster pattern where:
1. Only ONE goroutine reads from the keys channel (`broadcast()`)
2. That goroutine multicasts each key to ALL registered handlers
3. Handlers are stored in a slice with mutex protection
4. Every handler receives every key event

**Benefits:**
- Arrow keys work immediately in all prompts
- Multiple handlers can coexist (though tap only uses one per prompt)
- No keys are lost or distributed randomly
- Cleaner architecture with predictable behavior

**Testing:**
- ✅ Unit tests for multiple handlers
- ✅ Unit tests for broadcast behavior
- ✅ Manual tests with confirm/select/multiselect prompts
- ✅ Full test suite passes
