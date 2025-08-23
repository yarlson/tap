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

// Optional test I/O override. When set, helpers use these instead of opening
// a real terminal.
var (
	ioReader core.Reader
	ioWriter core.Writer
)

// SetTermIO sets a custom reader and writer used by helpers. Pass nil values to
// restore default terminal behavior.
func SetTermIO(in core.Reader, out core.Writer) { ioReader, ioWriter = in, out }

// runWithTerminal creates a temporary terminal for interactive prompts and
// ensures cleanup after the prompt completes.
func runWithTerminal[T any](fn func(core.Reader, core.Writer) T) T {
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
	return runWithTerminal(func(in core.Reader, out core.Writer) string {
		return prompts.Text(prompts.TextOptions{
			Message:      opts.Message,
			Placeholder:  opts.Placeholder,
			DefaultValue: opts.DefaultValue,
			InitialValue: opts.InitialValue,
			Validate:     opts.Validate,
			Input:        in,
			Output:       out,
		})
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
	return runWithTerminal(func(in core.Reader, out core.Writer) string {
		return prompts.Password(prompts.PasswordOptions{
			Message:      opts.Message,
			DefaultValue: opts.DefaultValue,
			InitialValue: opts.InitialValue,
			Validate:     opts.Validate,
			Input:        in,
			Output:       out,
		})
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
	return runWithTerminal(func(in core.Reader, out core.Writer) bool {
		return prompts.Confirm(prompts.ConfirmOptions{
			Message:      opts.Message,
			Active:       opts.Active,
			Inactive:     opts.Inactive,
			InitialValue: opts.InitialValue,
			Input:        in,
			Output:       out,
		})
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

	return runWithTerminal(func(in core.Reader, out core.Writer) T {
		return prompts.Select[T](prompts.SelectOptions[T]{
			Message:      opts.Message,
			Options:      items,
			InitialValue: opts.InitialValue,
			MaxItems:     opts.MaxItems,
			Input:        in,
			Output:       out,
		})
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
type Spinner struct {
	inner *prompts.Spinner
	term  *terminal.Terminal
}

// Start begins the spinner with an initial message.
func (s *Spinner) Start(msg string) { s.inner.Start(msg) }

// Message updates the spinner message.
func (s *Spinner) Message(msg string) { s.inner.Message(msg) }

// IsCancelled reports whether the spinner was cancelled by the user.
// Deprecated: Use IsCanceled for Go-idiomatic spelling.
func (s *Spinner) IsCancelled() bool { return s.inner.IsCancelled() }

// IsCanceled reports whether the spinner was canceled by the user.
func (s *Spinner) IsCanceled() bool { return s.inner.IsCancelled() }

// Stop stops the spinner with a final message and exit code (0=success, 1=cancel, >1=error).
func (s *Spinner) Stop(msg string, code int) {
	s.inner.Stop(msg, code)
	if s.term != nil {
		s.term.Close()
		s.term = nil
	}
}

// NewSpinner creates a spinner bound to a terminal writer (or the override
// writer set via SetTermIO in tests). The underlying terminal, when created,
// is cleaned up on Stop.
func NewSpinner(opts SpinnerOptions) *Spinner {
	out, term := resolveWriter()
	sp := prompts.NewSpinner(prompts.SpinnerOptions{
		Indicator:     opts.Indicator,
		Frames:        opts.Frames,
		Delay:         opts.Delay,
		Output:        out,
		CancelMessage: opts.CancelMessage,
		ErrorMessage:  opts.ErrorMessage,
	})

	return &Spinner{inner: sp, term: term}
}

// ProgressOptions configures a progress bar. Output is managed by tap.
type ProgressOptions struct {
	Style string
	Max   int
	Size  int
}

// Progress wraps a progress bar and ensures terminal cleanup on Stop.
type Progress struct {
	inner *prompts.Progress
	term  *terminal.Terminal
}

// Start begins the progress bar with an initial message.
func (p *Progress) Start(msg string) { p.inner.Start(msg) }

// Advance moves the progress bar forward by step and updates the message.
func (p *Progress) Advance(step int, msg string) { p.inner.Advance(step, msg) }

// Message updates the progress bar message.
func (p *Progress) Message(msg string) { p.inner.Message(msg) }

// Stop stops the progress bar with a final message and exit code (0=success, 1=cancel, >1=error).
func (p *Progress) Stop(msg string, code int) {
	p.inner.Stop(msg, code)
	if p.term != nil {
		p.term.Close()
		p.term = nil
	}
}

// NewProgress creates a progress bar bound to a terminal writer (or the
// override writer set via SetTermIO in tests). The underlying terminal, when
// created, is cleaned up on Stop.
func NewProgress(opts ProgressOptions) *Progress {
	out, term := resolveWriter()
	pr := prompts.NewProgress(prompts.ProgressOptions{
		Style:  opts.Style,
		Max:    opts.Max,
		Size:   opts.Size,
		Output: out,
	})

	return &Progress{inner: pr, term: term}
}

// resolveWriter returns the output writer and an optional terminal to close.
func resolveWriter() (core.Writer, *terminal.Terminal) {
	if ioWriter != nil {
		return ioWriter, nil
	}

	t, err := terminal.New()
	if err != nil {
		return nil, nil
	}

	return t.Writer, t
}

// Intro prints an introductory message using the current session writer or
// stdout if no session is active.
func Intro(title string) {
	_ = runWithTerminal(func(_ core.Reader, out core.Writer) any {
		prompts.Intro(title, prompts.MessageOptions{Output: out})
		return nil
	})
}

// Outro prints a closing message using the current session writer or stdout if
// no session is active.
func Outro(message string) {
	_ = runWithTerminal(func(_ core.Reader, out core.Writer) any {
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
	_ = runWithTerminal(func(_ core.Reader, out core.Writer) any {
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
