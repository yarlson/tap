package core

import (
	"io"
)

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
)

type CancelSymbol struct{}

var cancelSymbol = &CancelSymbol{}

func IsCancel(value any) bool {
	_, ok := value.(*CancelSymbol)
	return ok
}

func GetCancelSymbol() *CancelSymbol {
	return cancelSymbol
}
