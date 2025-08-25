package tap

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestStyledConfirm_RendersWithRadioButtons(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan bool, 1)
	go func() {
		result := Confirm(context.Background(), ConfirmOptions{
			Message: "Continue?",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("y", Key{Name: "y"})
	result := <-done

	if result != true {
		t.Errorf("Expected true, got %v", result)
	}

	frames := out.GetFrames()
	if len(frames) == 0 {
		t.Fatal("Expected output frames")
	}

	// Should show radio buttons in some frame
	found := false
	for _, frame := range frames {
		if strings.Contains(frame, "●") || strings.Contains(frame, "○") { // active/inactive radio
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected radio button symbols in frames")
	}
}

func TestStyledConfirm_ShowsActiveInactiveOptions(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan bool, 1)
	go func() {
		result := Confirm(context.Background(), ConfirmOptions{
			Message:  "Delete file?",
			Active:   "Delete",
			Inactive: "Keep",
			Input:    mock,
			Output:   out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("n", Key{Name: "n"})
	<-done

	frames := out.GetFrames()

	// Should show custom active/inactive labels
	foundActive := false
	foundInactive := false
	for _, frame := range frames {
		if strings.Contains(frame, "Delete") {
			foundActive = true
		}
		if strings.Contains(frame, "Keep") {
			foundInactive = true
		}
	}

	if !foundActive {
		t.Error("Expected 'Delete' label in frames")
	}
	if !foundInactive {
		t.Error("Expected 'Keep' label in frames")
	}
}

func TestStyledConfirm_ShowsSymbolsAndBars(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan bool, 1)
	go func() {
		result := Confirm(context.Background(), ConfirmOptions{
			Message: "Proceed?",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("y", Key{Name: "y"})
	<-done

	frames := out.GetFrames()

	// Should contain styled elements: symbol and bars
	foundSymbol := false
	foundBar := false
	for _, frame := range frames {
		if strings.Contains(frame, "◆") || strings.Contains(frame, "◇") { // active or submit symbol
			foundSymbol = true
		}
		if strings.Contains(frame, "│") { // bar
			foundBar = true
		}
	}

	if !foundSymbol {
		t.Error("Expected prompt symbol in frames")
	}
	if !foundBar {
		t.Error("Expected bar symbol in frames")
	}
}

func TestStyledConfirm_ShowsInitialValue(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan bool, 1)
	go func() {
		result := Confirm(context.Background(), ConfirmOptions{
			Message:      "Continue?",
			InitialValue: true, // Start with Yes selected
			Input:        mock,
			Output:       out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("", Key{Name: "left"})
	time.Sleep(time.Millisecond) // Give time for the value to update
	mock.EmitKeypress("", Key{Name: "return"})
	result := <-done

	// After pressing left arrow, should be false
	if result != false {
		t.Errorf("Expected false after toggling from true, got %v", result)
	}
}

func TestStyledConfirm_ShowsCancelState(t *testing.T) {
	mock := NewMockReadable()
	out := NewMockWritable()

	done := make(chan any, 1)
	go func() {
		result := Confirm(context.Background(), ConfirmOptions{
			Message: "Continue?",
			Input:   mock,
			Output:  out,
		})
		done <- result
	}()

	time.Sleep(time.Millisecond)
	mock.EmitKeypress("\x03", Key{Ctrl: true, Name: "c"}) // Ctrl+C
	result := <-done
	// typed API: cancel returns false
	if result != false {
		t.Error("Expected false on cancel")
	}

	frames := out.GetFrames()

	// Should show cancel state
	found := false
	for _, frame := range frames {
		if strings.Contains(frame, "■") { // cancel symbol
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected cancel symbol ■ in frames")
	}
}
