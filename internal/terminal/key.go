package terminal

// Key represents a parsed keyboard input event
type Key struct {
	Name string // "up", "down", "left", "right", "return", "escape", "backspace", "delete", "space", "tab", or lowercase letter
	Rune rune   // The actual character (0 for special keys)
	Ctrl bool   // True if Ctrl modifier was pressed
}
