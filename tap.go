// Package tap provides high-level, clack-style terminal prompts, spinners,
// progress bars, and message helpers. The package exposes simple synchronous
// helper functions and manages a default interactive session under the hood.
package tap

import (
	"time"

	"github.com/yarlson/tap/internal/core"
	"github.com/yarlson/tap/internal/prompts"
	"github.com/yarlson/tap/internal/terminal"
)

var term *terminal.Terminal

func init() {
	term, _ = terminal.New()
}

// IsCancel reports whether v is the cancel sentinel returned when the user
// cancels a prompt. Use this to branch on user cancellation.
func IsCancel(v any) bool { return core.IsCancel(v) }

// TextOptions configures the Text prompt. I/O fields are managed by tap.
type TextOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
}

// Text displays an interactive single-line text input prompt and returns the
// entered value, or a CancelSymbol if the user cancels (check with IsCancel).
// A default session is created and cleaned up automatically if needed.
func Text(opts TextOptions) any {
	return prompts.Text(prompts.TextOptions{
		Message:      opts.Message,
		Placeholder:  opts.Placeholder,
		DefaultValue: opts.DefaultValue,
		InitialValue: opts.InitialValue,
		Validate:     opts.Validate,
		Input:        term.Reader,
		Output:       term.Writer,
	})
}

// PasswordOptions configures the Password prompt. I/O fields are managed by tap.
type PasswordOptions struct {
	Message      string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
}

// Password displays a masked text input prompt and returns the entered value,
// or a CancelSymbol if the user cancels (check with IsCancel).
// A default session is created and cleaned up automatically if needed.
func Password(opts PasswordOptions) any {
	return prompts.Password(prompts.PasswordOptions{
		Message:      opts.Message,
		DefaultValue: opts.DefaultValue,
		InitialValue: opts.InitialValue,
		Validate:     opts.Validate,
		Input:        term.Reader,
		Output:       term.Writer,
	})
}

// ConfirmOptions configures the Confirm prompt. I/O fields are managed by tap.
type ConfirmOptions struct {
	Message      string
	Active       string
	Inactive     string
	InitialValue bool
}

// Confirm displays a yes/no confirmation prompt and returns a bool indicating
// the choice, or a CancelSymbol if the user cancels (check with IsCancel).
// A default session is created and cleaned up automatically if needed.
func Confirm(opts ConfirmOptions) any {
	return prompts.Confirm(prompts.ConfirmOptions{
		Message:      opts.Message,
		Active:       opts.Active,
		Inactive:     opts.Inactive,
		InitialValue: opts.InitialValue,
		Input:        term.Reader,
		Output:       term.Writer,
	})
}

// SelectOption represents a selectable item with a typed value, label, and
// optional hint for display.
type SelectOption[T any] struct {
	Value T
	Label string
	Hint  string
}

// SelectOptions configures the Select prompt. I/O fields are managed by tap.
type SelectOptions[T any] struct {
	Message      string
	Options      []SelectOption[T]
	InitialValue *T
	MaxItems     *int
}

// Select displays a single-selection list and returns the chosen typed value,
// or a CancelSymbol if the user cancels (check with IsCancel).
// A default session is created and cleaned up automatically if needed.
func Select[T any](opts SelectOptions[T]) any {
	items := make([]prompts.SelectOption[T], len(opts.Options))
	for i, o := range opts.Options {
		items[i] = prompts.SelectOption[T]{Value: o.Value, Label: o.Label, Hint: o.Hint}
	}

	return prompts.Select(prompts.SelectOptions[T]{
		Message:      opts.Message,
		Options:      items,
		InitialValue: opts.InitialValue,
		MaxItems:     opts.MaxItems,
		Input:        term.Reader,
		Output:       term.Writer,
	})
}

// SpinnerOptions configures a spinner. Output is managed by tap.
type SpinnerOptions struct {
	Indicator     string
	Frames        []string
	Delay         time.Duration
	CancelMessage string
	ErrorMessage  string
}

// NewSpinner creates a spinner that writes to the current session'term writer, or
// to stdout if no session is active.
func NewSpinner(opts SpinnerOptions) *prompts.Spinner {
	return prompts.NewSpinner(prompts.SpinnerOptions{
		Indicator:     opts.Indicator,
		Frames:        opts.Frames,
		Delay:         opts.Delay,
		Output:        term.Writer,
		CancelMessage: opts.CancelMessage,
		ErrorMessage:  opts.ErrorMessage,
	})
}

// ProgressOptions configures a progress bar. Output is managed by tap.
type ProgressOptions struct {
	Style string
	Max   int
	Size  int
}

// NewProgress creates a progress bar that writes to the current session'term
// writer, or to stdout if no session is active.
func NewProgress(opts ProgressOptions) *prompts.Progress {
	return prompts.NewProgress(prompts.ProgressOptions{
		Style:  opts.Style,
		Max:    opts.Max,
		Size:   opts.Size,
		Output: term.Writer,
	})
}

// Intro prints an introductory message using the current session writer or
// stdout if no session is active.
func Intro(title string) {
	prompts.Intro(title, prompts.MessageOptions{Output: term.Writer})
}

// Outro prints a closing message using the current session writer or stdout if
// no session is active.
func Outro(message string) {
	prompts.Outro(message, prompts.MessageOptions{Output: term.Writer})
}

// Cancel prints a cancellation message using the current session writer or
// stdout if no session is active.
func Cancel(message string) {
	prompts.Cancel(message, prompts.MessageOptions{Output: term.Writer})
}

// BoxAlignment is an alias of prompts.BoxAlignment to control box content
// alignment.
type BoxAlignment = prompts.BoxAlignment

// BoxOptions configures the Box message renderer.
type BoxOptions struct {
	Columns        int
	WidthFraction  float64
	WidthAuto      bool
	TitlePadding   int
	ContentPadding int
	TitleAlign     BoxAlignment
	ContentAlign   BoxAlignment
	Rounded        bool
	IncludePrefix  bool
	FormatBorder   func(string) string
}

// Box renders a framed message with optional title and alignment using the
// current session writer or stdout if no session is active.
func Box(message string, title string, opts BoxOptions) {
	prompts.Box(message, title, prompts.BoxOptions{
		Output:         term.Writer,
		Columns:        opts.Columns,
		WidthFraction:  opts.WidthFraction,
		WidthAuto:      opts.WidthAuto,
		TitlePadding:   opts.TitlePadding,
		ContentPadding: opts.ContentPadding,
		TitleAlign:     opts.TitleAlign,
		ContentAlign:   opts.ContentAlign,
		Rounded:        opts.Rounded,
		IncludePrefix:  opts.IncludePrefix,
		FormatBorder:   opts.FormatBorder,
	})
}

// GrayBorder formats a string with a gray box-drawing border.
func GrayBorder(s string) string { return prompts.GrayBorder(s) }

// CyanBorder formats a string with a cyan box-drawing border.
func CyanBorder(s string) string { return prompts.CyanBorder(s) }
