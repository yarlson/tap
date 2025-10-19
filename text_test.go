package tap

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestStyledText_RendersWithSymbolAndBars(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)

	go func() {
		result := Text(context.Background(), TextOptions{
			Message: "Enter your name:",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("h", Key{Name: "h"})
	mock.EmitKeypress("i", Key{Name: "i"})
	mock.EmitKeypress("", Key{Name: "return"})

	result := <-done

	// Should return the typed value
	if result != "hi" {
		t.Errorf("Expected 'hi', got %v", result)
	}

	frames := out.GetFrames()
	if len(frames) == 0 {
		t.Fatal("Expected output frames")
	}

	// Find the frame with the submit content (not the final cursor control frame)
	var submitFrame string

	for _, frame := range frames {
		if strings.Contains(frame, "◇") && strings.Contains(frame, "Enter your name:") {
			submitFrame = frame
			break
		}
	}

	if submitFrame == "" {
		t.Fatal("Could not find submit frame with content")
	}

	// Should contain styled elements: symbol, bars, and final value
	if !strings.Contains(submitFrame, "◇") { // submit symbol
		t.Error("Expected submit symbol ◇ in submit frame")
	}

	if !strings.Contains(submitFrame, "│") { // bar
		t.Error("Expected bar │ in submit frame")
	}

	if !strings.Contains(submitFrame, "Enter your name:") {
		t.Error("Expected message in submit frame")
	}
}

func TestStyledText_ShowsPlaceholderWhenEmpty(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)

	go func() {
		result := Text(context.Background(), TextOptions{
			Message:     "Enter text:",
			Placeholder: "Type something...",
			Input:       mock,
			Output:      out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", Key{Name: "return"})
	<-done

	frames := out.GetFrames()

	// Find a frame that should show the placeholder (with or without ANSI codes)
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "ype something...") { // Look for part of placeholder text
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected placeholder to be shown when input is empty")
	}
}

func TestStyledText_ShowsCursorDuringTyping(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)

	go func() {
		result := Text(context.Background(), TextOptions{
			Message: "Type:",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("a", Key{Name: "a"})
	mock.EmitKeypress("", Key{Name: "return"})
	<-done

	frames := out.GetFrames()

	// Should show the active state with cyan bars
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "◆") && strings.Contains(frame, "│") { // active symbol and cyan bar
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected active state with symbols during typing")
	}
}

func TestStyledText_ShowsErrorState(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	validator := func(val string) error {
		if len(val) < 3 {
			return &ValidationError{Message: "Too short"}
		}

		return nil
	}

	done := make(chan any, 1)

	go func() {
		result := Text(context.Background(), TextOptions{
			Message:  "Enter at least 3 chars:",
			Validate: validator,
			Input:    mock,
			Output:   out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("a", Key{Name: "a"})
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", Key{Name: "return"}) // This should trigger validation error
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("b", Key{Name: "b"})
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("c", Key{Name: "c"})
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", Key{Name: "return"}) // This should succeed
	<-done

	frames := out.GetFrames()

	// Should show error state with error symbol
	foundSymbol := false

	for _, frame := range frames {
		if strings.Contains(frame, "▲") { // error symbol
			foundSymbol = true
			break
		}
	}

	if !foundSymbol {
		t.Error("Expected error symbol ▲ in frames")
	}
}

func TestStyledText_ShowsDefaultValue(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)

	go func() {
		result := Text(context.Background(), TextOptions{
			Message:      "Enter name:",
			DefaultValue: "John",
			Input:        mock,
			Output:       out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", Key{Name: "return"})

	result := <-done

	if result != "John" {
		t.Errorf("Expected default value 'John', got %v", result)
	}
}

func TestStyledText_HandlesPastedText(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)

	go func() {
		result := Text(context.Background(), TextOptions{
			Message: "Enter text:",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)

	// Simulate paste: many rapid single-character events (how terminals actually work)
	// Test with 300+ characters to verify unbounded queue (smaller for race detector performance)
	pastedText := ""
	for i := 0; i < 5; i++ {
		pastedText += "Hello, this is a longer pasted text to test the behavior! "
	}

	for _, ch := range pastedText {
		mock.EmitKeypress(string(ch), Key{Name: string(ch)})
	}

	// Wait longer to ensure all events are processed before submitting
	time.Sleep(100 * time.Millisecond)
	mock.EmitKeypress("", Key{Name: "return"})

	select {
	case result := <-done:
		// Should return all pasted characters
		if result != pastedText {
			t.Errorf("Expected '%s', got '%v'", pastedText, result)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - app likely hung during paste simulation")
	}
}
