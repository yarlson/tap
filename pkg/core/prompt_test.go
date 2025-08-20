package core

import (
	"context"
	"errors"
	"sync/atomic"
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

	// Start the prompt in a goroutine
	done := make(chan any, 1)
	go func() {
		done <- p.Prompt()
	}()

	// Small delay to allow initial render
	time.Sleep(time.Millisecond)

	// Cancel to exit the prompt
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	<-done

	expected := []string{"\x1b[?25l", "foo", "\r\n", "\x1b[?25h"} // cursor.hide + "foo" + newline + cursor.show
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

	// Small delay to allow initial render
	time.Sleep(time.Millisecond)

	// Simulate return key press
	input.EmitKeypress("", Key{Name: "return"})

	// Wait for result
	result := <-resultCh

	assert.Equal(t, nil, result)
	assert.False(t, IsCancel(result))
	assert.Equal(t, StateSubmit, p.StateSnapshot())

	expectedOutput := []string{"\x1b[?25l", "foo", "\r\n", "\x1b[?25h"}
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

	// Small delay to allow initial render
	time.Sleep(time.Millisecond)

	// Simulate ctrl-c
	input.EmitKeypress("\x03", Key{Name: "c"})

	// Wait for result
	result := <-resultCh

	assert.True(t, IsCancel(result))
	assert.Equal(t, StateCancel, p.StateSnapshot())

	expectedOutput := []string{"\x1b[?25l", "foo", "\r\n", "\x1b[?25h"}
	assert.Equal(t, expectedOutput, output.Buffer)
}

func TestPrompt_CancelsOnEscape(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string { return "foo" },
	})

	// Start the prompt
	resultCh := make(chan any)
	go func() {
		resultCh <- p.Prompt()
	}()

	time.Sleep(time.Millisecond)

	// Simulate Escape key
	input.EmitKeypress("escape", Key{Name: "escape"})

	result := <-resultCh
	if !IsCancel(result) {
		t.Fatalf("expected cancel symbol, got %#v", result)
	}
	assert.Equal(t, StateCancel, p.StateSnapshot())
}

func TestPrompt_EmitsFinalizeOnSubmitAndCancel(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string { return "foo" },
	})

	var finalizeCount int32
	p.On("finalize", func() { atomic.AddInt32(&finalizeCount, 1) })

	// Submit path
	go p.Prompt()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)
	assert.True(t, atomic.LoadInt32(&finalizeCount) >= 1)

	// Cancel path
	p2 := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string { return "bar" },
	})
	atomic.StoreInt32(&finalizeCount, 0)
	p2.On("finalize", func() { atomic.AddInt32(&finalizeCount, 1) })
	go p2.Prompt()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	time.Sleep(time.Millisecond)
	assert.True(t, atomic.LoadInt32(&finalizeCount) >= 1)
}

func TestPrompt_InitialUserInputSetsValueAndEmitsEvent(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	got := ""
	p := NewPrompt(PromptOptions{
		Input:            input,
		Output:           output,
		InitialUserInput: "hello",
		Render:           func(p *Prompt) string { return "foo" },
	})

	p.On("userInput", func(v string) { got = v })

	go p.Prompt()
	time.Sleep(time.Millisecond)

	assert.Equal(t, "hello", p.UserInputSnapshot())
	assert.Equal(t, "hello", got)

	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_ReturnsCancelSymbolOnImmediateAbort(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string { return "foo" },
		Signal: ctx,
	})

	// Return cancel symbol without blocking
	result := p.Prompt()
	assert.True(t, IsCancel(result))
}

func TestPrompt_EmitsSubmitAndCancelEventsWithPayload(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string { return "foo" },
	})

	var submitted atomic.Value
	var cancelled atomic.Value
	p.On("submit", func(v any) { submitted.Store(v) })
	p.On("cancel", func(v any) { cancelled.Store(v) })

	// Submit path: preset value then press return
	go func() {
		_ = p.Prompt()
	}()
	time.Sleep(time.Millisecond)
	p.SetValue("answer")
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)
	assert.Equal(t, "answer", submitted.Load())

	// Cancel path on a new prompt
	p2 := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string { return "bar" },
	})
	submitted = atomic.Value{}
	cancelled = atomic.Value{}
	p2.On("submit", func(v any) { submitted.Store(v) })
	p2.On("cancel", func(v any) { cancelled.Store(v) })

	go func() { _ = p2.Prompt() }()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	time.Sleep(time.Millisecond)
	assert.True(t, IsCancel(cancelled.Load()))
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

	// We only assert that no value event fired
	assert.False(t, eventCalled)
}

