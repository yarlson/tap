package core

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// PromptOptions contains configuration for a prompt
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

// EventHandler represents an event handler function
type EventHandler any

// Prompt represents a CLI prompt
type Prompt struct {
	input  Reader
	output Writer
	opts   PromptOptions

	// Event system
	subscribers map[string][]EventHandler
	mutex       sync.RWMutex

	// State
	State     ClackState
	Error     string
	Value     any
	UserInput string

	// Internal state
	prevFrame string
	cursor    int
	track     bool
	done      chan any
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
		State:       StateInitial,
		Error:       "",
		subscribers: make(map[string][]EventHandler),
		track:       trackValue,
		done:        make(chan any, 1),
	}

	// Set up input event handling
	if p.input != nil {
		p.input.On("keypress", p.onKeypress)
	}

	// Set up output resize handling
	if p.output != nil {
		p.output.On("resize", func() {
			p.render()
		})
	}

	return p
}

// On subscribes to an event
func (p *Prompt) On(event string, handler any) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.subscribers[event] = append(p.subscribers[event], handler)
}

// Emit emits an event to all subscribers
func (p *Prompt) Emit(event string, args ...any) {
	p.mutex.RLock()
	handlers := p.subscribers[event]
	p.mutex.RUnlock()

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
	// Check for abort signal
	if p.opts.Signal != nil {
		select {
		case <-p.opts.Signal.Done():
			p.State = StateCancel
			p.close()
			return GetCancelSymbol()
		default:
		}

		// Watch for cancellation
		go func() {
			<-p.opts.Signal.Done()
			p.State = StateCancel
			p.close()
		}()
	}

	// Set initial user input if provided
	if p.opts.InitialUserInput != "" {
		p.setUserInput(p.opts.InitialUserInput, true)
	}

	// Force initial render to hide cursor and set state to active
	if p.State == StateInitial {
		_, _ = p.output.Write([]byte(CursorHide))
		frame := p.opts.Render(p)
		_, _ = p.output.Write([]byte(frame))
		p.State = StateActive
		p.prevFrame = frame
	}

	// Wait for completion
	result := <-p.done
	return result
}

// onKeypress handles keypress events
func (p *Prompt) onKeypress(char string, key Key) {
	// Handle user input tracking
	if p.track && key.Name != "return" {
		p.cursor = len(p.UserInput) // Simplified cursor tracking
		// In a real implementation, we'd handle more complex input tracking
	}

	// Reset error state on new input
	if p.State == StateError {
		p.State = StateActive
	}

	// Handle cursor movement
	if isMovementKey(key.Name) {
		p.Emit("cursor", key.Name)
	}

	// Handle movement key aliases when not tracking
	if !p.track {
		alias := getMovementAlias(key.Name)
		if alias != "" {
			p.Emit("cursor", alias)
		}
	}

	// Handle confirm keys
	if char != "" && (strings.ToLower(char) == "y" || strings.ToLower(char) == "n") {
		p.Emit("confirm", strings.ToLower(char) == "y")
	}

	// Emit key event
	p.Emit("key", strings.ToLower(char), key)

	// Handle return key
	if key.Name == "return" {
		if p.opts.Validate != nil {
			if err := p.opts.Validate(p.Value); err != nil {
				var validationErr *ValidationError
				if errors.As(err, &validationErr) {
					p.Error = validationErr.Message
				} else {
					p.Error = err.Error()
				}
				p.State = StateError
			}
		}

		if p.State != StateError {
			p.State = StateSubmit
		}
	}

	// Handle cancel (Ctrl+C)
	if isCancel(char, key) {
		p.State = StateCancel
	}

	// Handle finalization
	if p.State == StateSubmit || p.State == StateCancel {
		p.Emit("finalize")
	}

	p.render()

	// Close if done
	if p.State == StateSubmit || p.State == StateCancel {
		p.close()
	}
}

// setUserInput sets the user input and emits userInput event
func (p *Prompt) setUserInput(value string, _ bool) {
	p.UserInput = value
	p.Emit("userInput", p.UserInput)
	// In a real implementation, we'd handle writing to the input stream
}

// render renders the prompt output
func (p *Prompt) render() {
	if p.opts.Render == nil {
		return
	}

	frame := p.opts.Render(p)
	if frame == p.prevFrame {
		return
	}

	if p.State == StateInitial {
		_, _ = p.output.Write([]byte(CursorHide))
	}

	_, _ = p.output.Write([]byte(frame))

	if p.State == StateInitial {
		p.State = StateActive
	}

	p.prevFrame = frame
}

// close handles cleanup when the prompt is finished
func (p *Prompt) close() {
	_, _ = p.output.Write([]byte("\n"))
	_, _ = p.output.Write([]byte(CursorShow))

	var result any
	if p.State == StateCancel {
		result = GetCancelSymbol()
		p.Emit("cancel", result)
	} else {
		result = p.Value
		p.Emit("submit", result)
	}

	// Send result through done channel
	select {
	case p.done <- result:
	default:
	}
}

// Helper functions

func isMovementKey(keyName string) bool {
	return keyName == "up" || keyName == "down" || keyName == "left" || keyName == "right"
}

func isCancel(char string, key Key) bool {
	return char == "\x03" || (key.Ctrl && key.Name == "c")
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
