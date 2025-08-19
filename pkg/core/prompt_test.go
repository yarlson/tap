package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrompt_RendersRenderResult(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	// Start the prompt in a goroutine since it blocks
	go func() {
		p.Prompt()
	}()

	// Give a small delay to allow initial render
	// In production this wouldn't be needed as the prompt would handle timing properly
	time.Sleep(time.Millisecond)

	// Cancel to exit the prompt
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	expected := []string{"\x1b[?25l", "foo", "\n", "\x1b[?25h"} // cursor.hide + "foo" + newline + cursor.show
	assert.Equal(t, expected, output.Buffer)
}

func TestPrompt_SubmitsOnReturn(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	// Start the prompt
	resultCh := make(chan any)
	go func() {
		result := p.Prompt()
		resultCh <- result
	}()

	// Give a small delay to allow initial render
	time.Sleep(time.Millisecond)

	// Simulate return key press
	input.EmitKeypress("", Key{Name: "return"})

	// Wait for result
	result := <-resultCh

	assert.Equal(t, nil, result)
	assert.False(t, IsCancel(result))
	assert.Equal(t, StateSubmit, p.State)

	expectedOutput := []string{"\x1b[?25l", "foo", "\n", "\x1b[?25h"} // cursor.hide + "foo" + "\n" + cursor.show
	assert.Equal(t, expectedOutput, output.Buffer)
}

func TestPrompt_CancelsOnCtrlC(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	// Start the prompt
	resultCh := make(chan any)
	go func() {
		result := p.Prompt()
		resultCh <- result
	}()

	// Give a small delay to allow initial render
	time.Sleep(time.Millisecond)

	// Simulate ctrl-c
	input.EmitKeypress("\x03", Key{Name: "c"})

	// Wait for result
	result := <-resultCh

	assert.True(t, IsCancel(result))
	assert.Equal(t, StateCancel, p.State)

	expectedOutput := []string{"\x1b[?25l", "foo", "\n", "\x1b[?25h"} // cursor.hide + "foo" + "\n" + cursor.show
	assert.Equal(t, expectedOutput, output.Buffer)
}

func TestPrompt_DoesNotWriteInitialValueToValue(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var eventCalled bool
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		InitialValue: "bananas",
	})

	p.On("value", func(value any) {
		eventCalled = true
	})

	go p.Prompt()
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	assert.Equal(t, nil, p.Value)
	assert.False(t, eventCalled)
}

func TestPrompt_ReRendersOnResize(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	renderCallCount := 0
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			renderCallCount++
			return "foo"
		},
	})

	go p.Prompt()
	time.Sleep(time.Millisecond) // Allow initial render

	assert.Equal(t, 1, renderCallCount)

	// Simulate resize event
	output.Emit("resize")
	time.Sleep(time.Millisecond) // Allow re-render

	assert.Equal(t, 2, renderCallCount)

	// Cancel to exit
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_StateIsActiveAfterFirstRender(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	assert.Equal(t, StateInitial, p.State)

	go p.Prompt()
	time.Sleep(time.Millisecond) // Allow initial render

	assert.Equal(t, StateActive, p.State)

	// Cancel to exit
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_EmitsTruthyConfirmOnYPress(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var confirmValue *bool
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	p.On("confirm", func(value bool) {
		confirmValue = &value
	})

	go p.Prompt()
	input.EmitKeypress("y", Key{Name: "y"})
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	assert.NotNil(t, confirmValue)
	assert.True(t, *confirmValue)
}

func TestPrompt_EmitsFalseyConfirmOnNPress(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var confirmValue *bool
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	p.On("confirm", func(value bool) {
		confirmValue = &value
	})

	go p.Prompt()
	input.EmitKeypress("n", Key{Name: "n"})
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	assert.NotNil(t, confirmValue)
	assert.False(t, *confirmValue)
}

func TestPrompt_EmitsKeyEventForUnknownChars(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var keyChar string
	var keyInfo Key
	var eventCount int
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	p.On("key", func(char string, key Key) {
		eventCount++
		if eventCount == 1 { // Capture the first key event (should be 'z')
			keyChar = char
			keyInfo = key
		}
	})

	go p.Prompt()
	input.EmitKeypress("z", Key{Name: "z"})
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	assert.Equal(t, "z", keyChar)
	assert.Equal(t, "z", keyInfo.Name)
}

func TestPrompt_EmitsCursorEventsForMovementKeys(t *testing.T) {
	keys := []string{"up", "down", "left", "right"}

	for _, key := range keys {
		t.Run("key_"+key, func(t *testing.T) {
			input := NewMockReadable()
			output := NewMockWritable()

			var cursorEvent string
			p := NewPrompt(PromptOptions{
				Input:  input,
				Output: output,
				Render: func(p *Prompt) string {
					return "foo"
				},
			})

			p.On("cursor", func(direction string) {
				cursorEvent = direction
			})

			go p.Prompt()
			input.EmitKeypress(key, Key{Name: key})
			input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

			assert.Equal(t, key, cursorEvent)
		})
	}
}

func TestPrompt_ValidatesValueOnReturn(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Validate: func(value any) error {
			if value == "valid" {
				return nil
			}
			return NewValidationError("must be valid")
		},
	})

	go p.Prompt()

	p.Value = "invalid"
	input.EmitKeypress("", Key{Name: "return"})
	// Check state before canceling

	assert.Equal(t, StateError, p.State)
	assert.Equal(t, "must be valid", p.Error)

	// Now cancel to exit the test
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_AcceptsValidValueWithValidation(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Validate: func(value any) error {
			if value == "valid" {
				return nil
			}
			return NewValidationError("must be valid")
		},
	})

	go p.Prompt()

	p.Value = "valid"
	input.EmitKeypress("", Key{Name: "return"})

	assert.Equal(t, StateSubmit, p.State)
	assert.Equal(t, "", p.Error)
}

