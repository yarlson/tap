package terminal

// ANSI escape sequences for terminal control
const (
	CursorHide = "\x1b[?25l"
	CursorShow = "\x1b[?25h"
	ClearLine  = "\r\x1b[K"
	CursorUp   = "\x1b[A"
	EraseDown  = "\x1b[J"
	SaveCursor = "\x1b[s"
	RestCursor = "\x1b[u"
)

// MoveUp returns ANSI sequence to move cursor up n lines
func MoveUp(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += CursorUp
	}
	return result
}
