package prompts

import (
	"io"

	"github.com/yarlson/tap/internal/terminal"
)

// Type aliases for convenience

// TextOptions defines options for styled text prompt
type TextOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
	Input        Reader
	Output       Writer
}

// PasswordOptions defines options for styled password prompt
type PasswordOptions struct {
	Message      string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
	Input        Reader
	Output       Writer
}

// ConfirmOptions defines options for styled confirm prompt
type ConfirmOptions struct {
	Message      string
	Active       string
	Inactive     string
	InitialValue bool
	Input        Reader
	Output       Writer
}

// SelectOption represents an option in a styled select prompt
type SelectOption[T any] struct {
	Value T
	Label string
	Hint  string
}

// SelectOptions defines options for styled select prompt
type SelectOptions[T any] struct {
	Message      string
	Options      []SelectOption[T]
	InitialValue *T
	MaxItems     *int
	Input        Reader
	Output       Writer
}

// MultiSelectOptions defines options for styled multi-select prompt
type MultiSelectOptions[T any] struct {
	Message       string
	Options       []SelectOption[T]
	InitialValues []T
	MaxItems      *int
	Input         Reader
	Output        Writer
}

type ClackState string

const (
	StateInitial ClackState = "initial"
	StateActive  ClackState = "active"
	StateCancel  ClackState = "cancel"
	StateSubmit  ClackState = "submit"
	StateError   ClackState = "error"
)

type Key = terminal.Key

type ValidationError struct {
	Message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

func (e *ValidationError) Error() string {
	return e.Message
}

type Reader interface {
	io.Reader
	On(event string, handler func(string, Key))
}

type Writer interface {
	io.Writer
	On(event string, handler func())
	Emit(event string)
}

const (
	CursorHide = "\x1b[?25l"
	CursorShow = "\x1b[?25h"
	EraseLine  = "\x1b[K"
	CursorUp   = "\x1b[A"
	EraseDown  = "\x1b[J"
)

// Optional test I/O override. When set, helpers use these instead of opening
// a real terminal.
var (
	ioReader Reader
	ioWriter Writer
)

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
	defer t.Close()

	return fn(t.Reader, t.Writer)
}

// resolveWriter returns the output writer and an optional terminal to close.
func resolveWriter() (Writer, *terminal.Terminal) {
	// Check if we have override I/O set
	if out := getOverrideWriter(); out != nil {
		return out, nil
	}

	// Need to create a new terminal
	t, err := terminal.New()
	if err != nil {
		return nil, nil
	}

	return t.Writer, t
}

// getOverrideWriter returns the override writer if set
func getOverrideWriter() Writer {
	return runWithTerminal(func(in Reader, out Writer) Writer {
		return out
	})
}
