package core

import (
	"errors"
	"strings"
)

func Text(opts TextOptions) any {
	userInput := opts.InitialValue
	cursor := len([]rune(userInput))

	var validate func(any) error
	if opts.Validate != nil {
		validate = func(v any) error {
			str, _ := v.(string)
			return opts.Validate(str)
		}
	}

	p := NewPrompt(PromptOptions{
		Input:    opts.Input,
		Output:   opts.Output,
		Validate: validate,
		Render: func(p *Prompt) string {
			msg := opts.Message
			sep := ": "
			trimmed := strings.TrimRight(msg, " ")
			if strings.HasSuffix(trimmed, ":") {
				sep = " "
			}
			return msg + sep + userInput
		},
	})

	// Helper to set current value into prompt
	set := func() { p.SetImmediateValue(userInput) }

	p.On("key", func(char string, key Key) {
		// Clear error state on any key except return/cancel (base will also check)
		if key.Name == "return" || key.Name == "c" && key.Ctrl {
			return
		}
		// Editing keys
		switch key.Name {
		case "left":
			if cursor > 0 {
				cursor--
			}
		case "right":
			if cursor < len([]rune(userInput)) {
				cursor++
			}
		case "backspace":
			r := []rune(userInput)
			if cursor > 0 && len(r) > 0 {
				r = append(r[:cursor-1], r[cursor:]...)
				cursor--
				userInput = string(r)
				set()
			}
		default:
			if char != "" && len([]rune(char)) == 1 {
				r := []rune(userInput)
				c := []rune(char)[0]
				r = append(r[:cursor], append([]rune{c}, r[cursor:]...)...)
				cursor++
				userInput = string(r)
				set()
			}
		}
	})

	// Apply default just before submit
	p.On("key", func(_ string, key Key) {
		if key.Name == "return" {
			if userInput == "" {
				p.SetImmediateValue(opts.DefaultValue)
			} else {
				set()
			}
		}
	})

	// For safety: ensure value mirrors initial state
	set()
	res := p.Prompt()
	// If validate was provided and user attempted invalid submit, base prompt
	// will remain in error until another key; behavior tested in prompt tests.
	// Here we simply return what prompt finalized with.
	_, ok := res.(error)
	if ok {
		// Should not happen because prompt wraps errors; keep parity
		return errors.New("invalid")
	}
	return res
}
