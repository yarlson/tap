package tap

import (
	"regexp"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// Strip ANSI sequences for width calculations.
var ansiRegexp = regexp.MustCompile("\x1b\\[[0-9;?]*[\x20-\x2f]*[@-~]")

// scanANSIToken returns the next ANSI-aware token from s starting at idx.
// The returned token preserves escape sequences, while width reports the
// printable width contributed by the token (zero for control sequences).
func scanANSIToken(s string, idx int) (token string, width, next int) {
	if idx >= len(s) {
		return "", 0, len(s)
	}

	if s[idx] == '\x1b' {
		rel := ansiRegexp.FindStringIndex(s[idx:])
		if len(rel) > 0 && rel[0] == 0 {
			return s[idx : idx+rel[1]], 0, idx + rel[1]
		}

		// OSC sequences: ESC ] ... (ST or BEL terminator)
		if idx+1 < len(s) && s[idx+1] == ']' {
			j := idx + 2
			for j < len(s) {
				if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
					return s[idx : j+2], 0, j + 2
				}

				if s[j] == '\a' { // BEL terminator
					return s[idx : j+1], 0, j + 1
				}

				j++
			}

			return s[idx:], 0, len(s)
		}

		// Unrecognized escape, treat as zero-width char
		return s[idx : idx+1], 0, idx + 1
	}

	r, size := utf8.DecodeRuneInString(s[idx:])
	if r == utf8.RuneError && size == 1 {
		return s[idx : idx+1], 0, idx + 1
	}

	return s[idx : idx+size], runewidth.RuneWidth(r), idx + size
}
