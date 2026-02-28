package tap

import (
	"context"
	"fmt"
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

// --- TASK3: Validation, Navigation, Edge Case Tests ---

func TestTextarea_ValidationRejectsAndRecovers(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
			Validate: func(s string) error {
				if len(s) < 10 {
					return fmt.Errorf("at least 10 characters required")
				}
				return nil
			},
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "hi" (too short)
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	time.Sleep(5 * time.Millisecond)

	// Try to submit — should fail validation
	in.EmitKeypress("", Key{Name: "return"})
	time.Sleep(10 * time.Millisecond)

	// Check that error is visible in frames
	frames := out.GetFrames()
	foundError := false
	for _, frame := range frames {
		if strings.Contains(frame, "at least 10 characters required") {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("expected validation error message in frames")
	}

	// Check that error symbol is visible
	foundErrorSymbol := false
	for _, frame := range frames {
		if strings.Contains(frame, StepError) {
			foundErrorSymbol = true
			break
		}
	}
	if !foundErrorSymbol {
		t.Error("expected error symbol (▲) in frames")
	}

	// Type more text to clear error and meet validation
	for _, r := range "hello world" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}
	time.Sleep(5 * time.Millisecond)

	// Submit again — should succeed
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "hihello world" {
		t.Fatalf("expected %q, got %q", "hihello world", got)
	}
}

func TestTextarea_ValidationOnResolvedString(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
			Validate: func(s string) error {
				if len(s) < 6 {
					return fmt.Errorf("at least 6 characters required")
				}
				return nil
			},
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "a" then paste "bcdef" — resolved string is "abcdef" (6 chars)
	in.EmitKeypress("a", Key{Name: "a", Rune: 'a'})
	in.EmitPaste("bcdef")
	time.Sleep(5 * time.Millisecond)

	// Submit — validation runs against resolved string "abcdef" (6 chars, passes)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "abcdef" {
		t.Fatalf("expected %q, got %q", "abcdef", got)
	}
}

func TestTextarea_ErrorClearsOnKeypress(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
			Validate: func(s string) error {
				if len(s) < 5 {
					return fmt.Errorf("too short")
				}
				return nil
			},
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "hi" and submit to trigger error
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	time.Sleep(10 * time.Millisecond)

	// Type a character to clear error
	in.EmitKeypress("x", Key{Name: "x", Rune: 'x'})
	time.Sleep(10 * time.Millisecond)

	// Verify active symbol is back (error cleared)
	frames := out.GetFrames()
	// Find the last frame containing "hix" (after error clear + new char)
	foundActive := false
	for i := len(frames) - 1; i >= 0; i-- {
		if strings.Contains(frames[i], "hix") && strings.Contains(frames[i], StepActive) {
			foundActive = true
			break
		}
	}
	if !foundActive {
		t.Error("expected active state after typing a character post-error")
	}

	// Type more and submit
	for _, r := range "ab" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "hixab" {
		t.Fatalf("expected %q, got %q", "hixab", got)
	}
}

