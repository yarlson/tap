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
