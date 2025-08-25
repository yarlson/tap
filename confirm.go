package tap

// Confirm creates a styled confirm prompt
func Confirm(opts ConfirmOptions) bool {
	if opts.Input != nil && opts.Output != nil {
		return confirm(opts)
	}

	return runWithTerminal(func(in Reader, out Writer) bool {
		if opts.Input == nil {
			opts.Input = in
		}
		if opts.Output == nil {
			opts.Output = out
		}

		return confirm(opts)
	})
}

// confirm implements the core confirm prompt logic
func confirm(opts ConfirmOptions) bool {
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

	p := NewPrompt(PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *Prompt) string {
			s := p.StateSnapshot()

			// Create title with symbol and message
			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			// If we're submitting, show simplified version
			if s == StateSubmit {
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
