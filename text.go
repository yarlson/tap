package tap

import (
	"context"
	"strings"

	"github.com/yarlson/tap/internal/terminal"
)

// Text creates a styled text input prompt
func Text(ctx context.Context, opts TextOptions) string {
	if opts.Input != nil && opts.Output != nil {
		return text(ctx, opts, nil)
	}

	return runWithTerminalAndRef(func(in Reader, out Writer, term *terminal.Terminal) string {
		if opts.Input == nil {
			opts.Input = in
		}

		if opts.Output == nil {
			opts.Output = out
		}

		return text(ctx, opts, term)
	})
}

// text implements the core text prompt logic
func text(ctx context.Context, opts TextOptions, term *terminal.Terminal) string {
	var validate func(any) error
	if opts.Validate != nil {
		validate = func(v any) error {
			str, _ := v.(string)
			return opts.Validate(str)
		}
	}

	p := NewPrompt(PromptOptions{
		Input:            opts.Input,
		Output:           opts.Output,
		Validate:         validate,
		InitialUserInput: opts.InitialValue,
		InitialValue:     opts.DefaultValue,
		Render: func(p *Prompt) string {
			s := p.StateSnapshot()
			userInput := p.UserInputSnapshot()
			cursor := p.CursorSnapshot()

			// Create title with symbol and message
			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			// Handle placeholder and cursor
			var displayInput string

			if userInput == "" && opts.Placeholder != "" {
				// Show placeholder with inverted first character
				if len(opts.Placeholder) > 0 {
					displayInput = inverse(string(opts.Placeholder[0])) + dim(opts.Placeholder[1:])
				} else {
					displayInput = inverse(" ")
				}
			} else {
				// Show user input with cursor
				displayInput = renderTextWithCursor(userInput, cursor, s)
			}

			switch s {
			case StateError:
				errMsg := p.ErrorSnapshot()
				return title + yellow(Bar) + "  " + displayInput + "\n" + yellow(BarEnd) + "  " + yellow(errMsg)

			case StateSubmit:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				valueText := ""
				if value != "" {
					valueText = "  " + dim(value)
				}

				return title + gray(Bar) + valueText

			case StateCancel:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				valueText := ""
				if value != "" {
					valueText = "  " + strikethrough(dim(value))
				}

				result := title + gray(Bar) + valueText
				if strings.TrimSpace(value) != "" {
					result += "\n" + gray(Bar)
				}

				return result

			default:
				return title + cyan(Bar) + "  " + displayInput + "\n" + cyan(BarEnd)
			}
		},
	})

	p.On("userInput", func(input string) {
		p.SetImmediateValue(input)
	})

	// Set terminal reference so prompt can listen for Ctrl+C
	if term != nil {
		p.SetTerminal(term)
	}

	v := p.Prompt(ctx)
	if s, ok := v.(string); ok {
		return s
	}

	return ""
}

// renderTextWithCursor renders text with a cursor indicator
func renderTextWithCursor(text string, cursor int, state ClackState) string {
	if state != StateActive && state != StateInitial {
		return text
	}

	runes := []rune(text)
	if cursor >= len(runes) {
		return text + inverse(" ")
	}

	before := string(runes[:cursor])
	char := string(runes[cursor])
	after := string(runes[cursor+1:])

	return before + inverse(char) + after
}
