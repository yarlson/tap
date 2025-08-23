package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSelect_SubmitsSelectedOptionOnEnter(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue"},
		{Value: "green", Label: "Green"},
	}
	resCh := make(chan string, 1)
	go func() {
		resCh <- Select(SelectOptions[string]{Message: "Pick color:", Options: options, Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	// Should start at index 0 (red), submit it
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "red", res)
}

func TestSelect_NavigateWithArrowKeys(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "a", Label: "Option A"},
		{Value: "b", Label: "Option B"},
		{Value: "c", Label: "Option C"},
	}
	resCh := make(chan string, 1)
	go func() {
		resCh <- Select(SelectOptions[string]{Message: "Pick:", Options: options, Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	// Move down twice (0 -> 1 -> 2)
	in.EmitKeypress("", Key{Name: "down"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "down"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "c", res)
}

func TestSelect_WrapAroundNavigation(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "first", Label: "First"},
		{Value: "last", Label: "Last"},
	}
	resCh := make(chan string, 1)
	go func() {
		resCh <- Select(SelectOptions[string]{Message: "Pick:", Options: options, Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	// Move up from first option should wrap to last
	in.EmitKeypress("", Key{Name: "up"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "last", res)
}

func TestSelect_InitialValueSetsCorrectCursor(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue"},
		{Value: "green", Label: "Green"},
	}
	initialValue := "blue"
	resCh := make(chan string, 1)
	go func() {
		resCh <- Select(SelectOptions[string]{
			Message:      "Pick color:",
			Options:      options,
			InitialValue: &initialValue,
			Input:        in,
			Output:       out,
		})
	}()
	time.Sleep(time.Millisecond)
	// Should start at blue (index 1), submit it
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "blue", res)
}

func TestSelect_CancelWithCtrlC(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "option1", Label: "Option 1"},
	}
	resCh := make(chan string, 1)
	go func() {
		resCh <- Select(SelectOptions[string]{Message: "Pick:", Options: options, Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	res := <-resCh
	// typed API returns zero value on cancel; for string that's ""
	assert.Equal(t, "", res)
}

func TestSelect_LeftRightKeysAlsoNavigateUpDown(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "first", Label: "First"},
		{Value: "second", Label: "Second"},
		{Value: "third", Label: "Third"},
	}
	resCh := make(chan string, 1)
	go func() {
		resCh <- Select(SelectOptions[string]{Message: "Pick:", Options: options, Input: in, Output: out})
	}()
	time.Sleep(time.Millisecond)
	// Use right arrow to navigate down
	in.EmitKeypress("", Key{Name: "right"})
	time.Sleep(time.Millisecond)
	// Use left arrow to navigate up
	in.EmitKeypress("", Key{Name: "left"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "first", res) // Should be back at first option
}
