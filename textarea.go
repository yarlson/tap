package tap

import (
	"context"
	"slices"
	"strings"
)

// Textarea creates a styled multiline text input prompt.
func Textarea(ctx context.Context, opts TextareaOptions) string {
	if opts.Input != nil && opts.Output != nil {
		return textarea(ctx, opts)
	}

	return runWithTerminal(func(in Reader, out Writer) string {
		if opts.Input == nil {
			opts.Input = in
		}

		if opts.Output == nil {
			opts.Output = out
		}

		return textarea(ctx, opts)
	})
}

func textarea(ctx context.Context, opts TextareaOptions) string {
	// Local buffer state (track=false, same pattern as Autocomplete)
	var (
		buf []rune
		cur int
	)

	p := NewPromptWithTracking(PromptOptions{
		Input:        opts.Input,
		Output:       opts.Output,
		InitialValue: opts.DefaultValue,
		Render: func(p *Prompt) string {
			s := p.StateSnapshot()

			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			switch s {
			case StateSubmit:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				if value == "" {
					return title + gray(Bar)
				}

				lines := strings.Split(value, "\n")
				var parts []string

				for _, line := range lines {
					parts = append(parts, gray(Bar)+"  "+dim(line))
				}

				return title + strings.Join(parts, "\n")

			case StateCancel:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				if value == "" {
					return title + gray(Bar)
				}

				lines := strings.Split(value, "\n")
				var parts []string

				for _, line := range lines {
					parts = append(parts, gray(Bar)+"  "+strikethrough(dim(line)))
				}

				return title + strings.Join(parts, "\n")

			default:
				// Active/Initial/Error state
				content := string(buf)
				barColor := cyan

				if s == StateError {
					barColor = yellow
				}

				if len(buf) == 0 && opts.Placeholder != "" {
					placeholder := renderTextareaPlaceholder(opts.Placeholder)
					result := title + barColor(Bar) + "  " + placeholder + "\n" + barColor(BarEnd)

					if s == StateError {
						errMsg := p.ErrorSnapshot()
						result = title + barColor(Bar) + "  " + placeholder + "\n" + barColor(BarEnd) + "  " + yellow(errMsg)
					}

					return result
				}

				lines := strings.Split(content, "\n")
				var parts []string

				for i, line := range lines {
					display := renderTextareaLine(line, buf, cur, i, s)
					parts = append(parts, barColor(Bar)+"  "+display)
				}

				result := title + strings.Join(parts, "\n") + "\n" + barColor(BarEnd)

				if s == StateError {
					errMsg := p.ErrorSnapshot()
					result = title + strings.Join(parts, "\n") + "\n" + barColor(BarEnd) + "  " + yellow(errMsg)
				}

				return result
			}
		},
	}, false)

	// Initialize from InitialValue if provided
	if opts.InitialValue != "" {
		buf = []rune(opts.InitialValue)
		cur = len(buf)
		p.SetImmediateValue(string(buf))
	}

	// Key handling
	p.On("key", func(_ string, key Key) {
		switch key.Name {
		case "left":
			if cur > 0 {
				cur--
			}

		case "right":
			if cur < len(buf) {
				cur++
			}

		case "backspace":
			if cur > 0 {
				buf = slices.Delete(buf, cur-1, cur)
				cur--
			}

		case "delete":
			if cur < len(buf) {
				buf = slices.Delete(buf, cur, cur+1)
			}

		case "return":
			val := string(buf)
			if val == "" && opts.DefaultValue != "" {
				val = opts.DefaultValue
			}

			p.SetValue(val)

			return

		default:
			if key.Rune >= 32 && key.Rune <= 126 {
				buf = slices.Insert(buf, cur, key.Rune)
				cur++
			}
		}

		p.SetImmediateValue(string(buf))
	})

	v := p.Prompt(ctx)
	if s, ok := v.(string); ok {
		return s
	}

	return ""
}

// renderTextareaPlaceholder renders placeholder text with inverse first char + dim rest.
func renderTextareaPlaceholder(placeholder string) string {
	runes := []rune(placeholder)
	if len(runes) == 0 {
		return inverse(" ")
	}

	return inverse(string(runes[0])) + dim(string(runes[1:]))
}

// renderTextareaLine renders a single line of textarea content with cursor display.
func renderTextareaLine(line string, buf []rune, cursor, lineIdx int, state ClackState) string {
	if state != StateActive && state != StateInitial {
		return line
	}

	// Calculate cursor position within this line
	lineStart := 0
	for i := 0; i < lineIdx; i++ {
		// Find the next newline to determine where this line starts
		idx := indexRune(buf[lineStart:], '\n')
		if idx < 0 {
			break
		}

		lineStart += idx + 1
	}

	lineEnd := lineStart + len([]rune(line))
	if cursor < lineStart || cursor > lineEnd {
		return line
	}

	// Cursor is on this line
	posInLine := cursor - lineStart
	runes := []rune(line)

	if posInLine >= len(runes) {
		return line + inverse(" ")
	}

	before := string(runes[:posInLine])
	char := string(runes[posInLine])
	after := string(runes[posInLine+1:])

	return before + inverse(char) + after
}

// indexRune returns the index of the first occurrence of r in s, or -1.
func indexRune(s []rune, r rune) int {
	for i, v := range s {
		if v == r {
			return i
		}
	}

	return -1
}