func TestTextarea_HomeKey(t *testing.T) {
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

	// Type "hello"
	for _, r := range "hello" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Home → cursor at 0
	in.EmitKeypress("", Key{Name: "home"})

	// Type "X" → "Xhello"
	in.EmitKeypress("X", Key{Name: "X", Rune: 'X'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "Xhello" {
		t.Fatalf("expected %q, got %q", "Xhello", got)
	}
}

func TestTextarea_EndKey(t *testing.T) {
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

	// Type "hello"
	for _, r := range "hello" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Home → cursor at 0, then End → cursor at 5
	in.EmitKeypress("", Key{Name: "home"})
	in.EmitKeypress("", Key{Name: "end"})

	// Type "X" → "helloX"
	in.EmitKeypress("X", Key{Name: "X", Rune: 'X'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "helloX" {
		t.Fatalf("expected %q, got %q", "helloX", got)
	}
}

func TestTextarea_HomeEndMultiline(t *testing.T) {
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

	// Type "hello", Shift+Enter, "world"
	for _, r := range "hello" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}
	in.EmitKeypress("", Key{Name: "return", Shift: true})
	for _, r := range "world" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}

	// Cursor is at end of "world" (line 2, col 5)
	// Move left twice to mid-line-2 (col 3)
	in.EmitKeypress("", Key{Name: "left"})
	in.EmitKeypress("", Key{Name: "left"})

	// Home → cursor should be at start of line 2 (not line 1)
	in.EmitKeypress("", Key{Name: "home"})

	// Type "X" → "hello\nXworld" (not "Xhello\nworld")
	in.EmitKeypress("X", Key{Name: "X", Rune: 'X'})

	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "hello\nXworld" {
		t.Fatalf("expected %q, got %q", "hello\nXworld", got)
	}
}

func TestTextarea_PasteWithNewlines(t *testing.T) {
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
	in.EmitPaste("line1\nline2\nline3")
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "line1\nline2\nline3" {
		t.Fatalf("expected %q, got %q", "line1\nline2\nline3", got)
	}
}

func TestTextarea_EmptySubmitNoDefault(t *testing.T) {
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

func TestTextarea_CancelAfterPaste(t *testing.T) {
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
	in.EmitPaste("content")
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "escape"})

	got := <-resultCh
	if got != "" {
		t.Fatalf("expected empty string on cancel after paste, got %q", got)
	}
}

func TestTextarea_ThreePastes(t *testing.T) {
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
	in.EmitPaste("alpha")
	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("beta")
	time.Sleep(5 * time.Millisecond)
	in.EmitPaste("gamma")
	time.Sleep(5 * time.Millisecond)

	// Verify all three placeholders appear
	frames := out.GetFrames()
	for _, label := range []string{"[Text 1]", "[Text 2]", "[Text 3]"} {
		found := false
		for _, frame := range frames {
			if strings.Contains(frame, label) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s placeholder in rendered frames", label)
		}
	}

	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	if got != "alphabetagamma" {
		t.Fatalf("expected %q, got %q", "alphabetagamma", got)
	}
}

func TestTextarea_ContextCancel(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	ctx, cancel := context.WithCancel(context.Background())

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(ctx, TextareaOptions{
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

	// Cancel context
	cancel()

	got := <-resultCh
	if got != "" {
		t.Fatalf("expected empty string on context cancel, got %q", got)
	}
}

// --- TASK3: Unit tests ---

func TestResolve_PasteWithNewlines(t *testing.T) {
	buf := []rune{'x', idToPUA(1), 'y'}
	pastes := map[int]string{1: "a\nb\nc"}

	got := resolve(buf, pastes)
	if got != "xa\nb\ncy" {
		t.Fatalf("expected %q, got %q", "xa\nb\ncy", got)
	}
}

func TestCursorNavigation_Home(t *testing.T) {
	buf := []rune("hello\nworld")
	// Cursor at col 3, line 2 → "hello\nwor|ld"
	// Position: 'h'=0, 'e'=1, 'l'=2, 'l'=3, 'o'=4, '\n'=5, 'w'=6, 'o'=7, 'r'=8, 'l'=9, 'd'=10
	cursor := 9 // at 'l' in "world" (col 3)

	line, _ := cursorToLineCol(buf, cursor)
	if line != 1 {
		t.Fatalf("expected line 1, got %d", line)
	}

	// Home → start of line 2 (index 6)
	newCursor := lineColToCursor(buf, line, 0)
	if newCursor != 6 {
		t.Fatalf("expected cursor at 6 (start of line 2), got %d", newCursor)
	}
}

func TestCursorNavigation_End(t *testing.T) {
	buf := []rune("hello\nworld")
	// Cursor at col 0, line 1 → "hello\n|world"
	cursor := 6 // at 'w' in "world" (col 0, line 1)

	line, _ := cursorToLineCol(buf, cursor)
	if line != 1 {
		t.Fatalf("expected line 1, got %d", line)
	}

	// End → end of line 1 (index 11, end of "world")
	newCursor := lineColToCursor(buf, line, len(buf))
	if newCursor != 11 {
		t.Fatalf("expected cursor at 11 (end of line 2), got %d", newCursor)
	}
}

func TestTextarea_ShiftReturnClearsError(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	resultCh := make(chan string, 1)

	go func() {
		res := Textarea(context.Background(), TextareaOptions{
			Message: "Enter text:",
			Input:   in,
			Output:  out,
			Validate: func(s string) error {
				if len(s) < 5 {
					return fmt.Errorf("too short")
				}
				return nil
			},
		})
		resultCh <- res
	}()

	time.Sleep(5 * time.Millisecond)

	// Type "hi" and submit to trigger error
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	time.Sleep(10 * time.Millisecond)

	// Press Shift+Return to insert newline — should clear error
	in.EmitKeypress("", Key{Name: "return", Shift: true})
	time.Sleep(10 * time.Millisecond)

	// Verify active symbol is back (error cleared after Shift+Return)
	frames := out.GetFrames()
	foundActive := false
	for i := len(frames) - 1; i >= 0; i-- {
		if strings.Contains(frames[i], "hi") && strings.Contains(frames[i], StepActive) {
			foundActive = true
			break
		}
	}
	if !foundActive {
		t.Error("expected active state after Shift+Return to clear error")
	}

	// Type more to complete validation and submit
	for _, r := range "jkl" {
		in.EmitKeypress(string(r), Key{Name: string(r), Rune: r})
	}
	time.Sleep(5 * time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})

	got := <-resultCh
	// Expected: "hi\njkl" (newline inserted by Shift+Return, then "jkl" typed)
	if got != "hi\njkl" {
		t.Fatalf("expected %q, got %q", "hi\njkl", got)
	}
}
