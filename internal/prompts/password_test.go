package prompts

import (
	"strings"
	"testing"
	"time"
)

func TestStyledPassword_RendersWithSymbolBarsAndMasksValue(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)
	go func() {
		result := Password(PasswordOptions{
			Message: "Enter password:",
			Input:   in,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	in.EmitKeypress("h", Key{Name: "h"})
	in.EmitKeypress("i", Key{Name: "i"})
	in.EmitKeypress("", Key{Name: "return"})
	result := <-done

	// Should return the typed value
	if result != "hi" {
		t.Fatalf("expected 'hi', got %#v", result)
	}

	frames := out.GetFrames()
	if len(frames) == 0 {
		t.Fatal("expected output frames")
	}

	// Find a submit frame with bars and symbol, and ensure masked bullets present and raw text absent
	var submitFrame string
	for _, f := range frames {
		if strings.Contains(f, "◇") && strings.Contains(f, "Enter password:") {
			submitFrame = f
			break
		}
	}
	if submitFrame == "" {
		t.Fatal("could not find submit frame")
	}
	if !strings.Contains(submitFrame, "│") {
		t.Error("expected bar │ in submit frame")
	}
	if strings.Contains(submitFrame, "hi") {
		t.Error("password should not render raw text in submit frame")
	}
	if !strings.Contains(submitFrame, "●") {
		t.Error("expected masked bullets in submit frame")
	}
}

func TestStyledPassword_ShowsBulletsDuringTyping(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)
	go func() {
		done <- Password(PasswordOptions{
			Message: "Password:",
			Input:   in,
			Output:  out,
		})
	}()

	time.Sleep(time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	<-done

	frames := out.GetFrames()
	found := false
	for _, f := range frames {
		if strings.Contains(f, "◆") && strings.Contains(f, "│") && strings.Contains(f, "●") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected active state with masked bullets during typing")
	}
}

func TestStyledPassword_ShowsErrorState(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	validator := func(s string) error {
		if len(s) < 3 {
			return &ValidationError{Message: "Too short"}
		}
		return nil
	}

	done := make(chan any, 1)
	go func() {
		done <- Password(PasswordOptions{
			Message:  "Enter:",
			Validate: validator,
			Input:    in,
			Output:   out,
		})
	}()

	time.Sleep(time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"}) // should trigger error
	time.Sleep(time.Millisecond)
	in.EmitKeypress("b", Key{Name: "b"})
	in.EmitKeypress("c", Key{Name: "c"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"}) // should submit now
	<-done

	frames := out.GetFrames()
	foundError := false
	for _, f := range frames {
		if strings.Contains(f, "▲") { // error symbol
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("expected error symbol ▲ in frames")
	}
}
