package tap

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/mattn/go-runewidth"
	xterm "golang.org/x/term"

	"github.com/yarlson/tap/internal/terminal"
)

type PromptOptions struct {
	Render           func(*Prompt) string
	InitialValue     any
	InitialUserInput string
	Validate         func(any) error
	Input            Reader
	Output           Writer
	Debug            bool
}

type EventHandler any

type Prompt struct {
	input  Reader
	output Writer
	opts   PromptOptions

	evCh    chan<- func(*promptState) // Write-only: for sending events (never blocks with unbounded queue)
	evOutCh <-chan func(*promptState) // Read-only: for receiving events in the loop
	doneCh  chan any
	stopped chan struct{}

	subscribers map[string][]EventHandler
	preSubs     map[string][]EventHandler

	snap        atomic.Value
	inEventLoop atomic.Bool // true when inside event loop processing

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
	s, _ := p.snap.Load().(promptState)
	return s.UserInput
}

func (p *Prompt) CursorSnapshot() int {
	s, _ := p.snap.Load().(promptState)
	return s.Cursor
}

func (p *Prompt) ErrorSnapshot() string {
	s, _ := p.snap.Load().(promptState)
	return s.Error
}

func (p *Prompt) ValueSnapshot() any {
	s, _ := p.snap.Load().(promptState)
	return s.Value
}

func (p *Prompt) snapshot() (state ClackState, prevFrame string) {
	s, _ := p.snap.Load().(promptState)
	return s.State, s.PrevFrame
}

// removed; handled inside loop

