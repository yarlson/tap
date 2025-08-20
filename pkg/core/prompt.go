package core

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync/atomic"
)

type PromptOptions struct {
	Render           func(*Prompt) string
	InitialValue     any
	InitialUserInput string
	Validate         func(any) error
	Input            Reader
	Output           Writer
	Debug            bool
	Signal           context.Context
}

type EventHandler any

type Prompt struct {
	input  Reader
	output Writer
	opts   PromptOptions

	evCh    chan func(*promptState)
	doneCh  chan any
	stopped chan struct{}

	subscribers map[string][]EventHandler
	preSubs     map[string][]EventHandler

	snap atomic.Value

	track bool

	cleanup func()
}

type promptState struct {
	State     ClackState
	Error     string
	Value     any
	UserInput string
	PrevFrame string
}

func (p *Prompt) StateSnapshot() ClackState {
	st, _ := p.snapshot()
	return st
}

func (p *Prompt) UserInputSnapshot() string {
	s := p.snap.Load().(promptState)
	return s.UserInput
}

func (p *Prompt) snapshot() (ClackState, string) {
	s := p.snap.Load().(promptState)
	return s.State, s.PrevFrame
}

// removed; handled inside loop

// SetValue schedules a value update (for tests or programmatic flows).
// In the event-loop refactor, this will post to the loop; for now, set under lock.
func (p *Prompt) SetValue(v any) {
	select {
	case p.evCh <- func(s *promptState) { s.Value = v; p.Emit("value", v) }:
	case <-p.stopped:
	}
}

// NewPrompt creates a new prompt instance with default tracking
func NewPrompt(options PromptOptions) *Prompt {
	return NewPromptWithTracking(options, true)
}

// NewPromptWithTracking creates a new prompt instance with specified tracking
func NewPromptWithTracking(options PromptOptions, trackValue bool) *Prompt {
	p := &Prompt{
		input:       options.Input,
		output:      options.Output,
		opts:        options,
		subscribers: make(map[string][]EventHandler),
		preSubs:     make(map[string][]EventHandler),
		track:       trackValue,
		evCh:        make(chan func(*promptState), 64),
		doneCh:      make(chan any, 1),
		stopped:     make(chan struct{}),
	}
	// Provide default TTY input/output when not supplied
	if p.input == nil || p.output == nil {
		// Best-effort defaults using current process stdio
		if p.input == nil {
			if ti, restore, err := newDefaultTTYInput(os.Stdin); err == nil {
				p.input = ti
				p.cleanup = restore
			}
		}
		if p.output == nil {
			p.output = newDefaultTTYOutput(os.Stdout)
		}
	}
	p.snap.Store(promptState{State: StateInitial})
	return p
}

// On subscribes to an event
func (p *Prompt) On(event string, handler any) {
	p.preSubs[event] = append(p.preSubs[event], handler)
}

// Emit emits an event to all subscribers
func (p *Prompt) Emit(event string, args ...any) {
	handlers := p.subscribers[event]
	for _, handler := range handlers {
		switch event {
		case "value":
			if h, ok := handler.(func(any)); ok {
				h(args[0])
			}
		case "confirm":
			if h, ok := handler.(func(bool)); ok {
				h(args[0].(bool))
			}
		case "key":
			if h, ok := handler.(func(string, Key)); ok {
				h(args[0].(string), args[1].(Key))
			}
		case "cursor":
			if h, ok := handler.(func(string)); ok {
				h(args[0].(string))
			}
		case "userInput":
			if h, ok := handler.(func(string)); ok {
				h(args[0].(string))
			}
		case "finalize":
			if h, ok := handler.(func()); ok {
				h()
			}
		case "submit":
			if h, ok := handler.(func(any)); ok {
				h(args[0])
			}
		case "cancel":
			if h, ok := handler.(func(any)); ok {
				h(args[0])
			}
		}
	}
}

// Prompt starts the prompt and returns the result
func (p *Prompt) Prompt() any {
	if p.opts.Signal != nil {
		select {
		case <-p.opts.Signal.Done():
			return GetCancelSymbol()
		default:
		}
	}

	go p.loop()

	if p.input != nil {
		p.input.On("keypress", func(char string, key Key) {
			select {
			case p.evCh <- func(s *promptState) { p.handleKey(s, char, key) }:
			case <-p.stopped:
			}
		})
	}
	if p.output != nil {
		p.output.On("resize", func() {
			select {
			case p.evCh <- func(s *promptState) { p.handleResize(s) }:
			case <-p.stopped:
			}
		})
	}
	if p.opts.Signal != nil {
		go func() {
			<-p.opts.Signal.Done()
			select {
			case p.evCh <- func(s *promptState) { p.handleAbort(s) }:
			case <-p.stopped:
			}
		}()
	}

	if p.opts.InitialUserInput != "" {
		p.evCh <- func(s *promptState) {
			s.UserInput = p.opts.InitialUserInput
			p.Emit("userInput", s.UserInput)
		}
	}
	p.evCh <- func(s *promptState) { p.handleInitialRender(s) }

	return <-p.doneCh
}

