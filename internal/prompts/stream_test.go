package prompts

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yarlson/tap/internal/core"
)

func TestStream_StartWriteStop_Success(t *testing.T) {
	out := core.NewMockWritable()
	st := NewStream(StreamOptions{Output: out})

	st.Start("Building project")
	st.WriteLine("step 1: fetch deps")
	st.WriteLine("step 2: compile")
	st.Stop("Done", 0)

	frames := out.GetFrames()
	joined := strings.Join(frames, "\n")

	// Header initially shows active symbol and message
	assert.Contains(t, joined, Symbol(core.StateActive))
	assert.Contains(t, joined, "Building project")
	// Lines are prefixed
	assert.Contains(t, joined, cyan(Bar)+"  step 1: fetch deps")
	assert.Contains(t, joined, cyan(Bar)+"  step 2: compile")
	// Finalization repaints header/body and prints colored message without diamond
	assert.Contains(t, joined, green("Done"))
}

func TestStream_StopWithErrorAndCancel(t *testing.T) {
	out := core.NewMockWritable()
	st := NewStream(StreamOptions{Output: out})

	st.Start("Running tasks")
	st.WriteLine("doing things")
	st.Stop("Cancelled", 1)

	frames := out.GetFrames()
	joined := strings.Join(frames, "\n")
	// final line should not include a diamond symbol anymore; just the message inside the block
	assert.Contains(t, joined, "Cancelled")

	out2 := core.NewMockWritable()
	st2 := NewStream(StreamOptions{Output: out2})
	st2.Start("Running tasks")
	st2.WriteLine("doing things")
	st2.Stop("Failed", 2)

	frames2 := out2.GetFrames()
	joined2 := strings.Join(frames2, "\n")
	// no diamond symbol on error either
	assert.Contains(t, joined2, "Failed")
}

func TestStream_PipeReader(t *testing.T) {
	out := core.NewMockWritable()
	st := NewStream(StreamOptions{Output: out})

	st.Start("Streaming logs")
	data := bytes.NewBufferString("line 1\nline 2\nline 3\n")
	done := make(chan struct{})
	go func() {
		st.Pipe(data)
		st.Stop("OK", 0)
		close(done)
	}()
	// Allow goroutine to write
	time.Sleep(10 * time.Millisecond)
	<-done

	frames := out.GetFrames()
	joined := strings.Join(frames, "\n")
	assert.Contains(t, joined, cyan(Bar)+"  line 1")
	assert.Contains(t, joined, cyan(Bar)+"  line 2")
	assert.Contains(t, joined, cyan(Bar)+"  line 3")
}
