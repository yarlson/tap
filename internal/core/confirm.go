package core

func Confirm(opts ConfirmOptions) any {
	// Defaults are kept to mirror TS API, but not yet used in rendering
	if opts.Active == "" {
		opts.Active = "Yes"
	}
	if opts.Inactive == "" {
		opts.Inactive = "No"
	}
	initial := opts.InitialValue
	var lastPressed string

	p := NewPrompt(PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *Prompt) string {
			s := p.snap.Load().(promptState)
			state := s.State

			// If we have a pressed key and we're submitting, show it
			if (state == StateSubmit || state == StateCancel) && lastPressed != "" {
				return opts.Message + " " + lastPressed
			}

			// Otherwise just show the message
			return opts.Message
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
