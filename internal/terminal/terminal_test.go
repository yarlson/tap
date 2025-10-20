package terminal

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test: no TTY available: %v", err)
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
		t.Skipf("Skipping test: no TTY available: %v", err)
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
		t.Skipf("Skipping test: no TTY available: %v", err)
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
		t.Skipf("Skipping test: no TTY available: %v", err)
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
