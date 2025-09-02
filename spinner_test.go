package tap

import (
	"strings"
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

func TestSpinner_OSC94Signals(t *testing.T) {
	out := NewMockWritable()
	s := NewSpinner(SpinnerOptions{Output: out})

	s.Start("working")
	time.Sleep(2 * time.Millisecond)
	// success
	s.Stop("ok", 0)

	frames := strings.Join(out.GetFrames(), "")
	// Start should emit indeterminate spinner: ESC ] 9 ; 4 ; 3 ST
	assert.Contains(t, frames, "\x1b]9;4;3\x1b\\")
	// Stop with success should clear: ESC ] 9 ; 4 ; 0 ST
	assert.Contains(t, frames, "\x1b]9;4;0\x1b\\")

	// Error (still clears)
	out2 := NewMockWritable()
	s2 := NewSpinner(SpinnerOptions{Output: out2})
	s2.Start("working")
	time.Sleep(2 * time.Millisecond)
	s2.Stop("boom", 2)
	frames2 := strings.Join(out2.GetFrames(), "")
	assert.Contains(t, frames2, "\x1b]9;4;0\x1b\\")

	// Cancel (still clears)
	out3 := NewMockWritable()
	s3 := NewSpinner(SpinnerOptions{Output: out3})
	s3.Start("working")
	time.Sleep(2 * time.Millisecond)
	s3.Stop("cancel", 1)
	frames3 := strings.Join(out3.GetFrames(), "")
	assert.Contains(t, frames3, "\x1b]9;4;0\x1b\\")
}
