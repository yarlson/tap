package prompts

import "github.com/yarlson/glack/pkg/core"

// Confirm creates a styled confirm prompt
func Confirm(opts ConfirmOptions) any {
	active := opts.Active
	if active == "" {
		active = "Yes"
	}
	inactive := opts.Inactive
	if inactive == "" {
		inactive = "No"
	}

	initial := opts.InitialValue
	var lastPressed string

	p := core.NewPrompt(core.PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *core.Prompt) string {
			s := p.StateSnapshot()

			// Create title with symbol and message
			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			// If we have a pressed key and we're submitting, show simplified version
			if (s == core.StateSubmit || s == core.StateCancel) && lastPressed != "" {
				value := ""
				if val, ok := p.ValueSnapshot().(bool); ok {
					if val {
						value = active
					} else {
						value = inactive
					}
				}

				switch s {
				case core.StateSubmit:
					return title + gray(Bar) + "  " + dim(value)
				case core.StateCancel:
					return title + gray(Bar) + "  " + strikethrough(dim(value)) + "\n" + gray(Bar)
				}
			}

			currentValue := initial

			var activeOption, inactiveOption string
			if currentValue {
				activeOption = green(RadioActive) + " " + active
				inactiveOption = dim(RadioInactive) + " " + dim(inactive)
			} else {
				activeOption = dim(RadioInactive) + " " + dim(active)
				inactiveOption = green(RadioActive) + " " + inactive
			}

			return title + cyan(Bar) + "  " + activeOption + " " + dim("/") + " " + inactiveOption + "\n" + cyan(BarEnd) + "\n"
		},
	})

	p.On("cursor", func(dir string) {
		if dir == "left" || dir == "right" {
			initial = !initial
			p.SetValue(initial)
		}
	})

	p.On("confirm", func(val bool) {
		if val {
			lastPressed = "y"
		} else {
			lastPressed = "n"
		}
	})

	p.SetValue(initial)
	return p.Prompt()
}
