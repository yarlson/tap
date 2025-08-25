package prompts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpinner_API(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out})

	// Ensure methods exist and basic start/stop do not panic
	s.Start("")
	time.Sleep(time.Millisecond)
	s.Message("hello")
	s.Stop("", 0)
}

func TestSpinner_RendersFrames(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out})

	s.Start("")
	time.Sleep(2 * time.Millisecond)
	s.Stop("", 0)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
}

func TestSpinner_RendersMessage(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out})

	s.Start("foo")
	time.Sleep(time.Millisecond)
	s.Stop("", 0)

	frames := out.GetFrames()
	last := frames[len(frames)-1]
	assert.Contains(t, last, "foo")
}

func TestSpinner_TimerIndicator(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out, Indicator: "timer"})

	s.Start("")
	time.Sleep(time.Millisecond)
	s.Stop("", 0)

	frames := out.GetFrames()
	last := frames[len(frames)-1]
	assert.Contains(t, last, "[")
}

func TestSpinner_CustomFramesAndDelay(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out, Frames: []string{"🐴", "🦋", "🐙", "🐶"}, Delay: 200 * time.Millisecond})

	s.Start("")
	time.Sleep(210 * time.Millisecond)
	s.Stop("", 0)

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
}

func TestSpinner_MessageUpdate(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out})

	s.Start("")
	time.Sleep(time.Millisecond)
	s.Message("foo")
	time.Sleep(time.Millisecond)
	s.Stop("", 0)

	frames := out.GetFrames()
	last := frames[len(frames)-1]
	assert.Contains(t, last, "foo")
}

func TestSpinner_StopCodes(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out})

	s.Start("")
	time.Sleep(time.Millisecond)
	s.Stop("", 1)

	frames := out.GetFrames()
	last := frames[len(frames)-1]
	assert.Contains(t, last, red(StepCancel))
}
