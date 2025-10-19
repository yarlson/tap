package tap

import (
	"context"
	"fmt"
	"strings"
)

// Autocomplete renders a text prompt with inline suggestions.
func Autocomplete(ctx context.Context, opts AutocompleteOptions) string {
	if opts.Input != nil && opts.Output != nil {
		return autocomplete(ctx, opts)
	}

	return runWithTerminal(func(in Reader, out Writer) string {
		if opts.Input == nil {
			opts.Input = in
		}

		if opts.Output == nil {
			opts.Output = out
		}

		return autocomplete(ctx, opts)
	})
}

type acState struct {
	suggestions []string
	selected    int
	accepted    string // value accepted via Tab
}

func (st *acState) clampSelected() {
	if len(st.suggestions) == 0 {
		st.selected = 0
		return
	}

	if st.selected < 0 {
		st.selected = len(st.suggestions) - 1
	} else if st.selected >= len(st.suggestions) {
		st.selected = 0
	}
}

func autocomplete(ctx context.Context, opts AutocompleteOptions) string {
	// Wrap validator to match PromptOptions
	var validate func(any) error
	if opts.Validate != nil {
		validate = func(v any) error {
			s, _ := v.(string)
			return opts.Validate(s)
		}
	}

	max := opts.MaxResults
	if max <= 0 {
		max = 5
	}

	state := &acState{selected: 0}

	// Helper: compute suggestions respecting max
	getSugs := func(input string) []string {
		if opts.Suggest == nil {
			return nil
		}

		list := opts.Suggest(input)
		if len(list) > max {
			return append([]string{}, list[:max]...)
		}

		return append([]string{}, list...)
	}

	// Local input state
	var (
		inBuf []rune
		cur   int
	)

	p := NewPromptWithTracking(PromptOptions{
		Input:        opts.Input,
		Output:       opts.Output,
		Validate:     validate,
		InitialValue: opts.DefaultValue,
		Render: func(p *Prompt) string {
			s := p.StateSnapshot()

			// Title
			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			// Display input using local state
			var displayInput string

			if state.accepted == "" && len(inBuf) == 0 && opts.Placeholder != "" {
				r := []rune(opts.Placeholder)
				if len(r) > 0 {
					displayInput = inverse(string(r[0])) + dim(string(r[1:]))
				} else {
					displayInput = inverse(" ")
				}
			} else {
				displayInput = renderTextWithCursor(string(inBuf), cur, s)
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
				if strings.TrimSpace(value) != "" {
					valueText = "  " + dim(value)
				}
				// Add a prefixed newline (gray bar) after submit to visually separate
				// from subsequent messages, matching expectations in examples.
				return title + gray(Bar) + valueText + "\n" + gray(Bar)
			case StateCancel:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				valueText := ""
				if strings.TrimSpace(value) != "" {
					valueText = "  " + strikethrough(dim(value))
				}

				result := title + gray(Bar) + valueText
				if strings.TrimSpace(value) != "" {
					result += "\n" + gray(Bar)
				}

				return result
			default:
				if len(state.suggestions) == 0 {
					return title + cyan(Bar) + "  " + displayInput + "\n" + cyan(BarEnd)
				}

				var lines []string

				for i, sg := range state.suggestions {
					if i == state.selected {
						lines = append(lines, fmt.Sprintf("%s %s", green(RadioActive), sg))
					} else {
						lines = append(lines, fmt.Sprintf("%s %s", dim(RadioInactive), dim(sg)))
					}
				}

				sugs := strings.Join(lines, fmt.Sprintf("\n%s  ", cyan(Bar)))

				return fmt.Sprintf("%s%s  %s\n%s  %s\n%s\n", title, cyan(Bar), displayInput, cyan(Bar), sugs, cyan(BarEnd))
			}
		},
	}, false)

	// Initialize from InitialValue if provided
	if opts.InitialValue != "" {
		inBuf = []rune(opts.InitialValue)
		cur = len(inBuf)
		p.SetImmediateValue(string(inBuf))

		state.suggestions = getSugs(string(inBuf))
		if state.selected >= len(state.suggestions) {
			state.selected = 0
		}
	}

	// Key handling: build input, manage cursor, suggestions, and accept
	p.On("key", func(char string, key Key) {
		switch key.Name {
		case "left":
			if cur > 0 {
				cur--
			}
		case "right":
			if cur < len(inBuf) {
				cur++
			}
		case "backspace":
			if cur > 0 {
				inBuf = append(inBuf[:cur-1], inBuf[cur:]...)
				cur--
			}
		case "delete":
			if cur < len(inBuf) {
				inBuf = append(inBuf[:cur], inBuf[cur+1:]...)
			}
		case "up":
			if len(state.suggestions) > 0 {
				state.selected--
				state.clampSelected()
			}
		case "down":
			if len(state.suggestions) > 0 {
				state.selected++
				state.clampSelected()
			}
		case "tab":
			if len(state.suggestions) > 0 {
				accepted := state.suggestions[state.selected]
				inBuf = []rune(accepted)
				cur = len(inBuf)
				p.SetImmediateValue(string(inBuf))
				p.SetValue(accepted)
			}
		default:
			// Printable characters (including space) via char
			if char != "" {
				for _, r := range char {
					if r >= 32 && r <= 126 { // basic printable
						inBuf = append(inBuf[:cur], append([]rune{r}, inBuf[cur:]...)...)
						cur++
					}
				}
			}
		}

		// After any edit/update, reflect in value and recompute suggestions
		p.SetImmediateValue(string(inBuf))

		state.suggestions = getSugs(string(inBuf))
		if state.selected >= len(state.suggestions) {
			state.selected = 0
		}

		// If this key is return, prime the value so Prompt will submit it
		if key.Name == "return" {
			p.SetValue(string(inBuf))
		}
	})

	v := p.Prompt(ctx)
	if s, ok := v.(string); ok {
		return s
	}

	return ""
}