func TestPrompt_ReRendersOnResize(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var renderCallCount atomic.Int32
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			renderCallCount.Add(1)
			return "foo"
		},
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)

	assert.Equal(t, int32(1), renderCallCount.Load())

	// Simulate resize event
	output.Emit("resize")
	time.Sleep(time.Millisecond)

	assert.Equal(t, int32(2), renderCallCount.Load())

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

	assert.Equal(t, StateInitial, p.StateSnapshot())

	go p.Prompt()
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateActive, p.StateSnapshot())

	// Cancel to exit
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
}

func TestPrompt_EmitsTruthyConfirmOnYPress(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var confirmValue atomic.Value
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	p.On("confirm", func(value bool) {
		confirmValue.Store(value)
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("y", Key{Name: "y"})
	// wait up to 20ms for event delivery
	for i := 0; i < 20; i++ {
		if _, ok := confirmValue.Load().(bool); ok {
			break
		}
		time.Sleep(time.Millisecond)
	}
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	v, _ := confirmValue.Load().(bool)
	assert.True(t, v)
}

func TestPrompt_EmitsFalseyConfirmOnNPress(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var confirmValue atomic.Value
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	p.On("confirm", func(value bool) {
		confirmValue.Store(value)
	})

	go p.Prompt()
	input.EmitKeypress("n", Key{Name: "n"})
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	v2, _ := confirmValue.Load().(bool)
	assert.False(t, v2)
}

func TestPrompt_EmitsKeyEventForUnknownChars(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	var keyChar atomic.Value
	var keyInfo atomic.Value
	var eventCount int32
	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
	})

	p.On("key", func(char string, key Key) {
		c := atomic.AddInt32(&eventCount, 1)
		if c == 1 {
			keyChar.Store(char)
			keyInfo.Store(key)
		}
	})

	go p.Prompt()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("z", Key{Name: "z"})
	for i := 0; i < 20; i++ {
		if keyChar.Load() != nil {
			break
		}
		time.Sleep(time.Millisecond)
	}
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	assert.Equal(t, "z", keyChar.Load())
	loadedKey, _ := keyInfo.Load().(Key)
	assert.Equal(t, "z", loadedKey.Name)
}

func TestPrompt_EmitsCursorEventsForMovementKeys(t *testing.T) {
	keys := []string{"up", "down", "left", "right"}

	for _, key := range keys {
		t.Run("key_"+key, func(t *testing.T) {
			input := NewMockReadable()
			output := NewMockWritable()

			var cursorEvent atomic.Value
			p := NewPrompt(PromptOptions{
				Input:  input,
				Output: output,
				Render: func(p *Prompt) string {
					return "foo"
				},
			})

			p.On("cursor", func(direction string) {
				cursorEvent.Store(direction)
			})

			go p.Prompt()
			time.Sleep(time.Millisecond)
			input.EmitKeypress(key, Key{Name: key})
			for i := 0; i < 20; i++ {
				if cursorEvent.Load() != nil {
					break
				}
				time.Sleep(time.Millisecond)
			}
			input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

			assert.Equal(t, key, cursorEvent.Load())
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

	p.SetValue("invalid")
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)

	// Check state before canceling
	assert.Equal(t, StateError, p.StateSnapshot())

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

	p.SetValue("valid")
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateSubmit, p.StateSnapshot())
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

			var cursorEvent atomic.Value
			p := NewPromptWithTracking(PromptOptions{
				Input:  input,
				Output: output,
				Render: func(p *Prompt) string {
					return "foo"
				},
			}, false)

			p.On("cursor", func(direction string) {
				cursorEvent.Store(direction)
			})

			go p.Prompt()
			time.Sleep(time.Millisecond)

			input.EmitKeypress(alias, Key{Name: alias})
			for i := 0; i < 20; i++ {
				if cursorEvent.Load() != nil {
					break
				}
				time.Sleep(time.Millisecond)
			}
			time.Sleep(time.Millisecond)

			assert.Equal(t, expected, cursorEvent.Load())

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

	assert.Equal(t, StateActive, p.StateSnapshot())

	cancel()
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateCancel, p.StateSnapshot())
}

func TestPrompt_ReturnsImmediatelyIfSignalIsAlreadyAborted(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewPrompt(PromptOptions{
		Input:  input,
		Output: output,
		Render: func(p *Prompt) string {
			return "foo"
		},
		Signal: ctx,
	})

	result := p.Prompt()
	assert.True(t, IsCancel(result))
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

	assert.Equal(t, StateActive, p.StateSnapshot())

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

	p.SetValue("invalid")
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateError, p.StateSnapshot())

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
			// Uppercase letters only
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

	p.SetValue("Invalid Value $$$")
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateError, p.StateSnapshot())

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
			// Uppercase letters only
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

	p.SetValue("VALID")
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	time.Sleep(time.Millisecond)

	assert.Equal(t, StateSubmit, p.StateSnapshot())
}
