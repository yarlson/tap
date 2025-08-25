package prompts

import (
	"fmt"
	"strings"
)

// styledSelectState holds the state for a styled select prompt
type styledSelectState[T any] struct {
	cursor  int
	options []SelectOption[T]
}

// Select creates a styled select prompt
func Select[T any](opts SelectOptions[T]) T {
	coreOptions := make([]SelectOption[T], len(opts.Options))
	for i, opt := range opts.Options {
		coreOptions[i] = SelectOption[T]{
			Value: opt.Value,
			Label: opt.Label,
			Hint:  opt.Hint,
		}
	}

	initialCursor := 0
	initialValue := getInitialValue(opts, coreOptions)
	for i, option := range coreOptions {
		if isEqual(option.Value, initialValue) {
			initialCursor = i
			break
		}
	}

	state := &styledSelectState[T]{
		cursor:  initialCursor,
		options: coreOptions,
	}

	styledPrompt := NewPromptWithTracking(PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *Prompt) string {
			return renderStyledSelect(p, opts, state.options, state.cursor)
		},
		InitialValue: initialValue,
	}, false)

	styledPrompt.SetImmediateValue(initialValue)

	styledPrompt.On("cursor", func(direction string) {
		switch direction {
		case "up", "left":
			if state.cursor == 0 {
				state.cursor = len(state.options) - 1
			} else {
				state.cursor--
			}
		case "down", "right":
			if state.cursor == len(state.options)-1 {
				state.cursor = 0
			} else {
				state.cursor++
			}
		}

		newValue := state.options[state.cursor].Value
		styledPrompt.SetImmediateValue(newValue)
	})

	v := styledPrompt.Prompt()
	if t, ok := v.(T); ok {
		return t
	}
	var zero T
	return zero
}

func getInitialValue[T any](opts SelectOptions[T], coreOptions []SelectOption[T]) T {
	if opts.InitialValue != nil {
		return *opts.InitialValue
	}
	if len(coreOptions) > 0 {
		return coreOptions[0].Value
	}
	var zero T
	return zero
}

func isEqual[T any](a, b T) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func renderStyledSelect[T any](p *Prompt, opts SelectOptions[T], coreOptions []SelectOption[T], cursor int) string {
	state := p.StateSnapshot()

	// Build title
	title := fmt.Sprintf("%s\n%s  %s\n", gray(Bar), Symbol(state), opts.Message)

	switch state {
	case StateSubmit:
		selected := coreOptions[cursor]
		label := selected.Label
		if label == "" {
			label = fmt.Sprintf("%v", selected.Value)
		}
		return fmt.Sprintf("%s%s  %s", title, gray(Bar), dim(label))

	default:
		var lines []string
		for i, option := range coreOptions {
			label := option.Label
			if label == "" {
				label = fmt.Sprintf("%v", option.Value)
			}

			if i == cursor {
				line := fmt.Sprintf("%s %s", green(RadioActive), label)
				if option.Hint != "" {
					line += fmt.Sprintf(" %s", dim(fmt.Sprintf("(%s)", option.Hint)))
				}
				lines = append(lines, line)
			} else {
				lines = append(lines, fmt.Sprintf("%s %s", dim(RadioInactive), dim(label)))
			}
		}

		optionsText := strings.Join(lines, fmt.Sprintf("\n%s  ", cyan(Bar)))
		return fmt.Sprintf("%s%s  %s\n%s\n", title, cyan(Bar), optionsText, cyan(BarEnd))
	}
}