func isMovementKey(keyName string) bool {
	return keyName == "up" || keyName == "down" || keyName == "left" || keyName == "right"
}

func isCancel(char string, key Key) bool {
	if char == "\x03" || (key.Ctrl && key.Name == "c") {
		return true
	}
	if key.Name == "escape" || strings.ToLower(char) == "escape" {
		return true
	}
	return false
}

func getMovementAlias(keyName string) string {
	aliases := map[string]string{
		"k": "up",
		"j": "down",
		"h": "left",
		"l": "right",
	}
	return aliases[keyName]
}

func (p *Prompt) handleInitialRender(s *promptState) {}

func (p *Prompt) handleResize(s *promptState) {}

func (p *Prompt) handleAbort(s *promptState) { s.State = StateCancel }

func (p *Prompt) handleKey(s *promptState, char string, key Key) {
	if isMovementKey(key.Name) {
		p.Emit("cursor", key.Name)
	}
	if alias := getMovementAlias(key.Name); !p.track && alias != "" {
		p.Emit("cursor", alias)
	}
	if char != "" && (strings.ToLower(char) == "y" || strings.ToLower(char) == "n") {
		val := strings.ToLower(char) == "y"
		p.Emit("confirm", val)
		s.Value = val
		s.State = StateSubmit
	}
	p.Emit("key", strings.ToLower(char), key)
	if key.Name == "return" {
		if p.opts.Validate != nil {
			if err := p.opts.Validate(s.Value); err != nil {
				var ve *ValidationError
				if errors.As(err, &ve) {
					s.Error = ve.Message
				} else {
					s.Error = err.Error()
				}
				s.State = StateError
			}
		}
		if s.State != StateError {
			s.State = StateSubmit
		}
	}
	if isCancel(char, key) {
		s.State = StateCancel
	}
}

func (p *Prompt) loop() {
	st := promptState{State: StateInitial}
	p.adoptPreSubscribers()
	p.snap.Store(st)

	for ev := range p.evCh {
		ev(&st)
		p.renderIfNeeded(&st)
		p.snap.Store(st)

		if p.shouldFinalize(st.State) {
			res := p.finalize(&st)
			p.doneCh <- res
			close(p.stopped)

			return
		}
	}
}

// adoptPreSubscribers moves temporary handlers registered before the loop started
// into the active subscribers map.
func (p *Prompt) adoptPreSubscribers() {
	for k, v := range p.preSubs {
		p.subscribers[k] = append(p.subscribers[k], v...)
	}
}

// renderIfNeeded runs the render function, hides the cursor on the first frame,
// writes the frame, and updates state to active. It only writes when the frame
// content changes.
func (p *Prompt) renderIfNeeded(st *promptState) {
	if p.opts.Render == nil || p.output == nil {
		return
	}

	frame := p.opts.Render(p)
	if frame == st.PrevFrame {
		return
	}

	if st.State == StateInitial {
		_, _ = p.output.Write([]byte(CursorHide))
	}

	_, _ = p.output.Write([]byte(frame))
	if st.State == StateInitial {
		st.State = StateActive
	}

	st.PrevFrame = frame
}

func (p *Prompt) shouldFinalize(state ClackState) bool {
	return state == StateSubmit || state == StateCancel
}

// finalize performs teardown, emits finalize/submit/cancel, and returns the
// result to send to the caller.
func (p *Prompt) finalize(st *promptState) any {
	p.Emit("finalize")
	// Write trailing newline and show the cursor again
	if p.output != nil {
		_, _ = p.output.Write([]byte("\r\n"))
		_, _ = p.output.Write([]byte(CursorShow))
	}

	if p.cleanup != nil {
		p.cleanup()
	}

	if st.State == StateCancel {
		res := GetCancelSymbol()
		p.Emit("cancel", res)

		return res
	}

	res := st.Value
	p.Emit("submit", res)

	return res
}
