package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfirm_SubmitsTrueOnY(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	resultCh := make(chan any, 1)
	go func() { resultCh <- Confirm(ConfirmOptions{Input: input, Output: output, Message: "Are you sure?"}) }()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("y", Key{Name: "y"})
	result := <-resultCh
	b, ok := result.(bool)
	assert.True(t, ok)
	assert.True(t, b)
}

func TestConfirm_SubmitsFalseOnN(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	resultCh := make(chan any, 1)
	go func() { resultCh <- Confirm(ConfirmOptions{Input: input, Output: output, Message: "Are you sure?"}) }()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("n", Key{Name: "n"})
	result := <-resultCh
	b, ok := result.(bool)
	assert.True(t, ok)
	assert.False(t, b)
}

func TestConfirm_ToggleWithArrowsAndEnter(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	resultCh := make(chan any, 1)
	go func() {
		resultCh <- Confirm(ConfirmOptions{Input: input, Output: output, Message: "Proceed?", InitialValue: true})
	}()
	time.Sleep(time.Millisecond)
	// Toggle selection once (true -> false)
	input.EmitKeypress("", Key{Name: "right"})
	time.Sleep(time.Millisecond)
	input.EmitKeypress("", Key{Name: "return"})
	result := <-resultCh
	b, ok := result.(bool)
	assert.True(t, ok)
	assert.False(t, b)
}

func TestConfirm_CancelOnCtrlC(t *testing.T) {
	input := NewMockReadable()
	output := NewMockWritable()

	resultCh := make(chan any, 1)
	go func() { resultCh <- Confirm(ConfirmOptions{Input: input, Output: output, Message: "Cancel?"}) }()
	time.Sleep(time.Millisecond)
	input.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	result := <-resultCh
	assert.True(t, IsCancel(result))
}
