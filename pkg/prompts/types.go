package prompts

import "github.com/yarlson/tap/pkg/core"

// TextOptions defines options for styled text prompt
type TextOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
	Input        core.Reader
	Output       core.Writer
}

// ConfirmOptions defines options for styled confirm prompt
type ConfirmOptions struct {
	Message      string
	Active       string
	Inactive     string
	InitialValue bool
	Input        core.Reader
	Output       core.Writer
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
	Input        core.Reader
	Output       core.Writer
}
