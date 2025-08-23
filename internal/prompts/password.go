package prompts

import (
	"strings"

	"github.com/yarlson/tap/internal/core"
)

// Password creates a styled password input prompt that masks user input
func Password(opts PasswordOptions) string {
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

			// Title with symbol and message
			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			// Build masked display of input with cursor
			masked := renderMaskedWithCursor(userInput, cursor, s)

			switch s {
			case core.StateError:
				errMsg := p.ErrorSnapshot()
				return title + yellow(Bar) + "  " + masked + "\n" + yellow(BarEnd) + "  " + yellow(errMsg)

			case core.StateSubmit:
				// Do not show raw value; show bullets only
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}
				valueText := ""
				if value != "" {
					valueText = "  " + dim(strings.Repeat("●", len([]rune(value))))
				}
				return title + gray(Bar) + valueText

			case core.StateCancel:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}
				valueText := ""
				if strings.TrimSpace(value) != "" {
					valueText = "  " + strikethrough(dim(strings.Repeat("●", len([]rune(value)))))
				}
				result := title + gray(Bar) + valueText
				if strings.TrimSpace(value) != "" {
					result += "\n" + gray(Bar)
				}
				return result

			default:
				return title + cyan(Bar) + "  " + masked + "\n" + cyan(BarEnd)
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

// renderMaskedWithCursor renders bullets for each rune in input, and shows an inverted cursor block
// similar to the styled text behavior.
func renderMaskedWithCursor(text string, cursor int, state core.ClackState) string {
	if state != core.StateActive && state != core.StateInitial {
		return strings.Repeat("●", len([]rune(text)))
	}

	runes := []rune(text)
	maskedRunes := []rune(strings.Repeat("●", len(runes)))
	if cursor >= len(runes) {
		return string(maskedRunes) + inverse(" ")
	}

	before := string(maskedRunes[:cursor])
	char := string(maskedRunes[cursor])
	after := string(maskedRunes[cursor+1:])
	return before + inverse(char) + after
}
