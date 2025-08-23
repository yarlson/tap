package tap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yarlson/tap/internal/core"
)

func withIO(in *core.MockReadable, out *core.MockWritable) func() {
	oldIn, oldOut := ioReader, ioWriter
	SetTermIO(in, out)
	return func() { SetTermIO(oldIn, oldOut) }
}

func TestTap_Text(t *testing.T) {
	in := core.NewMockReadable()
	out := core.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	done := make(chan struct{})
	go func() {
		_ = Text(TextOptions{Message: "Your name:"})
		close(done)
	}()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("A", core.Key{Name: "a"})
	in.EmitKeypress("l", core.Key{Name: "l"})
	in.EmitKeypress("i", core.Key{Name: "i"})
	in.EmitKeypress("c", core.Key{Name: "c"})
	in.EmitKeypress("e", core.Key{Name: "e"})
	in.EmitKeypress("", core.Key{Name: "return"})
	<-done

	joined := ""
	for _, f := range out.Buffer {
		joined += f
	}
	assert.Contains(t, joined, "Your name:")
}

func TestTap_Confirm(t *testing.T) {
	in := core.NewMockReadable()
	out := core.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	resultCh := make(chan bool, 1)
	go func() { resultCh <- Confirm(ConfirmOptions{Message: "Proceed?"}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("y", core.Key{Name: "y"})
	res := <-resultCh
	assert.True(t, res)
}

func TestTap_Select(t *testing.T) {
	in := core.NewMockReadable()
	out := core.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	opts := []SelectOption[string]{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue"},
	}

	resCh := make(chan string, 1)
	go func() { resCh <- Select[string](SelectOptions[string]{Message: "Color?", Options: opts}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", core.Key{Name: "down"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", core.Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "blue", res)
}

func TestTap_Password(t *testing.T) {
	in := core.NewMockReadable()
	out := core.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	resCh := make(chan string, 1)
	go func() { resCh <- Password(PasswordOptions{Message: "Password:"}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("s", core.Key{Name: "s"})
	in.EmitKeypress("e", core.Key{Name: "e"})
	in.EmitKeypress("c", core.Key{Name: "c"})
	in.EmitKeypress("r", core.Key{Name: "r"})
	in.EmitKeypress("e", core.Key{Name: "e"})
	in.EmitKeypress("t", core.Key{Name: "t"})
	in.EmitKeypress("", core.Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "secret", res)
}
