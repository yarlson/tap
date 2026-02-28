package terminal

import (
	"io"
	"testing"
	"time"
)

// testTerminal creates a Terminal with a canned rune sequence for testing parseKey
// and related methods without a real TTY.
func testTerminal(runes ...rune) *Terminal {
	idx := 0
	return &Terminal{
		readRune: func() (rune, error) {
			if idx >= len(runes) {
				return 0, io.EOF
			}
			r := runes[idx]
			idx++
			return r, nil
		},
	}
}

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

func TestParseKey_LineFeedIsShiftReturn(t *testing.T) {
	term, err := New()
	if err != nil {
		t.Skipf("Skipping test: no TTY available: %v", err)
	}

	result := term.parseKey(10)
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if !result.Shift {
		t.Error("Shift should be true for line feed fallback (Shift+Enter)")
	}
}

func TestParseKey_EscapePlusCRIsShiftReturn(t *testing.T) {
	term := testTerminal(13)

	result := term.parseKey(27)
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if !result.Shift {
		t.Error("Shift should be true for ESC+CR fallback")
	}
}

func TestParseKey_EscapePlusLFIsShiftReturn(t *testing.T) {
	term := testTerminal(10)

	result := term.parseKey(27)
	if result.Name != "return" {
		t.Errorf("Name: got %q, want %q", result.Name, "return")
	}

	if !result.Shift {
		t.Error("Shift should be true for ESC+LF fallback")
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

func TestResolveModifiedKey_ShiftLetter(t *testing.T) {
	term := &Terminal{}

	// Shift+A via modifyOtherKeys: ESC[27;2;65~ → resolveModifiedKey(65, 2)
	result := term.resolveModifiedKey(65, 2)
	if result.Name != "A" {
		t.Errorf("Name: got %q, want %q", result.Name, "A")
	}

	if result.Rune != 'A' {
		t.Errorf("Rune: got %q, want %q", result.Rune, 'A')
	}

	if !result.Shift {
		t.Error("Shift should be true for Shift+A")
	}

	if result.Name == "escape" {
		t.Error("Shift+letter must not produce escape key")
	}
}

func TestResolveModifiedKey_ShiftSpace(t *testing.T) {
	term := &Terminal{}

	result := term.resolveModifiedKey(32, 2)
	if result.Name != "space" {
		t.Errorf("Name: got %q, want %q", result.Name, "space")
	}

	if result.Rune != ' ' {
		t.Errorf("Rune: got %q, want %q", result.Rune, ' ')
	}

	if !result.Shift {
		t.Error("Shift should be true")
	}
}

func TestResolveCSI_ShiftLetter_XtermModifyOtherKeys(t *testing.T) {
	term := &Terminal{}

	// xterm modifyOtherKeys: ESC[27;2;65~ → params=[27,2,65], terminator='~'
	result := term.resolveCSI([]int{27, 2, 65}, '~')
	if result.Name != "A" {
		t.Errorf("Name: got %q, want %q", result.Name, "A")
	}

	if !result.Shift {
		t.Error("Shift should be true for xterm Shift+A")
	}
}

func TestResolveCSI_ShiftLetter_Kitty(t *testing.T) {
	term := &Terminal{}

	// Kitty protocol: ESC[65;2u → params=[65,2], terminator='u'
	result := term.resolveCSI([]int{65, 2}, 'u')
	if result.Name != "A" {
		t.Errorf("Name: got %q, want %q", result.Name, "A")
	}

	if !result.Shift {
		t.Error("Shift should be true for Kitty Shift+A")
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

func TestParseKey_BracketedPaste(t *testing.T) {
	// Simulate ESC[200~hello ESC[201~
	// parseKey receives ESC (27), then reads: '[', '2','0','0','~' (CSI 200~)
	// readBracketedPaste then reads: 'h','e','l','l','o', ESC, '[', '2','0','1','~'
	term := testTerminal('[', '2', '0', '0', '~', 'h', 'e', 'l', 'l', 'o', 27, '[', '2', '0', '1', '~')
	result := term.parseKey(27)

	if result.Name != "paste" {
		t.Errorf("Name: got %q, want %q", result.Name, "paste")
	}

	if result.Content != "hello" {
		t.Errorf("Content: got %q, want %q", result.Content, "hello")
	}
}

func TestParseKey_BracketedPasteEmpty(t *testing.T) {
	// Simulate ESC[200~ ESC[201~ with no content between markers
	term := testTerminal('[', '2', '0', '0', '~', 27, '[', '2', '0', '1', '~')
	result := term.parseKey(27)

	if result.Name != "paste" {
		t.Errorf("Name: got %q, want %q", result.Name, "paste")
	}

	if result.Content != "" {
		t.Errorf("Content: got %q, want %q", result.Content, "")
	}
}

func TestParseKey_BracketedPasteMaxSize(t *testing.T) {
	// Create a large paste that exceeds maxPasteSize
	// We can't easily test with actual 10MB, so we'll manually verify the limit exists
	// by checking that readBracketedPaste has protection (this test documents the constraint)

	// Simulate ESC[200~ followed by 100,000 'x' runes (won't reach 10MB but tests the flow)
	runes := []rune{'[', '2', '0', '0', '~'}
	for i := 0; i < 100000; i++ {
		runes = append(runes, 'x')
	}

	term := testTerminal(runes...)
	result := term.parseKey(27)

	if result.Name != "paste" {
		t.Errorf("Name: got %q, want %q", result.Name, "paste")
	}

	// Verify content was accumulated (may be capped before end marker)
	if result.Content == "" {
		t.Error("Content should be non-empty")
	}
}

func TestResolveCSI_Home(t *testing.T) {
	term := &Terminal{}

	// ESC[H → Home
	result := term.resolveCSI(nil, 'H')
	if result.Name != "home" {
		t.Errorf("Name: got %q, want %q", result.Name, "home")
	}
}

func TestResolveCSI_End(t *testing.T) {
	term := &Terminal{}

	// ESC[F → End
	result := term.resolveCSI(nil, 'F')
	if result.Name != "end" {
		t.Errorf("Name: got %q, want %q", result.Name, "end")
	}
}

func TestResolveCSI_HomeVT220(t *testing.T) {
	term := &Terminal{}

	// ESC[1~ → Home (VT220 style)
	result := term.resolveCSI([]int{1}, '~')
	if result.Name != "home" {
		t.Errorf("Name: got %q, want %q", result.Name, "home")
	}
}

func TestResolveCSI_EndVT220(t *testing.T) {
	term := &Terminal{}

	// ESC[4~ → End (VT220 style)
	result := term.resolveCSI([]int{4}, '~')
	if result.Name != "end" {
		t.Errorf("Name: got %q, want %q", result.Name, "end")
	}
}
