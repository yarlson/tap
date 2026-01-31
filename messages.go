package tap

import (
	"fmt"
)

// MessageOptions configures simple message helpers output.
// If Output is nil, the helper functions are no-ops.
type MessageOptions struct {
	Output Writer
	Hint   string // Optional second line displayed in gray
}

// Cancel prints a cancel-styled message (bar end + red message).
func Cancel(message string, opts ...MessageOptions) {
	var (
		out  Writer
		hint string
	)

	if len(opts) > 0 {
		out = opts[0].Output
		hint = opts[0].Hint
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	if hint != "" {
		_, _ = fmt.Fprintf(out, "%s\n%s  %s\n   %s\n\n", gray(Bar), gray(BarEnd), red(message), gray(hint))
	} else {
		_, _ = fmt.Fprintf(out, "%s\n%s  %s\n\n", gray(Bar), gray(BarEnd), red(message))
	}
}

// Intro prints an intro title (bar start + title).
func Intro(title string, opts ...MessageOptions) {
	var (
		out  Writer
		hint string
	)

	if len(opts) > 0 {
		out = opts[0].Output
		hint = opts[0].Hint
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	if hint != "" {
		_, _ = fmt.Fprintf(out, "%s  %s\n%s  %s\n", gray(BarStart), bold(title), gray(Bar), gray(hint))
	} else {
		_, _ = fmt.Fprintf(out, "%s  %s\n%s\n", gray(BarStart), bold(title), gray(Bar))
	}
}

// Outro prints a final outro (bar line, then bar end + message).
func Outro(message string, opts ...MessageOptions) {
	var (
		out  Writer
		hint string
	)

	if len(opts) > 0 {
		out = opts[0].Output
		hint = opts[0].Hint
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	if hint != "" {
		_, _ = fmt.Fprintf(out, "%s\n%s  %s\n   %s\n\n", gray(Bar), gray(BarEnd), bold(message), gray(hint))
	} else {
		_, _ = fmt.Fprintf(out, "%s\n%s  %s\n\n", gray(Bar), gray(BarEnd), bold(message))
	}
}

func Message(message string, opts ...MessageOptions) {
	var (
		out  Writer
		hint string
	)

	if len(opts) > 0 {
		out = opts[0].Output
		hint = opts[0].Hint
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	_, _ = fmt.Fprintf(out, "%s\n", gray(Bar))
	_, _ = fmt.Fprintf(out, "%s  %s\n", green(StepSubmit), bold(message))

	if hint != "" {
		_, _ = fmt.Fprintf(out, "%s  %s\n", gray(Bar), gray(hint))
	} else {
		_, _ = fmt.Fprintf(out, "%s\n", gray(Bar))
	}
}
