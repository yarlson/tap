package prompts

import (
	"strings"
	"testing"
	"time"

	"github.com/yarlson/tap/internal/core"
)

func TestStyledText_RendersWithSymbolAndBars(t *testing.T) {
	mock := core.NewMockReadable()
	out := core.NewMockWritable()

	done := make(chan any, 1)
	go func() {
		result := Text(TextOptions{
			Message: "Enter your name:",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("h", core.Key{Name: "h"})
	mock.EmitKeypress("i", core.Key{Name: "i"})
	mock.EmitKeypress("", core.Key{Name: "return"})

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
	mock := core.NewMockReadable()
	out := core.NewMockWritable()

	done := make(chan any, 1)
	go func() {
		result := Text(TextOptions{
			Message:     "Enter text:",
			Placeholder: "Type something...",
			Input:       mock,
			Output:      out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", core.Key{Name: "return"})
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
	mock := core.NewMockReadable()
	out := core.NewMockWritable()

	done := make(chan any, 1)
	go func() {
		result := Text(TextOptions{
			Message: "Type:",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("a", core.Key{Name: "a"})
	mock.EmitKeypress("", core.Key{Name: "return"})
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
	mock := core.NewMockReadable()
	out := core.NewMockWritable()

	validator := func(val string) error {
		if len(val) < 3 {
			return &core.ValidationError{Message: "Too short"}
		}
		return nil
	}

	done := make(chan any, 1)
	go func() {
		result := Text(TextOptions{
			Message:  "Enter at least 3 chars:",
			Validate: validator,
			Input:    mock,
			Output:   out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("a", core.Key{Name: "a"})
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", core.Key{Name: "return"}) // This should trigger validation error
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("b", core.Key{Name: "b"})
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("c", core.Key{Name: "c"})
	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", core.Key{Name: "return"}) // This should succeed
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
	mock := core.NewMockReadable()
	out := core.NewMockWritable()

	done := make(chan any, 1)
	go func() {
		result := Text(TextOptions{
			Message:      "Enter name:",
			DefaultValue: "John",
			Input:        mock,
			Output:       out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", core.Key{Name: "return"})
	result := <-done

	if result != "John" {
		t.Errorf("Expected default value 'John', got %v", result)
	}
}
