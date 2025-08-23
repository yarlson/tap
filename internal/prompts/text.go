package prompts

import (
	"strings"

	"github.com/yarlson/tap/internal/core"
)

// Text creates a styled text input prompt
func Text(opts TextOptions) string {
	var validate func(any) error
	if opts.Validate != nil {
		validate = func(v any) error {
			str, _ := v.(string)
			return opts.Validate(str)
		}
	}

	p := core.NewPrompt(core.PromptOptions{
		Input:            opts.Input,
		Output:           opts.Output,
		Validate:         validate,
		InitialUserInput: opts.InitialValue,
		InitialValue:     opts.DefaultValue,
		Render: func(p *core.Prompt) string {
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
			case core.StateError:
				errorText := ""
				if err := p.ErrorSnapshot(); err != "" {
					errorText = " " + yellow("("+err+")")
				}
				return Symbol(s) + " " + opts.Message + " " + displayInput + errorText

			case core.StateSubmit:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}
				valueText := ""
				if value != "" {
					valueText = "  " + dim(value)
				}
				return title + gray(Bar) + valueText

			case core.StateCancel:
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

	v := p.Prompt()
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// renderTextWithCursor renders text with a cursor indicator
func renderTextWithCursor(text string, cursor int, state core.ClackState) string {
	if state != core.StateActive && state != core.StateInitial {
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
