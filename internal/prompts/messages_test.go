package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yarlson/tap/internal/core"
)

func TestIntro_WritesBarStartAndTitle(t *testing.T) {
	out := core.NewMockWritable()

	Intro("Welcome", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	last := frames[len(frames)-1]

	assert.Contains(t, last, gray(BarStart))
	assert.Contains(t, last, "Welcome")
}

func TestCancel_WritesBarEndAndRedMessage(t *testing.T) {
	out := core.NewMockWritable()

	Cancel("Operation cancelled", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	last := frames[len(frames)-1]

	assert.Contains(t, last, gray(BarEnd))
	assert.Contains(t, last, red("Operation cancelled"))
}

func TestOutro_WritesBarAndBarEndWithMessage(t *testing.T) {
	out := core.NewMockWritable()

	Outro("All done", MessageOptions{Output: out})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	last := frames[len(frames)-1]

	// Should include a gray bar line and a final line with message
	assert.Contains(t, last, gray(Bar))
	assert.Contains(t, last, gray(BarEnd))
	assert.Contains(t, last, "All done")
}