// SetValue schedules a value update (for tests or programmatic flows).
// When called from within the event loop, it updates immediately.
// When called from outside, it enqueues an event.
func (p *Prompt) SetValue(v any) {
	// If we're inside the event loop, modify state directly to avoid re-entrant channel send
	if p.inEventLoop.Load() {
		p.cur.Value = v
		p.Emit("value", v)

		return
	}

	// Outside the event loop, send an event
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

// NewPrompt creates a new prompt instance with default tracking.
func NewPrompt(options PromptOptions) *Prompt {
	return NewPromptWithTracking(options, true)
}

// unboundedQueue creates an unbounded event queue by using a goroutine with a slice buffer.
// Returns input and output channels. Input never blocks. Output delivers events in order.
func unboundedQueue() (input chan<- func(*promptState), output <-chan func(*promptState)) {
	in := make(chan func(*promptState))
	out := make(chan func(*promptState))

	go func() {
		var queue []func(*promptState)
		for {
			if len(queue) == 0 {
				// Queue is empty, just receive
				fn, ok := <-in
				if !ok {
					close(out)
					return
				}

				queue = append(queue, fn)
			} else {
				// Queue has items, try to send the first one or receive more
				select {
				case out <- queue[0]:
					queue = queue[1:]
				case fn, ok := <-in:
					if !ok {
						// Input closed, drain queue
						for _, f := range queue {
							out <- f
						}

						close(out)

						return
					}

					queue = append(queue, fn)
				}
			}
		}
	}()

	return in, out
}

// NewPromptWithTracking creates a new prompt instance with specified tracking.
func NewPromptWithTracking(options PromptOptions, trackValue bool) *Prompt {
	evIn, evOut := unboundedQueue()

	p := &Prompt{
		input:       options.Input,
		output:      options.Output,
		opts:        options,
		subscribers: make(map[string][]EventHandler),
		preSubs:     make(map[string][]EventHandler),
		track:       trackValue,
		evCh:        evIn,
		evOutCh:     evOut,
		doneCh:      make(chan any, 1),
		stopped:     make(chan struct{}),
	}
	// Default TTY will be provided by a higher-level adapter when needed
	p.snap.Store(promptState{State: StateInitial})

	return p
}

// On subscribes to an event.
func (p *Prompt) On(event string, handler any) {
	p.preSubs[event] = append(p.preSubs[event], handler)
}

// Emit emits an event to all subscribers.
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
				if v, ok := args[0].(bool); ok {
					h(v)
				}
			}
		case "key":
			if h, ok := handler.(func(string, Key)); ok && len(args) >= 2 {
				s, _ := args[0].(string)
				k, _ := args[1].(Key)
				h(s, k)
			}
		case "cursor":
			if h, ok := handler.(func(string)); ok {
				if v, ok := args[0].(string); ok {
					h(v)
				}
			}
		case "userInput":
			if h, ok := handler.(func(string)); ok {
				if v, ok := args[0].(string); ok {
					h(v)
				}
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

// Prompt starts the prompt and returns the result.
func (p *Prompt) Prompt(ctx context.Context) any {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}

	// Adopt pre-subscribers synchronously BEFORE starting the loop or registering
	// keypress handlers. This fixes a race condition where the first keypress could
	// be ignored if it arrived before adoptPreSubscribers() ran in the loop goroutine.
	p.adoptPreSubscribers()

	go p.loop()

	if p.input != nil {
		p.input.On("keypress", func(char string, key Key) {
			select {
			case p.evCh <- func(s *promptState) {
				p.handleKey(s, char, key)
			}:
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

	if ctx != nil {
		go func() {
			<-ctx.Done()

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

	if key.Name == "escape" || strings.EqualFold(char, "escape") {
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

	hasConfirmSubscribers := len(p.subscribers["confirm"]) > 0 || len(p.preSubs["confirm"]) > 0
	if char != "" && (strings.EqualFold(char, "y") || strings.EqualFold(char, "n")) && hasConfirmSubscribers {
		val := strings.EqualFold(char, "y")
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

// updateUserInputWithCursor handles cursor-based input tracking.
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
			return string(slices.Delete(runes, cursor-1, cursor)), cursor - 1
		}

		return current, cursor

	case "delete":
		// Delete character at cursor
		if cursor < len(runes) {
			return string(slices.Delete(runes, cursor, cursor+1)), cursor
		}

		return current, cursor

	case "up", "down", "escape":
		// These keys don't change input or cursor
		return current, cursor

	case "tab":
		// Insert tab at cursor
		return string(slices.Insert(runes, cursor, '\t')), cursor + 1

	case "space":
		// Insert space at cursor
		return string(slices.Insert(runes, cursor, ' ')), cursor + 1

	default:
		// Regular printable characters - insert at cursor position
		if char != "" {
			for _, r := range char {
				if r >= 32 && r <= 126 { // Printable ASCII
					runes = slices.Insert(runes, cursor, r)
					cursor++
				}
			}

			return string(runes), cursor
		}

		return current, cursor
	}
}

func (p *Prompt) loop() {
	st := promptState{State: StateInitial}

	// Note: adoptPreSubscribers() is now called synchronously in Prompt() before
	// the loop starts, to avoid a race condition with keypress events.
	p.snap.Store(st)

	for ev := range p.evOutCh {
		p.cur = &st
		p.inEventLoop.Store(true)

		ev(&st)

		p.inEventLoop.Store(false)
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

// Detect terminal width; fall back to 80.
func getColumns() int {
	fd := int(os.Stdout.Fd())
	if cols, _, err := xterm.GetSize(fd); err == nil && cols > 0 {
		return cols
	}

	return 80
}

// Printable width ignoring ANSI; rune-count approximation.
func visibleWidth(s string) int {
	clean := ansiRegexp.ReplaceAllString(s, "")
	return runewidth.StringWidth(clean)
}

// Rows occupied by frame, accounting for soft-wrapping.
func countPhysicalLines(s string) int {
	if s == "" {
		return 0
	}

	cols := getColumns()
	if cols <= 0 {
		cols = 80
	}

	total := 0
	// Split by hard newlines, then estimate wraps for each logical line
	segments := strings.Split(s, "\n")
	for _, line := range segments {
		w := visibleWidth(line)
		if w == 0 {
			total++
			continue
		}
		// Ceiling division for wrapped rows.
		rows := (w-1)/cols + 1
		total += rows
	}

	return total
}

// renderIfNeeded runs the render function, hides the cursor on the first frame,
// writes the frame, and updates state to active. It only writes when the frame
// content changes.
func (p *Prompt) renderIfNeeded(st *promptState) {
	if p.opts.Render == nil || p.output == nil {
		return
	}

	// Ensure render sees the current state by updating the snapshot first.
	// Without this, snapshot accessors would lag one event behind.
	p.snap.Store(*st)

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
	st.PrevFrameLines = countPhysicalLines(frame)
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
		var res any
		p.Emit("cancel", res)

		return res
	}

	res := st.Value
	p.Emit("submit", res)

	return res
}

// SetTermIO sets a custom reader and writer used by helpers. Pass nil values to
// restore default terminal behavior.
func SetTermIO(in Reader, out Writer) { ioReader, ioWriter = in, out }

// runWithTerminal creates a temporary terminal for interactive prompts and
// ensures cleanup after the prompt completes.
func runWithTerminal[T any](fn func(Reader, Writer) T) T {
	if ioReader != nil || ioWriter != nil {
		return fn(ioReader, ioWriter)
	}

	t, err := terminal.New()
	if err != nil {
		var zero T
		return zero
	}

	return fn(t.Reader, t.Writer)
}

// resolveWriter returns the output writer for simple output operations
// (like Intro, Outro, Message). It uses a lightweight stdout wrapper that
// doesn't start a readKeys goroutine, preventing zombie terminals from
// stealing keypresses from interactive prompts.
func resolveWriter() Writer {
	// Check if we have override I/O set
	if out := getOverrideWriter(); out != nil {
		return out
	}

	// Return a simple stdout writer - no full terminal needed for output-only operations.
	// This avoids the bug where resolveWriter would create a terminal that keeps reading
	// keys and interferes with interactive prompts.
	return &stdoutWriter{}
}

// stdoutWriter is a simple Writer implementation that writes to stdout.
// It doesn't start any goroutines or open a TTY, making it safe to use
// for output-only utilities like Intro, Outro, Message, etc.
type stdoutWriter struct{}

func (w *stdoutWriter) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (w *stdoutWriter) On(_ string, _ func()) {
	// No-op: output-only writer doesn't need resize handling
}

func (w *stdoutWriter) Emit(_ string) {
	// No-op: output-only writer doesn't emit events
}

// getOverrideWriter returns the override writer if set.
func getOverrideWriter() Writer {
	// Just return the override writer directly - do NOT create a terminal here!
	// Creating a terminal just to check for an override would leave a zombie
	// readKeys goroutine that steals keypresses from the real terminal.
	return ioWriter
}
