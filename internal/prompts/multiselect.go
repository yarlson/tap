package prompts

import (
	"fmt"
	"strings"

	"github.com/yarlson/tap/internal/core"
)

type styledMultiSelectState[T any] struct {
	cursor   int
	options  []core.SelectOption[T]
	selected map[int]bool
	order    []int
}

// MultiSelect renders a styled multi-select and returns selected values.
func MultiSelect[T any](opts MultiSelectOptions[T]) []T {
	coreOptions := make([]core.SelectOption[T], len(opts.Options))
	for i, opt := range opts.Options {
		coreOptions[i] = core.SelectOption[T]{Value: opt.Value, Label: opt.Label, Hint: opt.Hint}
	}

	sel := make(map[int]bool)
	order := make([]int, 0, len(coreOptions))
	if len(opts.InitialValues) > 0 {
		for i, o := range coreOptions {
			for _, iv := range opts.InitialValues {
				if isEqual(o.Value, iv) {
					sel[i] = true
					order = append(order, i)
					break
				}
			}
		}
	}

	state := &styledMultiSelectState[T]{
		cursor:   0,
		options:  coreOptions,
		selected: sel,
		order:    order,
	}

	prompt := core.NewPromptWithTracking(core.PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *core.Prompt) string {
			return renderStyledMultiSelect(p, opts, state)
		},
	}, false)

	// Initialize with any preselected items
	{
		var initVals []T
		for i, opt := range state.options {
			if state.selected[i] {
				initVals = append(initVals, opt.Value)
			}
		}
		if len(initVals) > 0 {
			prompt.SetImmediateValue(initVals)
		}
	}

	// Cursor movement
	prompt.On("cursor", func(direction string) {
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
	})

	// Space toggles selection
	prompt.On("key", func(_ string, key core.Key) {
		if key.Name == "space" {
			idx := state.cursor
			if state.selected[idx] {
				delete(state.selected, idx)
				for i, v := range state.order {
					if v == idx {
						state.order = append(state.order[:i], state.order[i+1:]...)
						break
					}
				}
			} else {
				// Enforce MaxItems if specified
				if opts.MaxItems != nil {
					selCount := 0
					for _, v := range state.selected {
						if v {
							selCount++
						}
					}
					if selCount >= *opts.MaxItems {
						// at limit; ignore additional selection
						return
					}
				}
				state.selected[idx] = true
				state.order = append(state.order, idx)
			}
			var cur []T
			for i, opt := range state.options {
				if state.selected[i] {
					cur = append(cur, opt.Value)
				}
			}
			prompt.SetImmediateValue(cur)
		}
	})

	v := prompt.Prompt()
	if t, ok := v.([]T); ok {
		return t
	}
	return nil
}

func renderStyledMultiSelect[T any](p *core.Prompt, opts MultiSelectOptions[T], st *styledMultiSelectState[T]) string {
	state := p.StateSnapshot()
	// Build title with selection count indicator
	count := 0
	for _, v := range st.selected {
		if v {
			count++
		}
	}
	countText := ""
	if opts.MaxItems != nil {
		countText = fmt.Sprintf(" %s", dim(fmt.Sprintf("(%d/%d)", count, *opts.MaxItems)))
	} else if count > 0 {
		countText = fmt.Sprintf(" %s", dim(fmt.Sprintf("(%d)", count)))
	}
	title := fmt.Sprintf("%s\n%s  %s%s\n", gray(Bar), Symbol(state), opts.Message, countText)

	switch state {
	case core.StateSubmit:
		labels := []string{}
		for i, option := range st.options {
			if st.selected[i] {
				label := option.Label
				if label == "" {
					label = fmt.Sprintf("%v", option.Value)
				}
				labels = append(labels, label)
			}
		}
		text := strings.Join(labels, ", ")
		return fmt.Sprintf("%s%s  %s", title, gray(Bar), dim(text))
	default:
		var lines []string
		for i, option := range st.options {
			label := option.Label
			if label == "" {
				label = fmt.Sprintf("%v", option.Value)
			}
			checked := st.selected[i]
			box := CheckboxUnchecked
			if checked {
				box = CheckboxChecked
			}

			text := label
			if !checked {
				text = dim(label)
			}

			if i == st.cursor {
				line := fmt.Sprintf("%s %s", green(box), text)
				if option.Hint != "" {
					line += fmt.Sprintf(" %s", dim(fmt.Sprintf("(%s)", option.Hint)))
				}
				lines = append(lines, line)
			} else {
				if checked {
					line := fmt.Sprintf("%s %s", green(box), text)
					lines = append(lines, line)
				} else {
					line := fmt.Sprintf("%s %s", dim(box), text)
					lines = append(lines, line)
				}
			}
		}
		optionsText := strings.Join(lines, fmt.Sprintf("\n%s  ", cyan(Bar)))
		return fmt.Sprintf("%s%s  %s\n%s\n", title, cyan(Bar), optionsText, cyan(BarEnd))
	}
}
