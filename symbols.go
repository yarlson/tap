package tap

// Unicode symbols for drawing styled prompts.
const (
	// Step symbols.
	StepActive = "◆"
	StepCancel = "■"
	StepError  = "▲"
	StepSubmit = "◇"

	// Bar symbols.
	Bar           = "│"
	BarH          = "─"
	BarStart      = "┌"
	BarStartRight = "┐"
	BarEnd        = "└"
	BarEndRight   = "┘"

	// Corner symbols (rounded).
	CornerTopLeft     = "╭"
	CornerTopRight    = "╮"
	CornerBottomLeft  = "╰"
	CornerBottomRight = "╯"

	// Radio symbols.
	RadioActive   = "●"
	RadioInactive = "○"

	// Checkbox symbols for multiselect.
	CheckboxChecked   = "◼"
	CheckboxUnchecked = "◻"
)

// ANSI color codes.
const (
	Reset = "\033[0m"

	// Colors.
	Gray   = "\033[90m"
	Red    = "\033[91m"
	Green  = "\033[92m"
	Yellow = "\033[93m"
	Cyan   = "\033[96m"

	// Text styles.
	Dim           = "\033[2m"
	Bold          = "\033[1m"
	Inverse       = "\033[7m"
	Strikethrough = "\033[9m"
)

// Color helper functions.
func gray(s string) string          { return Gray + s + Reset }
func red(s string) string           { return Red + s + Reset }
func green(s string) string         { return Green + s + Reset }
func yellow(s string) string        { return Yellow + s + Reset }
func cyan(s string) string          { return Cyan + s + Reset }
func dim(s string) string           { return Dim + s + Reset }
func bold(s string) string          { return Bold + s + Reset }
func inverse(s string) string       { return Inverse + s + Reset }
func strikethrough(s string) string { return Strikethrough + s + Reset }

// Symbol returns the appropriate symbol for a given state with color.
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

// Table symbols.
const (
	// Table border symbols.
	TableTopLeft     = "┌"
	TableTopRight    = "┐"
	TableBottomLeft  = "└"
	TableBottomRight = "┘"
	TableTopTee      = "┬"
	TableBottomTee   = "┴"
	TableLeftTee     = "├"
	TableRightTee    = "┤"
	TableCross       = "┼"
	TableHorizontal  = "─"
	TableVertical    = "│"
)

// Table styling helper functions.
func tableStyle(text string, style TableStyle, color TableColor) string {
	result := text

	// Apply color first
	switch color {
	case TableColorGray:
		result = gray(result)
	case TableColorRed:
		result = red(result)
	case TableColorGreen:
		result = green(result)
	case TableColorYellow:
		result = yellow(result)
	case TableColorCyan:
		result = cyan(result)
	}

	// Apply style - but ensure we reset properly
	switch style {
	case TableStyleBold:
		// Remove any existing reset codes and apply bold with proper reset
		result = "\033[1m" + result + "\033[0m"
	case TableStyleDim:
		// Remove any existing reset codes and apply dim with proper reset
		result = "\033[2m" + result + "\033[0m"
	}

	return result
}
