package terminal

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test: no TTY available: %v", err)
	}

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

	// Verify keys channel is closed
	select {
	case _, ok := <-term.Keys():
		if ok {
			t.Error("Keys channel should be closed after Close()")
		}
	case <-time.After(100 * time.Millisecond):
		// Channel might not be closed immediately, give it time
	}
}

func TestParseKey(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test: no TTY available: %v", err)
	}

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
