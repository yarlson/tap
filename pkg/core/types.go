package core

import (
	"io"
)

// ClackState represents the state of the prompt
type ClackState string

const (
	StateInitial ClackState = "initial"
	StateActive  ClackState = "active"
	StateCancel  ClackState = "cancel"
	StateSubmit  ClackState = "submit"
	StateError   ClackState = "error"
)

// Key represents a keyboard key event
type Key struct {
	Name     string
	Sequence string
	Ctrl     bool
	Meta     bool
	Shift    bool
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Reader interface for input streams
type Reader interface {
	io.Reader
	On(event string, handler func(string, Key))
}

// Writer interface for output streams
type Writer interface {
	io.Writer
	On(event string, handler func())
	Emit(event string)
}

// ANSI escape codes for cursor manipulation
const (
	CursorHide = "\x1b[?25l"
	CursorShow = "\x1b[?25h"
)

// CancelSymbol is a unique symbol to represent cancellation
type CancelSymbol struct{}

var cancelSymbol = &CancelSymbol{}

// IsCancel checks if a value represents cancellation
func IsCancel(value any) bool {
	_, ok := value.(*CancelSymbol)
	return ok
}

// GetCancelSymbol returns the cancellation symbol
func GetCancelSymbol() *CancelSymbol {
	return cancelSymbol
}
