package tap

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestTextarea_BasicInput(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("e", Key{Name: "e", Rune: 'e'})
	in.EmitKeypress("l", Key{Name: "l", Rune: 'l'})
	in.EmitKeypress("l", Key{Name: "l", Rune: 'l'})
	in.EmitKeypress("o", Key{Name: "o", Rune: 'o'})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestTextarea_Cancel(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "escape"})

	got := <-resultCh
	if got != "" {
		t.Fatalf("expected empty string on cancel, got %q", got)
	}
}

func TestTextarea_Placeholder(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message:     "Enter text:",
			Placeholder: "Type something...",
			Input:       in,
			Output:      out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Check frames contain placeholder text before submitting
	frames := out.GetFrames()
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "ype something...") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected placeholder text in frames")
	}

	in.EmitKeypress("", Key{Name: "return"})
	<-resultCh
}

func TestTextarea_DefaultValue(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message:      "Enter text:",
			DefaultValue: "default text",
			Input:        in,
			Output:       out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "default text" {
		t.Fatalf("expected 'default text', got %q", got)
	}
}

func TestTextarea_InitialValue(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message:      "Enter text:",
			InitialValue: "initial",
			Input:        in,
			Output:       out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "initial" {
		t.Fatalf("expected 'initial', got %q", got)
	}
}

