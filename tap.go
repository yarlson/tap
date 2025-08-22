package tap

import (
    "sync"
    "time"

    "github.com/yarlson/tap/core"
    "github.com/yarlson/tap/prompts"
    "github.com/yarlson/tap/terminal"
)

// Session owns a terminal and provides high-level prompt helpers
// that hide reader/writer from the caller.
type Session struct {
	term *terminal.Terminal
}

// New creates a new Session with its own terminal.
func New() (*Session, error) {
	term, err := terminal.New()
	if err != nil {
		return nil, err
	}
	return &Session{term: term}, nil
}

// Close releases session resources.
func (s *Session) Close() {
	if s != nil && s.term != nil {
		s.term.Close()
	}
}

var (
	defaultMu      sync.Mutex
	defaultSession *Session
)

// Init initializes a default session for package-level helpers.
func Init() (*Session, error) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultSession != nil {
		return defaultSession, nil
	}
	s, err := New()
	if err != nil {
		return nil, err
	}
	defaultSession = s
	return s, nil
}

// CloseDefault closes the default session, if any.
func CloseDefault() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultSession != nil {
		defaultSession.Close()
		defaultSession = nil
	}
}

func ensureDefault() *Session {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultSession == nil {
		// Best-effort init; ignore error to keep API simple.
		if s, err := New(); err == nil {
			defaultSession = s
		}
	}
	return defaultSession
}

// Re-export cancel helpers for convenience.

type CancelSymbol = core.CancelSymbol

func IsCancel(v any) bool { return core.IsCancel(v) }

// TextOptions mirrors prompts.TextOptions but without Input/Output.
type TextOptions struct {
	Message      string
	Placeholder  string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
}

func (s *Session) Text(opts TextOptions) any {
	return prompts.Text(prompts.TextOptions{
		Message:      opts.Message,
		Placeholder:  opts.Placeholder,
		DefaultValue: opts.DefaultValue,
		InitialValue: opts.InitialValue,
		Validate:     opts.Validate,
		Input:        s.term.Reader,
		Output:       s.term.Writer,
	})
}

func Text(opts TextOptions) any {
	s := ensureDefault()
	if s == nil {
		return core.GetCancelSymbol()
	}
	return s.Text(opts)
}

// PasswordOptions mirrors prompts.PasswordOptions but without Input/Output.
type PasswordOptions struct {
	Message      string
	DefaultValue string
	InitialValue string
	Validate     func(string) error
}

func (s *Session) Password(opts PasswordOptions) any {
	return prompts.Password(prompts.PasswordOptions{
		Message:      opts.Message,
		DefaultValue: opts.DefaultValue,
		InitialValue: opts.InitialValue,
		Validate:     opts.Validate,
		Input:        s.term.Reader,
		Output:       s.term.Writer,
	})
}

func Password(opts PasswordOptions) any {
	s := ensureDefault()
	if s == nil {
		return core.GetCancelSymbol()
	}
	return s.Password(opts)
}

// ConfirmOptions mirrors prompts.ConfirmOptions but without Input/Output.
type ConfirmOptions struct {
	Message      string
	Active       string
	Inactive     string
	InitialValue bool
}

func (s *Session) Confirm(opts ConfirmOptions) any {
	return prompts.Confirm(prompts.ConfirmOptions{
		Message:      opts.Message,
		Active:       opts.Active,
		Inactive:     opts.Inactive,
		InitialValue: opts.InitialValue,
		Input:        s.term.Reader,
		Output:       s.term.Writer,
	})
}

func Confirm(opts ConfirmOptions) any {
	s := ensureDefault()
	if s == nil {
		return core.GetCancelSymbol()
	}
	return s.Confirm(opts)
}

// SelectOption mirrors prompts.SelectOption.
type SelectOption[T any] struct {
	Value T
	Label string
	Hint  string
}

