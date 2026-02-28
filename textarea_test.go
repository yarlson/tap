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
