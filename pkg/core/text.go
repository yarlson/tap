package core

import (
	"strings"
)

func Text(opts TextOptions) any {
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
		Render: func(p *Prompt) string {
			userInput := p.UserInputSnapshot()
			cursor := p.CursorSnapshot()

			const invOn = "\x1b[7m"
			const invOff = "\x1b[27m"
			const block = "█"

			state := p.StateSnapshot()
			var withCursor string
			if state == StateActive || state == StateInitial {
				runes := []rune(userInput)
				if cursor >= len(runes) {
					withCursor = userInput + block
				} else {
					withCursor = string(runes[:cursor]) + invOn + string(runes[cursor]) + invOff + string(runes[cursor+1:])
				}
			} else {
				withCursor = userInput
			}

			msg := opts.Message
			sep := ": "
			trimmed := strings.TrimRight(msg, " ")
			if strings.HasSuffix(trimmed, ":") {
				sep = " "
			}
			return msg + sep + withCursor
		},
	})

	p.On("userInput", func(input string) {
		p.SetImmediateValue(input)
	})

	p.On("finalize", func() {
		if currentValue := p.StateSnapshot(); currentValue != StateCancel {
			if userInput := p.UserInputSnapshot(); userInput == "" && opts.DefaultValue != "" {
				p.SetImmediateValue(opts.DefaultValue)
			}
		}
	})

	return p.Prompt()
}
