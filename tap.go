// Package tap provides high-level, clack-style terminal prompts, spinners,
// progress bars, and message helpers. The package exposes simple synchronous
// helper functions and manages a default interactive session under the hood.
package tap

import (
	"time"

	"github.com/yarlson/tap/internal/prompts"
)

// SetTermIO sets a custom reader and writer used by helpers. Pass nil values to
// restore default terminal behavior.
func SetTermIO(in prompts.Reader, out prompts.Writer) { prompts.SetTermIO(in, out) }

// TextOptions configures the Text prompt. I/O fields are managed by tap.
type TextOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
}

// Text displays an interactive single-line text input prompt and returns the
// entered value. A terminal is created and cleaned up automatically per call.
func Text(opts TextOptions) string {
	return prompts.Text(prompts.TextOptions{
		Message:      opts.Message,
		Placeholder:  opts.Placeholder,
		DefaultValue: opts.DefaultValue,
		InitialValue: opts.InitialValue,
		Validate:     opts.Validate,
	})
}

// PasswordOptions configures the Password prompt. I/O fields are managed by tap.
type PasswordOptions struct {
	Message      string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
}

// Password displays a masked text input prompt and returns the entered value.
// A terminal is created and cleaned up automatically per call.
func Password(opts PasswordOptions) string {
	return prompts.Password(prompts.PasswordOptions{
		Message:      opts.Message,
		DefaultValue: opts.DefaultValue,
		InitialValue: opts.InitialValue,
		Validate:     opts.Validate,
	})
}

// ConfirmOptions configures the Confirm prompt. I/O fields are managed by tap.
type ConfirmOptions struct {
	Message      string
	Active       string
	Inactive     string
	InitialValue bool
}

// Confirm displays a yes/no confirmation prompt and returns the selection.
// A terminal is created and cleaned up automatically per call.
func Confirm(opts ConfirmOptions) bool {
	return prompts.Confirm(prompts.ConfirmOptions{
		Message:      opts.Message,
		Active:       opts.Active,
		Inactive:     opts.Inactive,
		InitialValue: opts.InitialValue,
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

// Select displays a single-selection list and returns the chosen typed value.
// A terminal is created and cleaned up automatically per call.
func Select[T any](opts SelectOptions[T]) T {
	items := make([]prompts.SelectOption[T], len(opts.Options))
	for i, o := range opts.Options {
		items[i] = prompts.SelectOption[T]{Value: o.Value, Label: o.Label, Hint: o.Hint}
	}

	return prompts.Select[T](prompts.SelectOptions[T]{
		Message:      opts.Message,
		Options:      items,
		InitialValue: opts.InitialValue,
		MaxItems:     opts.MaxItems,
	})
}

// MultiSelectOptions configures the MultiSelect prompt. I/O fields are managed by tap.
type MultiSelectOptions[T any] struct {
	Message       string
	Options       []SelectOption[T]
	InitialValues []T
	MaxItems      *int
}

// MultiSelect displays a multi-selection list and returns the chosen typed values.
// A terminal is created and cleaned up automatically per call.
func MultiSelect[T any](opts MultiSelectOptions[T]) []T {
	items := make([]prompts.SelectOption[T], len(opts.Options))
	for i, o := range opts.Options {
		items[i] = prompts.SelectOption[T]{Value: o.Value, Label: o.Label, Hint: o.Hint}
	}

	return prompts.MultiSelect[T](prompts.MultiSelectOptions[T]{
		Message:       opts.Message,
		Options:       items,
		InitialValues: opts.InitialValues,
		MaxItems:      opts.MaxItems,
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

// Spinner wraps a spinner and ensures terminal cleanup on Stop.
type Spinner = prompts.Spinner

// NewSpinner creates a spinner bound to a terminal writer (or the override
// writer set via SetTermIO in tests). The underlying terminal, when created,
// is cleaned up on Stop.
func NewSpinner(opts SpinnerOptions) *Spinner {
	return prompts.NewSpinner(prompts.SpinnerOptions{
		Indicator:     opts.Indicator,
		Frames:        opts.Frames,
		Delay:         opts.Delay,
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

// Progress wraps a progress bar and ensures terminal cleanup on Stop.
type Progress = prompts.Progress

// NewProgress creates a progress bar bound to a terminal writer (or the
// override writer set via SetTermIO in tests). The underlying terminal, when
// created, is cleaned up on Stop.
func NewProgress(opts ProgressOptions) *Progress {
	return prompts.NewProgress(prompts.ProgressOptions{
		Style: opts.Style,
		Max:   opts.Max,
		Size:  opts.Size,
	})
}

// StreamOptions configures a live output stream. Output is managed by tap.
type StreamOptions struct {
	ShowTimer bool
}

// Stream wraps a styled live stream renderer and ensures terminal cleanup on Stop.
type Stream = prompts.Stream

// NewStream creates a live stream bound to a terminal writer (or override),
// and ensures the underlying terminal is closed on Stop.
func NewStream(opts StreamOptions) *Stream {
	return prompts.NewStream(prompts.StreamOptions{
		ShowTimer: opts.ShowTimer,
	})
}

// Intro prints an introductory message using the current session writer or
// stdout if no session is active.
func Intro(title string) {
	prompts.RunWithTerminal(func(_ prompts.Reader, out prompts.Writer) any {
		prompts.Intro(title, prompts.MessageOptions{Output: out})
		return nil
	})
}

// Outro prints a closing message using the current session writer or stdout if
// no session is active.
func Outro(message string) {
	prompts.RunWithTerminal(func(_ prompts.Reader, out prompts.Writer) any {
		prompts.Outro(message, prompts.MessageOptions{Output: out})
		return nil
	})
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
	prompts.RunWithTerminal(func(_ prompts.Reader, out prompts.Writer) any {
		prompts.Box(message, title, prompts.BoxOptions{
			Output:         out,
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
		return nil
	})
}

// GrayBorder formats a string with a gray box-drawing border.
func GrayBorder(s string) string { return prompts.GrayBorder(s) }

// CyanBorder formats a string with a cyan box-drawing border.
func CyanBorder(s string) string { return prompts.CyanBorder(s) }
