package core

import (
	"strings"
)

// Password implements an unstyled password prompt that masks user input
func Password(opts PasswordOptions) any {
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
			userInput := p.UserInputSnapshot()
			cursor := p.CursorSnapshot()

			state := p.StateSnapshot()
			msg := opts.Message
			sep := ": "
			trimmed := strings.TrimRight(msg, " ")
			if strings.HasSuffix(trimmed, ":") {
				sep = " "
			}

			// Build masked display of input with cursor block
			masked := maskWithCursor(userInput, cursor, state)
			return msg + sep + masked
		},
	})

	p.On("userInput", func(input string) {
		p.SetImmediateValue(input)
	})

	return p.Prompt()
}

// maskWithCursor returns a string of mask characters the same length as the
// input, with an inverse block on the current cursor position when active.
func maskWithCursor(input string, cursor int, state ClackState) string {
	const bullet = "●"

	runes := []rune(input)
	masked := strings.Repeat(bullet, len(runes))

	if state != StateActive && state != StateInitial {
		return masked
	}

	// Show a block cursor after the last char if at end; otherwise invert the
	// mask char at the cursor. Use same approach as core.Text.
	const invOn = "\x1b[7m"
	const invOff = "\x1b[27m"
	const block = "█"

	if cursor >= len(runes) {
		if masked == "" {
			return invOn + " " + invOff // keep a visible cursor when empty
		}
		return masked + block
	}

	runesMasked := []rune(masked)
	return string(runesMasked[:cursor]) + invOn + string(runesMasked[cursor]) + invOff + string(runesMasked[cursor+1:])
}
