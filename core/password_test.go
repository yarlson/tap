package core

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPassword_SubmitsMaskedInputOnEnter(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan any, 1)
	go func() { resCh <- Password(PasswordOptions{Message: "Password:", Input: in, Output: out}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("s", Key{Name: "s"})
	in.EmitKeypress("e", Key{Name: "e"})
	in.EmitKeypress("c", Key{Name: "c"})
	in.EmitKeypress("r", Key{Name: "r"})
	in.EmitKeypress("e", Key{Name: "e"})
	in.EmitKeypress("t", Key{Name: "t"})
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "secret", res)

	// Ensure that at least one frame masked the input (no plain 'secret')
	foundMasked := false
	for _, frame := range out.Buffer {
		if strings.Contains(frame, "Password:") && strings.Contains(frame, "●") {
			foundMasked = true
			break
		}
	}
	assert.True(t, foundMasked)
}

func TestPassword_DefaultAppliedOnEmptySubmit(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan any, 1)
	go func() {
		resCh <- Password(PasswordOptions{Message: "Enter:", DefaultValue: "fallback", Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "fallback", res)
}

func TestPassword_ValidationBlocksSubmitThenClearsOnKey(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan any, 1)
	go func() {
		resCh <- Password(PasswordOptions{Message: "Pass:", Input: in, Output: out, Validate: func(s string) error {
			if len(s) < 4 {
				return NewValidationError("too short")
			}
			return nil
		}})
	}()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	// still running (blocked by validation error)
	time.Sleep(time.Millisecond)
	in.EmitKeypress("b", Key{Name: "b"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("c", Key{Name: "c"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("d", Key{Name: "d"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "abcd", res)
}

// Intentionally no placeholder rendering for password prompts
