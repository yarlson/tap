package tap

import (
	"testing"
	"time"

	"github.com/yarlson/tap/internal/prompts"

	"github.com/stretchr/testify/assert"
)

func withIO(in *prompts.MockReadable, out *prompts.MockWritable) func() {
	SetTermIO(in, out)
	return func() { SetTermIO(nil, nil) }
}

func TestTap_Text(t *testing.T) {
	in := prompts.NewMockReadable()
	out := prompts.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	done := make(chan struct{})
	go func() {
		_ = Text(TextOptions{Message: "Your name:"})
		close(done)
	}()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("A", prompts.Key{Name: "a"})
	in.EmitKeypress("l", prompts.Key{Name: "l"})
	in.EmitKeypress("i", prompts.Key{Name: "i"})
	in.EmitKeypress("c", prompts.Key{Name: "c"})
	in.EmitKeypress("e", prompts.Key{Name: "e"})
	in.EmitKeypress("", prompts.Key{Name: "return"})
	<-done

	joined := ""
	for _, f := range out.Buffer {
		joined += f
	}
	assert.Contains(t, joined, "Your name:")
}

func TestTap_Confirm(t *testing.T) {
	in := prompts.NewMockReadable()
	out := prompts.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	resultCh := make(chan bool, 1)
	go func() { resultCh <- Confirm(ConfirmOptions{Message: "Proceed?"}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("y", prompts.Key{Name: "y"})
	res := <-resultCh
	assert.True(t, res)
}

func TestTap_Select(t *testing.T) {
	in := prompts.NewMockReadable()
	out := prompts.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	opts := []SelectOption[string]{
		{Value: "red", Label: "Red"},
		{Value: "blue", Label: "Blue"},
	}

	resCh := make(chan string, 1)
	go func() { resCh <- Select[string](SelectOptions[string]{Message: "Color?", Options: opts}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", prompts.Key{Name: "down"})
	time.Sleep(time.Millisecond)
	in.EmitKeypress("", prompts.Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "blue", res)
}

func TestTap_Password(t *testing.T) {
	in := prompts.NewMockReadable()
	out := prompts.NewMockWritable()
	cleanup := withIO(in, out)
	defer cleanup()

	resCh := make(chan string, 1)
	go func() { resCh <- Password(PasswordOptions{Message: "Password:"}) }()
	time.Sleep(time.Millisecond)
	in.EmitKeypress("s", prompts.Key{Name: "s"})
	in.EmitKeypress("e", prompts.Key{Name: "e"})
	in.EmitKeypress("c", prompts.Key{Name: "c"})
	in.EmitKeypress("r", prompts.Key{Name: "r"})
	in.EmitKeypress("e", prompts.Key{Name: "e"})
	in.EmitKeypress("t", prompts.Key{Name: "t"})
	in.EmitKeypress("", prompts.Key{Name: "return"})
	res := <-resCh
	assert.Equal(t, "secret", res)
}