func TestTextarea_EmptySubmit(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestTextarea_CursorMovement(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "abc"
	in.EmitKeypress("a", Key{Name: "a", Rune: 'a'})
	in.EmitKeypress("b", Key{Name: "b", Rune: 'b'})
	in.EmitKeypress("c", Key{Name: "c", Rune: 'c'})

	// Move left twice
	in.EmitKeypress("", Key{Name: "left"})
	in.EmitKeypress("", Key{Name: "left"})

	// Insert "X"
	in.EmitKeypress("X", Key{Name: "X", Rune: 'X'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "aXbc" {
		t.Fatalf("expected 'aXbc', got %q", got)
	}
}

func TestTextarea_Backspace(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a", Rune: 'a'})
	in.EmitKeypress("b", Key{Name: "b", Rune: 'b'})
	in.EmitKeypress("c", Key{Name: "c", Rune: 'c'})
	in.EmitKeypress("", Key{Name: "backspace"})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "ab" {
		t.Fatalf("expected 'ab', got %q", got)
	}
}

func TestTextarea_RendersBars(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	time.Sleep(5 * time.Millisecond)

	frames := out.GetFrames()
	foundBar := false
	foundSymbol := false

	for _, frame := range frames {
		if strings.Contains(frame, Bar) {
			foundBar = true
		}

		if strings.Contains(frame, StepActive) {
			foundSymbol = true
		}
	}

	if !foundBar {
		t.Error("expected bar prefix in active frames")
	}

	if !foundSymbol {
		t.Error("expected active symbol in frames")
	}

	in.EmitKeypress("", Key{Name: "return"})
	<-resultCh

	// Check submit frame has submit symbol
	frames = out.GetFrames()
	foundSubmit := false

	for _, frame := range frames {
		if strings.Contains(frame, StepSubmit) {
			foundSubmit = true
			break
		}
	}

	if !foundSubmit {
		t.Error("expected submit symbol in frames")
	}
}

func TestTextarea_CancelSymbol(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "escape"})
	<-resultCh

	frames := out.GetFrames()
	foundCancel := false

	for _, frame := range frames {
		if strings.Contains(frame, StepCancel) {
			foundCancel = true
			break
		}
	}

	if !foundCancel {
		t.Error("expected cancel symbol in frames")
	}
}

// --- TASK1: Multiline editing tests ---

func TestTextarea_MultilineInput(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "line1"
	for _, r := range "line1" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Shift+Enter to insert newline
	in.EmitKeypress("", Key{Name: "return", Shift: true})

	// Type "line2"
	for _, r := range "line2" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	time.Sleep(5 * time.Millisecond)

	// Submit
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "line1\nline2" {
		t.Fatalf("expected %q, got %q", "line1\nline2", got)
	}
}

func TestTextarea_UpDownNavigation(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "abc", Shift+Enter, "def"
	for _, r := range "abc" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	in.EmitKeypress("", Key{Name: "return", Shift: true})

	for _, r := range "def" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Cursor is at end of "def" (line 2, col 3)
	// Up arrow → cursor moves to line 1, col 3 (end of "abc")
	in.EmitKeypress("", Key{Name: "up"})

	// Insert "X" at cursor position (after "abc")
	in.EmitKeypress("X", Key{Name: "X", Rune: 'X'})

	// Down arrow → cursor moves back to line 2
	in.EmitKeypress("", Key{Name: "down"})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "abcX\ndef" {
		t.Fatalf("expected %q, got %q", "abcX\ndef", got)
	}
}

func TestTextarea_ColumnClamp(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "longline" (8 chars), Shift+Enter, "ab" (2 chars)
	for _, r := range "longline" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	in.EmitKeypress("", Key{Name: "return", Shift: true})

	for _, r := range "ab" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Cursor at end of "ab" (line 2, col 2)
	// Up arrow → should clamp to end of... wait, line 1 is longer, so col 2 is fine
	// Let me reverse: cursor on long line, move down to short line
	// Go up first
	in.EmitKeypress("", Key{Name: "up"})
	// Now on line 1, col 2 — cursor is at position 2 in "longline"
	// Move to end of line 1
	for i := 0; i < 6; i++ {
		in.EmitKeypress("", Key{Name: "right"})
	}

	// Now at col 8 in "longline", move down to "ab" (length 2)
	in.EmitKeypress("", Key{Name: "down"})

	// Cursor should be clamped to col 2 (end of "ab")
	// Insert "Z" — should appear at end of "ab"
	in.EmitKeypress("Z", Key{Name: "Z", Rune: 'Z'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "longline\nabZ" {
		t.Fatalf("expected %q, got %q", "longline\nabZ", got)
	}
}

func TestTextarea_RendersBarsPerLine(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type two lines
	for _, r := range "first" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	in.EmitKeypress("", Key{Name: "return", Shift: true})

	for _, r := range "second" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	time.Sleep(5 * time.Millisecond)

	// Check frames for bar prefix on each content line
	frames := out.GetFrames()
	found := false

	for _, frame := range frames {
		// A multiline frame should have at least two lines with Bar prefix containing content
		if strings.Contains(frame, "first") && strings.Contains(frame, "second") {
			lines := strings.Split(frame, "\n")
			barCount := 0

			for _, line := range lines {
				if strings.Contains(line, Bar) {
					barCount++
				}
			}

			// Should have at least 3 bars: separator, first line, second line
			if barCount >= 3 {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("expected bar prefix on each content line in multiline frame")
	}

	in.EmitKeypress("", Key{Name: "return"})
	<-resultCh
}

func TestTextarea_MultilineSubmitDisplay(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	for _, r := range "hello" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	in.EmitKeypress("", Key{Name: "return", Shift: true})

	for _, r := range "world" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	<-resultCh

	// After submit, frames should contain submit symbol and both lines with bar prefix
	frames := out.GetFrames()
	foundSubmitWithBars := false

	for _, frame := range frames {
		if !strings.Contains(frame, StepSubmit) || !strings.Contains(frame, "hello") || !strings.Contains(frame, "world") {
			continue
		}

		// Verify each content line has a bar prefix
		lines := strings.Split(frame, "\n")
		helloHasBar := false
		worldHasBar := false

		for _, line := range lines {
			if strings.Contains(line, Bar) && strings.Contains(line, "hello") {
				helloHasBar = true
			}

			if strings.Contains(line, Bar) && strings.Contains(line, "world") {
				worldHasBar = true
			}
		}

		if helloHasBar && worldHasBar {
			foundSubmitWithBars = true
			break
		}
	}

	if !foundSubmitWithBars {
		t.Error("expected submitted frame to show both lines with bar prefix and submit symbol")
	}
}

func TestTextarea_MultilineCancelDisplay(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	for _, r := range "hello" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	in.EmitKeypress("", Key{Name: "return", Shift: true})

	for _, r := range "world" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "escape"})
	<-resultCh

	// After cancel, frames should contain cancel symbol and both lines with bar prefix
	frames := out.GetFrames()
	foundCancelWithBars := false

	for _, frame := range frames {
		if !strings.Contains(frame, StepCancel) || !strings.Contains(frame, "hello") || !strings.Contains(frame, "world") {
			continue
		}

		// Verify each content line has a bar prefix
		lines := strings.Split(frame, "\n")
		helloHasBar := false
		worldHasBar := false

		for _, line := range lines {
			if strings.Contains(line, Bar) && strings.Contains(line, "hello") {
				helloHasBar = true
			}

			if strings.Contains(line, Bar) && strings.Contains(line, "world") {
				worldHasBar = true
			}
		}

		if helloHasBar && worldHasBar {
			foundCancelWithBars = true
			break
		}
	}

	if !foundCancelWithBars {
		t.Error("expected cancelled frame to show both lines with bar prefix and cancel symbol")
	}
}

// --- TASK2: Paste buffer unit tests ---

func TestResolve_NoPUA(t *testing.T) {
	buf := []rune("hello world")
	pastes := map[int]string{}

	got := resolve(buf, pastes)
	if got != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", got)
	}
}

func TestResolve_WithPUA(t *testing.T) {
	buf := []rune{'h', 'i', idToPUA(1), '!'}
	pastes := map[int]string{1: "pasted content"}

	got := resolve(buf, pastes)
	if got != "hipasted content!" {
		t.Fatalf("expected %q, got %q", "hipasted content!", got)
	}
}

func TestResolve_MultiplePUA(t *testing.T) {
	buf := []rune{'a', idToPUA(1), 'b', idToPUA(2), 'c'}
	pastes := map[int]string{1: "X", 2: "YZ"}

	got := resolve(buf, pastes)
	if got != "aXbYZc" {
		t.Fatalf("expected %q, got %q", "aXbYZc", got)
	}
}

func TestIsPUA(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected bool
	}{
		{"PUA start", 0xE000, true},
		{"PUA middle", 0xE100, true},
		{"PUA end", 0xF8FF, true},
		{"regular letter", 'a', false},
		{"space", ' ', false},
		{"newline", '\n', false},
		{"zero", 0, false},
		{"just below PUA", 0xDFFF, false},
		{"just above PUA", 0xF900, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPUA(tt.r)
			if got != tt.expected {
				t.Errorf("isPUA(%U) = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

func TestPUARoundtrip(t *testing.T) {
	for id := 1; id <= 10; id++ {
		r := idToPUA(id)
		if !isPUA(r) {
			t.Errorf("idToPUA(%d) produced non-PUA rune %U", id, r)
		}

		gotID := puaToID(r)
		if gotID != id {
			t.Errorf("puaToID(idToPUA(%d)) = %d, want %d", id, gotID, id)
		}
	}
}

// --- TASK2: Paste buffer integration tests ---

func TestTextarea_PasteInsertsPlaceholder(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("hello")
	time.Sleep(5 * time.Millisecond)

	frames := out.GetFrames()
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "[Text 1]") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected [Text 1] placeholder in rendered frames")
	}

	in.EmitKeypress("", Key{Name: "return"})
	<-resultCh
}

func TestTextarea_PasteAndResolve(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("hello")
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestTextarea_MultiplePastes(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("first")
	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("second")
	time.Sleep(5 * time.Millisecond)

	// Check that both placeholders appear in frames
	frames := out.GetFrames()
	foundText1 := false
	foundText2 := false

	for _, frame := range frames {
		if strings.Contains(frame, "[Text 1]") {
			foundText1 = true
		}

		if strings.Contains(frame, "[Text 2]") {
			foundText2 = true
		}
	}

	if !foundText1 {
		t.Error("expected [Text 1] placeholder in rendered frames")
	}

	if !foundText2 {
		t.Error("expected [Text 2] placeholder in rendered frames")
	}

	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "firstsecond" {
		t.Fatalf("expected %q, got %q", "firstsecond", got)
	}
}

func TestTextarea_AtomicPlaceholderDelete(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("pasted")
	time.Sleep(5 * time.Millisecond)

	// Backspace should remove the entire placeholder
	in.EmitKeypress("", Key{Name: "backspace"})
	time.Sleep(5 * time.Millisecond)

	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "" {
		t.Fatalf("expected empty string after deleting paste placeholder, got %q", got)
	}
}

func TestTextarea_CursorSkipsPlaceholder(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "ab"
	in.EmitKeypress("a", Key{Name: "a", Rune: 'a'})
	in.EmitKeypress("b", Key{Name: "b", Rune: 'b'})

	// Paste "X"
	in.EmitPaste("X")

	// Type "cd"
	in.EmitKeypress("c", Key{Name: "c", Rune: 'c'})
	in.EmitKeypress("d", Key{Name: "d", Rune: 'd'})

	// Cursor is at end: a b [PUA] c d
	// Move left twice → should be at position after 'b' (skipping over PUA)
	in.EmitKeypress("", Key{Name: "left"})
	in.EmitKeypress("", Key{Name: "left"})

	// Insert "Z" — should appear between 'b' and placeholder
	in.EmitKeypress("Z", Key{Name: "Z", Rune: 'Z'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "abZXcd" {
		t.Fatalf("expected %q, got %q", "abZXcd", got)
	}
}

func TestTextarea_TypedAndPastedMixed(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	in.EmitKeypress("a", Key{Name: "a", Rune: 'a'})
	in.EmitPaste("B")
	in.EmitKeypress("c", Key{Name: "c", Rune: 'c'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "aBc" {
		t.Fatalf("expected %q, got %q", "aBc", got)
	}
}

func TestTextarea_PasteDeleteCleansStore(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "x", paste, then delete at placeholder position
	in.EmitKeypress("x", Key{Name: "x", Rune: 'x'})
	in.EmitPaste("DELETED")
	time.Sleep(5 * time.Millisecond)

	// Move left to position cursor before the PUA rune, then use delete
	in.EmitKeypress("", Key{Name: "left"})
	// Now cursor is before PUA, delete forward removes it
	in.EmitKeypress("", Key{Name: "delete"})
	time.Sleep(5 * time.Millisecond)

	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "x" {
		t.Fatalf("expected %q after deleting paste placeholder with delete key, got %q", "x", got)
	}
}

func TestTextarea_BracketedPasteMode(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Check that bracketed paste mode enable sequence was written to output
	frames := out.GetFrames()
	foundEnable := false

	for _, frame := range frames {
		if strings.Contains(frame, "\x1b[?2004h") {
			foundEnable = true
			break
		}
	}

	if !foundEnable {
		t.Error("expected bracketed paste enable sequence (ESC[?2004h) in output")
	}

	in.EmitKeypress("", Key{Name: "return"})
	<-resultCh

	// Check that bracketed paste mode disable sequence was written on finalize
	frames = out.GetFrames()
	foundDisable := false

	for _, frame := range frames {
		if strings.Contains(frame, "\x1b[?2004l") {
			foundDisable = true
			break
		}
	}

	if !foundDisable {
		t.Error("expected bracketed paste disable sequence (ESC[?2004l) in output")
	}
}

func TestTextarea_CursorRightSkipsPlaceholder(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "ab"
	in.EmitKeypress("a", Key{Name: "a", Rune: 'a'})
	in.EmitKeypress("b", Key{Name: "b", Rune: 'b'})

	// Move left to before 'b'
	in.EmitKeypress("", Key{Name: "left"})

	// Paste "X" — inserts PUA at cursor (between a and b)
	in.EmitPaste("X")

	// Cursor is now after PUA at position 2 (between PUA and b)
	// Move left to before PUA (between a and PUA)
	in.EmitKeypress("", Key{Name: "left"})

	// Cursor is before PUA. Type "Z"
	in.EmitKeypress("Z", Key{Name: "Z", Rune: 'Z'})

	// Move right over PUA — should land between PUA and 'b' (position 2)
	in.EmitKeypress("", Key{Name: "right"})

	// Type "c" — should appear after placeholder and before 'b'
	in.EmitKeypress("c", Key{Name: "c", Rune: 'c'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	// Expected: a Z [PUA=X] c b
	if got != "aZXcb" {
		t.Fatalf("expected %q, got %q", "aZXcb", got)
	}
}

func TestTextarea_ShiftEnterMidLine(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "helloworld"
	for _, r := range "helloworld" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Move left 5 times to position cursor between "hello" and "world"
	for i := 0; i < 5; i++ {
		in.EmitKeypress("", Key{Name: "left"})
	}

	// Shift+Enter splits the line
	in.EmitKeypress("", Key{Name: "return", Shift: true})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", got)
	}
}
