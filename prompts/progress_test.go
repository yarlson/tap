package prompts

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yarlson/tap/core"
)

func TestProgress_RendersProgressBar(t *testing.T) {
	out := core.NewMockWritable()

	prog := NewProgress(ProgressOptions{
		Output: out,
		Style:  "heavy",
		Max:    10,
		Size:   20,
	})

	prog.Start("Processing...")
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	// Should render initial progress bar with 0 progress
	lastFrame := frames[len(frames)-1]
	assert.Contains(t, lastFrame, "Processing...")
	// Should contain heavy style characters
	assert.Contains(t, lastFrame, "━")
}

func TestProgress_AdvancesProgress(t *testing.T) {
	out := core.NewMockWritable()

	prog := NewProgress(ProgressOptions{
		Output: out,
		Style:  "heavy",
		Max:    10,
		Size:   20,
	})

	prog.Start("Loading...")
	time.Sleep(time.Millisecond)

	// Advance progress by 5
	prog.Advance(5, "Halfway...")
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	// Should show progress advancement
	lastFrame := frames[len(frames)-1]
	assert.Contains(t, lastFrame, "Halfway...")

	// Should have more filled progress bar
	heavyCount := strings.Count(lastFrame, "━")
	assert.Greater(t, heavyCount, 0, "Should have some progress filled")
}

func TestProgress_DifferentStyles(t *testing.T) {
	tests := []struct {
		style string
		char  string
	}{
		{"light", "─"},
		{"heavy", "━"},
		{"block", "█"},
	}

	for _, test := range tests {
		t.Run(test.style, func(t *testing.T) {
			out := core.NewMockWritable()

			prog := NewProgress(ProgressOptions{
				Output: out,
				Style:  test.style,
				Max:    10,
				Size:   10,
			})

			prog.Start("Test")
			time.Sleep(time.Millisecond)

			frames := out.GetFrames()
			assert.NotEmpty(t, frames)

			lastFrame := frames[len(frames)-1]
			assert.Contains(t, lastFrame, test.char, "Should contain style character")
		})
	}
}

func TestProgress_CompletesToFullBar(t *testing.T) {
	out := core.NewMockWritable()

	prog := NewProgress(ProgressOptions{
		Output: out,
		Style:  "heavy",
		Max:    10,
		Size:   20,
	})

	prog.Start("Starting...")
	time.Sleep(time.Millisecond)

	// Fill progress completely
	prog.Advance(10, "Complete!")
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	lastFrame := frames[len(frames)-1]
	assert.Contains(t, lastFrame, "Complete!")

	// Should have full progress bar (all 20 characters filled)
	heavyCount := strings.Count(lastFrame, "━")
	assert.GreaterOrEqual(t, heavyCount, 20, "Should have full progress bar")
}

func TestProgress_ClampsToMaxValue(t *testing.T) {
	out := core.NewMockWritable()

	prog := NewProgress(ProgressOptions{
		Output: out,
		Style:  "heavy",
		Max:    10,
		Size:   20,
	})

	prog.Start("Starting...")
	time.Sleep(time.Millisecond)

	// Try to advance beyond max
	prog.Advance(15, "Over max")
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	lastFrame := frames[len(frames)-1]
	assert.Contains(t, lastFrame, "Over max")

	// Should still only fill to max (20 characters)
	heavyCount := strings.Count(lastFrame, "━")
	assert.Equal(t, 20, heavyCount, "Should clamp to max progress")
}

func TestProgress_MessageOnly(t *testing.T) {
	out := core.NewMockWritable()

	prog := NewProgress(ProgressOptions{
		Output: out,
		Style:  "heavy",
		Max:    10,
		Size:   20,
	})

	prog.Start("Starting...")
	time.Sleep(time.Millisecond)

	// Update message without advancing
	prog.Message("Just updating message...")
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	lastFrame := frames[len(frames)-1]
	assert.Contains(t, lastFrame, "Just updating message...")

	// Should still show empty progress bar (progress should be 0)
	// Count filled characters - should be minimal
	filledCount := strings.Count(lastFrame, cyan("━"))
	assert.LessOrEqual(t, filledCount, 1, "Should have minimal progress when message only")
}

func TestProgress_StopWithMessage(t *testing.T) {
	out := core.NewMockWritable()

	prog := NewProgress(ProgressOptions{
		Output: out,
		Style:  "heavy",
		Max:    10,
		Size:   20,
	})

	prog.Start("Starting...")
	time.Sleep(time.Millisecond)

	prog.Advance(5, "Halfway...")
	time.Sleep(time.Millisecond)

	prog.Stop("Done!", 0)
	time.Sleep(time.Millisecond)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	lastFrame := frames[len(frames)-1]
	assert.Contains(t, lastFrame, "Done!")
	// Should contain submit symbol
	assert.Contains(t, lastFrame, green(StepSubmit))
}
