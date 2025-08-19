package core

import (
	"context"
	"errors"
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
		p.Emit("confirm", strings.ToLower(char) == "y")
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
	for k, v := range p.preSubs {
		p.subscribers[k] = append(p.subscribers[k], v...)
	}
	p.snap.Store(st)
	for ev := range p.evCh {
		ev(&st)
		if p.opts.Render != nil {
			frame := p.opts.Render(p)
			if frame != st.PrevFrame {
				if st.State == StateInitial {
					_, _ = p.output.Write([]byte(CursorHide))
				}
				_, _ = p.output.Write([]byte(frame))
				if st.State == StateInitial {
					st.State = StateActive
				}
				st.PrevFrame = frame
			}
		}
		p.snap.Store(st)
		if st.State == StateSubmit || st.State == StateCancel {
			p.Emit("finalize")
			_, _ = p.output.Write([]byte("\n"))
			_, _ = p.output.Write([]byte(CursorShow))
			var res any
			if st.State == StateCancel {
				res = GetCancelSymbol()
				p.Emit("cancel", res)
			} else {
				res = st.Value
				p.Emit("submit", res)
			}
			p.doneCh <- res
			close(p.stopped)
			return
		}
	}
}
