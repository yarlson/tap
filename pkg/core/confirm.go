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

	p := NewPrompt(PromptOptions{
		Input:  opts.Input,
		Output: opts.Output,
		Render: func(p *Prompt) string { return opts.Message },
	})

	p.On("cursor", func(dir string) {
		if dir == "left" || dir == "right" {
			initial = !initial
			p.SetValue(initial)
		}
	})

	p.SetValue(initial)
	return p.Prompt()
}
