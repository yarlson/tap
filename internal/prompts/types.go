package prompts

import "io"

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

type Key struct {
	Name     string
	Sequence string
	Ctrl     bool
	Meta     bool
	Shift    bool
}

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
