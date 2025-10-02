package tap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntro_WritesBarStartAndTitle(t *testing.T) {
	out := NewMockWritable()

	Intro("Welcome", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	last := frames[len(frames)-1]

	assert.Contains(t, last, gray(BarStart))
	assert.Contains(t, last, bold("Welcome"))
	assert.Contains(t, last, gray(Bar)+"\n")
}

func TestCancel_WritesBarEndAndRedMessage(t *testing.T) {
	out := NewMockWritable()

	Cancel("Operation cancelled", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	last := frames[len(frames)-1]

	assert.Contains(t, last, gray(BarEnd))
	assert.Contains(t, last, red("Operation cancelled"))
	assert.Contains(t, last, gray(Bar)+"\n")
}

func TestOutro_WritesBarAndBarEndWithMessage(t *testing.T) {
	out := NewMockWritable()

	Outro("All done", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	last := frames[len(frames)-1]

	// Should include a gray bar line and a final line with message
	assert.Contains(t, last, gray(Bar))
	assert.Contains(t, last, gray(BarEnd))
	assert.Contains(t, last, bold("All done"))
}

func TestMessage_SurroundsContentWithBars(t *testing.T) {
	out := NewMockWritable()

	Message("Summary", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.Len(t, frames, 3)
	assert.Equal(t, gray(Bar)+"\n", frames[0])
	assert.Contains(t, frames[1], green(StepSubmit))
	assert.Contains(t, frames[1], bold("Summary"))
	assert.Equal(t, gray(Bar)+"\n", frames[2])
}
