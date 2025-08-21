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

	cleanup func()
	cur     *promptState
}

type promptState struct {
	State          ClackState
	Error          string
	Value          any
	UserInput      string
	Cursor         int
	PrevFrame      string
	PrevFrameLines int
}

func (p *Prompt) StateSnapshot() ClackState {
	st, _ := p.snapshot()
	return st
}

func (p *Prompt) UserInputSnapshot() string {
	s := p.snap.Load().(promptState)
	return s.UserInput
}

func (p *Prompt) CursorSnapshot() int {
	s := p.snap.Load().(promptState)
	return s.Cursor
}

func (p *Prompt) ErrorSnapshot() string {
	s := p.snap.Load().(promptState)
	return s.Error
}

func (p *Prompt) ValueSnapshot() any {
	s := p.snap.Load().(promptState)
	return s.Value
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

// SetImmediateValue updates the value in the current event-loop tick if possible.
// Falls back to enqueuing when called outside the loop.
func (p *Prompt) SetImmediateValue(v any) {
	p.SetValue(v)
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
	// Default TTY will be provided by a higher-level adapter when needed
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
			s.Cursor = len([]rune(s.UserInput))
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

func (p *Prompt) handleInitialRender(_ *promptState) {}

func (p *Prompt) handleResize(_ *promptState) {}

func (p *Prompt) handleAbort(s *promptState) { s.State = StateCancel }

func (p *Prompt) handleKey(s *promptState, char string, key Key) {
	// Clear error on any keypress other than return/cancel (do this first)
	if s.State == StateError && key.Name != "return" && !isCancel(char, key) {
		s.State = StateActive
		s.Error = ""
	}

	// Track user input when tracking is enabled
	if p.track && key.Name != "return" {
		oldInput := s.UserInput
		oldCursor := s.Cursor
		newInput, newCursor := p.updateUserInputWithCursor(s.UserInput, s.Cursor, char, key)

		inputChanged := newInput != oldInput
		cursorChanged := newCursor != oldCursor

		if inputChanged {
			s.UserInput = newInput
			p.Emit("userInput", s.UserInput)
		}
		if cursorChanged {
			s.Cursor = newCursor
		}

		// Force re-render when state changes
		if inputChanged || cursorChanged {
			s.PrevFrame = ""
		}
	}

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
		// For text input tracking, set value from user input if no value is set
		if p.track && s.Value == nil {
			if s.UserInput != "" {
				s.Value = s.UserInput
			} else if p.opts.InitialValue != nil {
				s.Value = p.opts.InitialValue
			}
		}

		if p.opts.Validate != nil {
			if err := p.opts.Validate(s.Value); err != nil {
				var ve *ValidationError
				if errors.As(err, &ve) {
					s.Error = ve.Message
				} else {
					s.Error = err.Error()
				}
				s.State = StateError
			} else {
				s.Error = ""
				s.State = StateSubmit
			}
		} else {
			s.State = StateSubmit
		}
	}
	if isCancel(char, key) {
		s.State = StateCancel
	}
}

// updateUserInputWithCursor handles cursor-based input tracking
func (p *Prompt) updateUserInputWithCursor(current string, cursor int, char string, key Key) (newInput string, newCursor int) {
	runes := []rune(current)

	// Ensure cursor is within bounds
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}

	switch key.Name {
	case "left":
		// Move cursor left
		if cursor > 0 {
			return current, cursor - 1
		}
		return current, cursor

	case "right":
		// Move cursor right
		if cursor < len(runes) {
			return current, cursor + 1
		}
		return current, cursor

	case "backspace":
		// Delete character before cursor
		if cursor > 0 && len(runes) > 0 {
			newRunes := append(runes[:cursor-1], runes[cursor:]...)
			return string(newRunes), cursor - 1
		}
		return current, cursor

	case "delete":
		// Delete character at cursor
		if cursor < len(runes) {
			newRunes := append(runes[:cursor], runes[cursor+1:]...)
			return string(newRunes), cursor
		}
		return current, cursor

	case "up", "down", "escape":
		// These keys don't change input or cursor
		return current, cursor

	case "tab":
		// Insert tab at cursor
		newRunes := append(runes[:cursor], append([]rune{'\t'}, runes[cursor:]...)...)
		return string(newRunes), cursor + 1

	case "space":
		// Insert space at cursor
		newRunes := append(runes[:cursor], append([]rune{' '}, runes[cursor:]...)...)
		return string(newRunes), cursor + 1

	default:
		// Regular printable characters - insert at cursor position
		if char != "" && len(char) > 0 {
			for _, r := range char {
				if r >= 32 && r <= 126 { // Printable ASCII
					newRunes := append(runes[:cursor], append([]rune{r}, runes[cursor:]...)...)
					return string(newRunes), cursor + 1
				}
			}
		}
		return current, cursor
	}
}

func (p *Prompt) loop() {
	st := promptState{State: StateInitial}
	p.adoptPreSubscribers()
	p.snap.Store(st)

	for ev := range p.evCh {
		p.cur = &st
		ev(&st)
		p.renderIfNeeded(&st)
		p.snap.Store(st)

		if p.shouldFinalize(st.State) {
			p.renderIfNeeded(&st)
			p.snap.Store(st)
			res := p.finalize(&st)
			p.doneCh <- res
			close(p.stopped)
			p.cur = nil
			return
		}
		p.cur = nil
	}
}

// adoptPreSubscribers moves temporary handlers registered before the loop started
// into the active subscribers map.
func (p *Prompt) adoptPreSubscribers() {
	for k, v := range p.preSubs {
		p.subscribers[k] = append(p.subscribers[k], v...)
	}
}

// countLines counts the number of lines in a string
func countLines(s string) int {
	if s == "" {
		return 0
	}
	lines := 1
	for _, char := range s {
		if char == '\n' {
			lines++
		}
	}
	return lines
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
	} else {
		// Clear previous frame
		if st.PrevFrameLines > 1 {
			// Multi-line frame: move cursor up and clear from current position down
			for i := 0; i < st.PrevFrameLines-1; i++ {
				_, _ = p.output.Write([]byte(CursorUp))
			}
			_, _ = p.output.Write([]byte("\r"))
			_, _ = p.output.Write([]byte(EraseDown))
		} else {
			// Single-line frame: use fast single-line clear
			_, _ = p.output.Write([]byte("\r"))
			_, _ = p.output.Write([]byte(EraseLine))
		}
	}

	_, _ = p.output.Write([]byte(frame))
	if st.State == StateInitial {
		st.State = StateActive
	}

	st.PrevFrame = frame
	st.PrevFrameLines = countLines(frame)
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
