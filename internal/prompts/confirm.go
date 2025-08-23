package prompts

import "github.com/yarlson/tap/internal/core"

// Confirm creates a styled confirm prompt
func Confirm(opts ConfirmOptions) bool {
	active := opts.Active
	if active == "" {
		active = "Yes"
	}
	inactive := opts.Inactive
	if inactive == "" {
		inactive = "No"
	}

	initial := opts.InitialValue
	currentValue := initial

	p := core.NewPrompt(core.PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *core.Prompt) string {
			s := p.StateSnapshot()

			// Create title with symbol and message
			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			// If we're submitting, show simplified version
			if s == core.StateSubmit {
				value := ""
				if val, ok := p.ValueSnapshot().(bool); ok {
					if val {
						value = active
					} else {
						value = inactive
					}
				}
				return title + gray(Bar) + "  " + dim(value)
			}

			var activeOption, inactiveOption string
			if currentValue {
				activeOption = green(RadioActive) + " " + active
				inactiveOption = dim(RadioInactive) + " " + dim(inactive)
			} else {
				activeOption = dim(RadioInactive) + " " + dim(active)
				inactiveOption = green(RadioActive) + " " + inactive
			}

			return title + cyan(Bar) + "  " + activeOption + " " + dim("/") + " " + inactiveOption + "\n" + cyan(BarEnd)
		},
	})

	p.On("cursor", func(dir string) {
		if dir == "left" || dir == "right" {
			currentValue = !currentValue
			p.SetValue(currentValue)
		}
	})

	p.On("confirm", func(val bool) {})

	p.SetValue(currentValue)
	v := p.Prompt()
	if b, ok := v.(bool); ok {
		return b
	}

	return false
}
