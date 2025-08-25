package tap

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MultiSelect should behave similarly to Select but allow toggling multiple
// items with space and submit a slice of values.

func TestStyledMultiSelect_RendersTitleAndOptions(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	options := []SelectOption[string]{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue"},
	}

	go func() {
		_ = MultiSelect[string](MultiSelectOptions[string]{
			Message: "Pick colors:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.Greater(t, len(frames), 0)

	foundTitle := false
	foundMarkers := false
	for _, f := range frames {
		if strings.Contains(f, "Pick colors:") {
			foundTitle = true
		}
		// Initial frame should show at least unchecked checkboxes
		if strings.Contains(f, CheckboxUnchecked) {
			foundMarkers = true
		}
		if foundTitle && foundMarkers {
			break
		}
	}
	assert.True(t, foundTitle, "should render the message title")
	assert.True(t, foundMarkers, "should render active and inactive markers")
}

func TestStyledMultiSelect_ToggleAndSubmit(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	options := []SelectOption[string]{
		{Value: "a", Label: "Option A"},
		{Value: "b", Label: "Option B"},
		{Value: "c", Label: "Option C"},
	}

	resCh := make(chan []string, 1)
	go func() {
		resCh <- MultiSelect[string](MultiSelectOptions[string]{
			Message: "Choose many:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()
	time.Sleep(time.Millisecond)

	// Cursor at 0 -> toggle A
	in.EmitKeypress("", Key{Name: "space"})
	time.Sleep(time.Millisecond)
	// Move down -> 1, toggle B
	in.EmitKeypress("", Key{Name: "down"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", Key{Name: "space"})
	time.Sleep(time.Millisecond)
	// Submit
	in.EmitKeypress("", Key{Name: "return"})

	res := <-resCh
	assert.ElementsMatch(t, []string{"a", "b"}, res)
}

func TestStyledMultiSelect_InitialValuesPreselected(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	options := []SelectOption[string]{
		{Value: "one", Label: "One"},
		{Value: "two", Label: "Two"},
		{Value: "three", Label: "Three"},
	}

	initial := []string{"two", "three"}

	resCh := make(chan []string, 1)
	go func() {
		resCh <- MultiSelect[string](MultiSelectOptions[string]{
			Message:       "Pick:",
			Options:       options,
			InitialValues: initial,
			Input:         in,
			Output:        out,
		})
	}()
	time.Sleep(time.Millisecond)

	// Submit immediately; should keep initial selections
	in.EmitKeypress("", Key{Name: "return"})

	res := <-resCh
	assert.ElementsMatch(t, []string{"two", "three"}, res)
}

func TestStyledMultiSelect_CancelWithCtrlC(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()

	options := []SelectOption[string]{
		{Value: "x", Label: "X"},
	}

	resCh := make(chan []string, 1)
	go func() {
		resCh <- MultiSelect[string](MultiSelectOptions[string]{
			Message: "Pick:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()
	time.Sleep(time.Millisecond)

	in.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})
	res := <-resCh
	// On cancel, typed API should return the zero value for []string which is nil
	assert.Nil(t, res)
}
