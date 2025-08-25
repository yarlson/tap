package tap

import (
	"fmt"
)

// MessageOptions configures simple message helpers output.
// If Output is nil, the helper functions are no-ops.
type MessageOptions struct {
	Output Writer
}

// Cancel prints a cancel-styled message (bar end + red message).
func Cancel(message string, opts ...MessageOptions) {
	var out Writer
	if len(opts) > 0 {
		out = opts[0].Output
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	_, _ = fmt.Fprintf(out, "%s  %s\n\n", gray(BarEnd), red(message))
}

// Intro prints an intro title (bar start + title).
func Intro(title string, opts ...MessageOptions) {
	var out Writer
	if len(opts) > 0 {
		out = opts[0].Output
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	_, _ = fmt.Fprintf(out, "%s  %s\n", gray(BarStart), title)
}

// Outro prints a final outro (bar line, then bar end + message).
func Outro(message string, opts ...MessageOptions) {
	var out Writer
	if len(opts) > 0 {
		out = opts[0].Output
	}

	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	_, _ = fmt.Fprintf(out, "%s\n%s  %s\n\n", gray(Bar), gray(BarEnd), message)
}
