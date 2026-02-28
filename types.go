package tap

import (
	"io"

	"github.com/yarlson/tap/internal/terminal"
)

// Type aliases for convenience

// TextOptions defines options for styled text prompt.
type TextOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
	Input        Reader
	Output       Writer
}

// PasswordOptions defines options for styled password prompt.
type PasswordOptions struct {
	Message      string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
	Input        Reader
	Output       Writer
}

// ConfirmOptions defines options for styled confirm prompt.
type ConfirmOptions struct {
	Message      string
	Active       string
	Inactive     string
	InitialValue bool
	Input        Reader
	Output       Writer
}

// SelectOption represents an option in a styled select prompt.
type SelectOption[T any] struct {
	Value T
	Label string
	Hint  string
}

// SelectOptions defines options for styled select prompt.
type SelectOptions[T any] struct {
	Message      string
	Options      []SelectOption[T]
	InitialValue *T
	MaxItems     *int
	Input        Reader
	Output       Writer
}

// MultiSelectOptions defines options for styled multi-select prompt.
type MultiSelectOptions[T any] struct {
	Message       string
	Options       []SelectOption[T]
	InitialValues []T
	MaxItems      *int
	Input         Reader
	Output        Writer
}

// AutocompleteOptions defines options for styled autocomplete text prompt.
type AutocompleteOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
	Suggest      func(string) []string // returns suggestion list for current input
	MaxResults   int                   // maximum suggestions to show (default 5)
	Input        Reader
	Output       Writer
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
	CursorHide = terminal.CursorHide
	CursorShow = terminal.CursorShow
	EraseLine  = terminal.ClearLine
	CursorUp   = terminal.CursorUp
	EraseDown  = terminal.EraseDown
)

// StopOptions configures the Stop behavior for Spinner and Progress.
type StopOptions struct {
	Hint string // Optional second line displayed in gray below the message
}

// Optional test I/O override. When set, helpers use these instead of opening
// a real terminal.
var (
	ioReader Reader
	ioWriter Writer
)

// Table-related types

type TableAlignment string

const (
	TableAlignLeft   TableAlignment = "left"
	TableAlignCenter TableAlignment = "center"
	TableAlignRight  TableAlignment = "right"
)

type TableStyle string

const (
	TableStyleNormal TableStyle = "normal"
	TableStyleBold   TableStyle = "bold"
	TableStyleDim    TableStyle = "dim"
)

type TableColor string

const (
	TableColorDefault TableColor = "default"
	TableColorGray    TableColor = "gray"
	TableColorRed     TableColor = "red"
	TableColorGreen   TableColor = "green"
	TableColorYellow  TableColor = "yellow"
	TableColorCyan    TableColor = "cyan"
)

// TableOptions defines options for styled table rendering.
type TableOptions struct {
	Output           Writer
	ShowBorders      bool
	IncludePrefix    bool
	MaxWidth         int
	ColumnAlignments []TableAlignment
	HeaderStyle      TableStyle
	HeaderColor      TableColor
	FormatBorder     func(string) string
}
