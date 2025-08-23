package core

import "fmt"

// SelectOption represents an option in a select prompt
type SelectOption[T any] struct {
	Value T
	Label string
	Hint  string
}

// SelectOptions holds the configuration for a select prompt
type SelectOptions[T any] struct {
	Message      string
	Options      []SelectOption[T]
	InitialValue *T
	Input        Reader
	Output       Writer
	Validate     func(T) error
}

// SelectPrompt is the core select prompt implementation
type SelectPrompt[T any] struct {
	*Prompt
	options []SelectOption[T]
	cursor  int
}

// NewSelectPrompt creates a new select prompt
func NewSelectPrompt[T any](opts SelectOptions[T]) *SelectPrompt[T] {
	sp := &SelectPrompt[T]{
		options: opts.Options,
		cursor:  0,
	}

	if opts.InitialValue != nil {
		for i, option := range opts.Options {
			if isEqual(*opts.InitialValue, option.Value) {
				sp.cursor = i
				break
			}
		}
	}

	promptOpts := PromptOptions{
		Render: func(p *Prompt) string {
			return sp.render()
		},
		Input:  opts.Input,
		Output: opts.Output,
		Validate: func(value any) error {
			if opts.Validate != nil && value != nil {
				if v, ok := value.(T); ok {
					return opts.Validate(v)
				}
			}
			return nil
		},
		InitialValue: sp.getSelectedValue(),
	}

	sp.Prompt = NewPromptWithTracking(promptOpts, false)

	sp.SetImmediateValue(sp.getSelectedValue())

	sp.On("cursor", func(direction string) {
		sp.handleCursor(direction)
	})

	return sp
}

// isEqual compares two values for equality
func isEqual[T any](a, b T) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (sp *SelectPrompt[T]) getSelectedValue() T {
	if len(sp.options) == 0 {
		var zero T
		return zero
	}
	return sp.options[sp.cursor].Value
}

func (sp *SelectPrompt[T]) handleCursor(direction string) {
	switch direction {
	case "up", "left":
		if sp.cursor == 0 {
			sp.cursor = len(sp.options) - 1
		} else {
			sp.cursor--
		}
	case "down", "right":
		if sp.cursor == len(sp.options)-1 {
			sp.cursor = 0
		} else {
			sp.cursor++
		}
	}
	sp.SetImmediateValue(sp.getSelectedValue())
}

func (sp *SelectPrompt[T]) render() string {
	if len(sp.options) == 0 {
		return "No options available"
	}

	selected := sp.options[sp.cursor]
	label := selected.Label
	if label == "" {
		label = fmt.Sprintf("%v", selected.Value)
	}

	return fmt.Sprintf("Selected: %s", label)
}

// Select creates and runs a select prompt
func Select[T any](opts SelectOptions[T]) T {
	prompt := NewSelectPrompt(opts)
	v := prompt.Prompt.Prompt()
	if t, ok := v.(T); ok {
		return t
	}
	var zero T
	return zero
}