// SelectOptions mirrors prompts.SelectOptions but without Input/Output.
type SelectOptions[T any] struct {
	Message      string
	Options      []SelectOption[T]
	InitialValue *T
	MaxItems     *int
}

func Select[T any](opts SelectOptions[T]) any {
	s := ensureDefault()
	if s == nil {
		return core.GetCancelSymbol()
	}
    items := make([]prompts.SelectOption[T], len(opts.Options))
    for i, o := range opts.Options {
        items[i] = prompts.SelectOption[T]{Value: o.Value, Label: o.Label, Hint: o.Hint}
    }
    return prompts.Select(prompts.SelectOptions[T]{
        Message:      opts.Message,
        Options:      items,
        InitialValue: opts.InitialValue,
        MaxItems:     opts.MaxItems,
        Input:        s.term.Reader,
        Output:       s.term.Writer,
    })
}

// SpinnerOptions mirrors prompts.SpinnerOptions but without Output.
type SpinnerOptions struct {
	Indicator     string
	Frames        []string
    Delay         time.Duration
	CancelMessage string
	ErrorMessage  string
}

// NewSpinner creates a spinner bound to the session's output.
func (s *Session) NewSpinner(opts SpinnerOptions) *prompts.Spinner {
	po := prompts.SpinnerOptions{
		Indicator:     opts.Indicator,
		Frames:        opts.Frames,
		Output:        s.term.Writer,
		CancelMessage: opts.CancelMessage,
		ErrorMessage:  opts.ErrorMessage,
	}
    po.Delay = opts.Delay
	return prompts.NewSpinner(po)
}

func NewSpinner(opts SpinnerOptions) *prompts.Spinner {
	s := ensureDefault()
	if s == nil {
		return prompts.NewSpinner(prompts.SpinnerOptions{})
	}
	return s.NewSpinner(opts)
}

// ProgressOptions mirrors prompts.ProgressOptions but without Output.
type ProgressOptions struct {
	Style string
	Max   int
	Size  int
}

func (s *Session) NewProgress(opts ProgressOptions) *prompts.Progress {
	return prompts.NewProgress(prompts.ProgressOptions{
		Style:  opts.Style,
		Max:    opts.Max,
		Size:   opts.Size,
		Output: s.term.Writer,
	})
}

func NewProgress(opts ProgressOptions) *prompts.Progress {
	s := ensureDefault()
	if s == nil {
		return prompts.NewProgress(prompts.ProgressOptions{})
	}
	return s.NewProgress(opts)
}

// Message helpers bound to the session writer.
func (s *Session) Intro(title string) {
	prompts.Intro(title, prompts.MessageOptions{Output: s.term.Writer})
}

func Intro(title string) {
	s := ensureDefault()
	if s == nil {
		return
	}
	s.Intro(title)
}

func (s *Session) Outro(message string) {
	prompts.Outro(message, prompts.MessageOptions{Output: s.term.Writer})
}

func Outro(message string) {
	s := ensureDefault()
	if s == nil {
		return
	}
	s.Outro(message)
}

func (s *Session) Cancel(message string) {
	prompts.Cancel(message, prompts.MessageOptions{Output: s.term.Writer})
}

func Cancel(message string) {
	s := ensureDefault()
	if s == nil {
		return
	}
	s.Cancel(message)
}

// Box wrappers to render framed messages via high-level API

type BoxAlignment = prompts.BoxAlignment

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

func (s *Session) Box(message string, title string, opts BoxOptions) {
    prompts.Box(message, title, prompts.BoxOptions{
        Output:         s.term.Writer,
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

func Box(message string, title string, opts BoxOptions) {
    s := ensureDefault()
    if s == nil {
        return
    }
    s.Box(message, title, opts)
}

// Re-export common border formatters for convenience
func GrayBorder(s string) string { return prompts.GrayBorder(s) }
func CyanBorder(s string) string { return prompts.CyanBorder(s) }