func TestPrompt_EmitsCursorEventsForMovementKeyAliasesWhenNotTracking(t *testing.T) {
	keys := [][]string{
		{"k", "up"},
		{"j", "down"},
		{"h", "left"},
		{"l", "right"},
	}

	for _, keyPair := range keys {
		alias := keyPair[0]
		expected := keyPair[1]

		t.Run("alias_"+alias, func(t *testing.T) {
			input := NewMockReadable()
			output := NewMockWritable()

			var cursorEvent string
			p := NewPromptWithTracking(PromptOptions{
				Input:  input,
				Output: output,
				Render: func(p *Prompt) string {
					return "foo"
				},
			}, false) // Set tracking to false

			p.On("cursor", func(direction string) {
				cursorEvent = direction
			})

			go p.Prompt()
			time.Sleep(time.Millisecond)

			input.EmitKeypress(alias, Key{Name: alias})
			time.Sleep(time.Millisecond)

			assert.Equal(t, expected, cursorEvent)

			input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
		})
	}
}

func TestPrompt_AbortsOnAbortSignal(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	ctx, cancel := context.WithCancel(context.Background())

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Signal: ctx,
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateActive, p.State)

	cancel()
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateCancel, p.State)
}

func TestPrompt_ReturnsImmediatelyIfSignalIsAlreadyAborted(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Signal: ctx,
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateCancel, p.State)
}

func TestPrompt_AcceptsInvalidInitialValue(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		InitialValue: "invalid",
		Validate: func(value any) error {
			if value == "valid" {
				return nil
			}
			return NewValidationError("must be valid")
		},
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateActive, p.State)
	assert.Equal(t, "", p.Error)

	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_ValidatesValueWithErrorObject(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Validate: func(value any) error {
			if value == "valid" {
				return nil
			}
			return errors.New("must be valid")
		},
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	p.Value = "invalid"
	input.EmitKeypress("", Key{Name: "return"})

	assert.Equal(t, StateError, p.State)
	assert.Equal(t, "must be valid", p.Error)

	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_ValidatesValueWithRegexValidation(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Validate: func(value any) error {
			str, ok := value.(string)
			if !ok {
				str = ""
			}
			// Test for uppercase letters only
			matched := true
			for _, r := range str {
				if r < 'A' || r > 'Z' {
					matched = false
					break
				}
			}
			if matched && len(str) > 0 {
				return nil
			}
			return NewValidationError("Invalid value")
		},
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	p.Value = "Invalid Value $$$"
	input.EmitKeypress("", Key{Name: "return"})

	assert.Equal(t, StateError, p.State)
	assert.Equal(t, "Invalid value", p.Error)

	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_AcceptsValidValueWithRegexValidation(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Validate: func(value any) error {
			str, ok := value.(string)
			if !ok {
				str = ""
			}
			// Test for uppercase letters only
			matched := true
			for _, r := range str {
				if r < 'A' || r > 'Z' {
					matched = false
					break
				}
			}
			if matched && len(str) > 0 {
				return nil
			}
			return NewValidationError("Invalid value")
		},
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	p.Value = "VALID"
	input.EmitKeypress("", Key{Name: "return"})

	assert.Equal(t, StateSubmit, p.State)
	assert.Equal(t, "", p.Error)
}
