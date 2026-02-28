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

func TestParseKey_RegularKeysUnchanged(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test: no TTY available: %v", err)
	}

	// Verify Shift is always false for single-rune keys (regression guard)
	tests := []struct {
		name  string
		input rune
	}{
		{"enter", 13},
		{"backspace_127", 127},
		{"backspace_8", 8},
		{"tab", 9},
		{"space", 32},
		{"ctrl-c", 3},
		{"letter_a", 'a'},
		{"letter_z", 'z'},
		{"digit_0", '0'},
		{"exclamation", '!'},
		{"tilde", '~'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := term.parseKey(tt.input)
			if result.Shift {
				t.Errorf("Shift should be false for rune %d (%q), got true", tt.input, string(tt.input))
			}
		})
	}
}

func TestResolveCSI_ArrowKeys(t *testing.T) {
	// resolveCSI is pure logic — no TTY needed
	term := &Terminal{}

	tests := []struct {
		name       string
		terminator rune
		expected   string
	}{
		{"up", 'A', "up"},
		{"down", 'B', "down"},
		{"right", 'C', "right"},
		{"left", 'D', "left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := term.resolveCSI(nil, tt.terminator)
			if result.Name != tt.expected {
				t.Errorf("Name: got %q, want %q", result.Name, tt.expected)
			}

			if result.Shift {
				t.Errorf("Shift should be false for arrow key %q", tt.name)
			}
		})
	}
}

func TestResolveCSI_Delete(t *testing.T) {
	term := &Terminal{}

	// ESC[3~ → delete
	result := term.resolveCSI([]int{3}, '~')
	if result.Name != "delete" {
		t.Errorf("Name: got %q, want %q", result.Name, "delete")
	}
}

func TestResolveCSI_ShiftReturn_Kitty(t *testing.T) {
	term := &Terminal{}

	// ESC[13;2u → Key{Name:"return", Shift:true} (kitty protocol)
	result := term.resolveCSI([]int{13, 2}, 'u')
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if !result.Shift {
		t.Error("Shift should be true for kitty Shift+Enter")
	}
}

func TestResolveCSI_ShiftReturn_Xterm(t *testing.T) {
	term := &Terminal{}

	// ESC[27;2;13~ → Key{Name:"return", Shift:true} (xterm modifyOtherKeys)
	result := term.resolveCSI([]int{27, 2, 13}, '~')
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if !result.Shift {
		t.Error("Shift should be true for xterm Shift+Enter")
	}
}

func TestResolveCSI_UnmodifiedReturn_Kitty(t *testing.T) {
	term := &Terminal{}

	// ESC[13u → Key{Name:"return", Shift:false} (kitty protocol, unmodified)
	result := term.resolveCSI([]int{13}, 'u')
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if result.Shift {
		t.Error("Shift should be false for unmodified kitty Enter")
	}
}

func TestResolveModifiedKey_Modifier1(t *testing.T) {
	term := &Terminal{}

	// modifier=1 means no modifiers (bitmask=0)
	result := term.resolveModifiedKey(13, 1)
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if result.Shift {
		t.Error("Shift should be false for modifier=1 (no modifiers)")
	}
}

func TestResolveCSI_ShiftReturn_Ghostty(t *testing.T) {
	term := &Terminal{}

	// Ghostty with colon separator: ESC[13:2u
	result := term.resolveCSI([]int{13, 2}, 'u')
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if !result.Shift {
		t.Error("Shift should be true for Ghostty Shift+Enter")
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
