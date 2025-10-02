package tap

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStyledSelect_RendersWithSymbolAndBars(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue"},
	}

	go func() {
		Select(context.Background(), SelectOptions[string]{
			Message: "Pick color:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()

	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.True(t, len(frames) > 0, "Should have rendered frames")

	// Should contain the title with message
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "Pick color:") {
			found = true
			break
		}
	}

	assert.True(t, found, "Should render message")
}

func TestStyledSelect_ShowsActiveInactiveOptions(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "first", Label: "First Option"},
		{Value: "second", Label: "Second Option"},
	}

	go func() {
		Select(context.Background(), SelectOptions[string]{
			Message: "Choose:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()

	time.Sleep(time.Millisecond)

	frames := out.GetFrames()

	// Should show active (●) and inactive (○) radio buttons
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, RadioActive) && strings.Contains(frame, RadioInactive) {
			found = true
			break
		}
	}

	assert.True(t, found, "Should show both active and inactive radio buttons")
}

func TestStyledSelect_ShowsHints(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "option1", Label: "Option 1", Hint: "This is a hint"},
		{Value: "option2", Label: "Option 2"},
	}

	go func() {
		Select(context.Background(), SelectOptions[string]{
			Message: "Pick:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()

	time.Sleep(time.Millisecond)

	frames := out.GetFrames()

	// Should show hint for active option
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "This is a hint") {
			found = true
			break
		}
	}

	assert.True(t, found, "Should show hint for active option")
}

func TestStyledSelect_ShowsSubmitState(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "selected", Label: "Selected Item"},
	}

	resCh := make(chan string, 1)

	go func() {
		resCh <- Select(context.Background(), SelectOptions[string]{
			Message: "Choose:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()

	time.Sleep(time.Millisecond)

	// Submit the selection
	in.EmitKeypress("", Key{Name: "return"})
	<-resCh

	frames := out.GetFrames()

	// Should show submit state with dimmed selected option
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "Selected Item") {
			found = true
			break
		}
	}

	assert.True(t, found, "Should show submitted option")
}

func TestStyledSelect_ShowsCancelState(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "option", Label: "Test Option"},
	}

	resCh := make(chan any, 1)

	go func() {
		resCh <- Select(context.Background(), SelectOptions[string]{
			Message: "Choose:",
			Options: options,
			Input:   in,
			Output:  out,
		})
	}()

	time.Sleep(time.Millisecond)

	// Cancel the selection
	in.EmitKeypress("\x03", Key{Name: "c", Ctrl: true})

	res := <-resCh
	// typed API returns zero value on cancel for string
	assert.Equal(t, "", res, "Should return zero value on cancel")

	frames := out.GetFrames()

	// Should show cancel state
	found := false

	for _, frame := range frames {
		if strings.Contains(frame, "Test Option") {
			found = true
			break
		}
	}

	assert.True(t, found, "Should show cancelled option")
}

func TestStyledSelect_InitialValuePositioning(t *testing.T) {
	in := NewMockReadable()
	out := NewMockWritable()
	options := []SelectOption[string]{
		{Value: "first", Label: "First"},
		{Value: "second", Label: "Second"},
		{Value: "third", Label: "Third"},
	}
	initialValue := "second"

	resCh := make(chan string, 1)

	go func() {
		resCh <- Select(context.Background(), SelectOptions[string]{
			Message:      "Choose:",
			Options:      options,
			InitialValue: &initialValue,
			Input:        in,
			Output:       out,
		})
	}()

	time.Sleep(time.Millisecond)

	// Submit immediately to test initial positioning
	in.EmitKeypress("", Key{Name: "return"})

	res := <-resCh

	assert.Equal(t, "second", res, "Should select initial value")
}
