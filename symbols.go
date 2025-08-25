package tap

// Unicode symbols for drawing styled prompts
const (
	// Step symbols
	StepActive = "◆"
	StepCancel = "■"
	StepError  = "▲"
	StepSubmit = "◇"

	// Bar symbols
	Bar           = "│"
	BarH          = "─"
	BarStart      = "┌"
	BarStartRight = "┐"
	BarEnd        = "└"
	BarEndRight   = "┘"

	// Corner symbols (rounded)
	CornerTopLeft     = "╭"
	CornerTopRight    = "╮"
	CornerBottomLeft  = "╰"
	CornerBottomRight = "╯"

	// Radio symbols
	RadioActive   = "●"
	RadioInactive = "○"

	// Checkbox symbols for multiselect
	CheckboxChecked   = "◼"
	CheckboxUnchecked = "◻"
)

// ANSI color codes
const (
	Reset = "\033[0m"

	// Colors
	Gray   = "\033[90m"
	Red    = "\033[91m"
	Green  = "\033[92m"
	Yellow = "\033[93m"
	Cyan   = "\033[96m"

	// Text styles
	Dim           = "\033[2m"
	Inverse       = "\033[7m"
	Strikethrough = "\033[9m"
)

// Color helper functions
func gray(s string) string          { return Gray + s + Reset }
func red(s string) string           { return Red + s + Reset }
func green(s string) string         { return Green + s + Reset }
func yellow(s string) string        { return Yellow + s + Reset }
func cyan(s string) string          { return Cyan + s + Reset }
func dim(s string) string           { return Dim + s + Reset }
func inverse(s string) string       { return Inverse + s + Reset }
func strikethrough(s string) string { return Strikethrough + s + Reset }

// Symbol returns the appropriate symbol for a given state with color
func Symbol(state ClackState) string {
	switch state {
	case StateInitial, StateActive:
		return cyan(StepActive)
	case StateCancel:
		return red(StepCancel)
	case StateError:
		return yellow(StepError)
	case StateSubmit:
		return green(StepSubmit)
	}
	return StepActive
}
