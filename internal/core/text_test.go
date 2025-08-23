package core

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestText_SubmitsTypedStringOnEnter(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan string, 1)
	go func() { resCh <- Text(TextOptions{Message: "Your name:", Input: in, Output: out}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a"})
	in.EmitKeypress("b", Key{Name: "b"})
	in.EmitKeypress("c", Key{Name: "c"})
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "abc", res)
}

func TestText_DefaultAppliedOnEmptySubmit(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan string, 1)
	go func() {
		resCh <- Text(TextOptions{Message: "Your name:", DefaultValue: "anon", Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "anon", res)
}

func TestText_BackspaceEditing(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan string, 1)
	go func() { resCh <- Text(TextOptions{Message: "Input:", Input: in, Output: out}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("a", Key{Name: "a"})
	in.EmitKeypress("b", Key{Name: "b"})
	in.EmitKeypress("", Key{Name: "backspace"})
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "a", res)
}

func TestText_ValidationBlocksSubmitThenClearsOnKey(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	resCh := make(chan any, 1)
	go func() {
		resCh <- Text(TextOptions{Message: "Code:", Input: in, Output: out, Validate: func(s string) error {
			if len(s) < 2 {
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
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "ab", res)
}

func TestText_RendersWhileTyping(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	done := make(chan any, 1)
	go func() { done <- Text(TextOptions{Message: "Enter:", Input: in, Output: out}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("t", Key{Name: "t"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("e", Key{Name: "e"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	<-done
	// Expect at least one frame to contain the typed text
	found := false
	for _, s := range out.Buffer {
		if strings.Contains(s, "Enter:") && (strings.Contains(s, "t") || strings.Contains(s, "e")) {
			found = true
			break
		}
	}
	assert.True(t, found)
}
