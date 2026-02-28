package tap

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Bracketed paste mode escape sequences.
const (
	bracketedPasteEnable  = "\x1b[?2004h"
	bracketedPasteDisable = "\x1b[?2004l"
)

// PUA rune helpers for paste placeholder encoding.
// Paste placeholders are stored as Private Use Area runes (U+E000+) in the buffer.

func isPUA(r rune) bool   { return r >= 0xE000 && r <= 0xF8FF }
func puaToID(r rune) int  { return int(r-0xE000) + 1 }
func idToPUA(id int) rune { return rune(0xE000 + id - 1) }

// resolve expands all PUA runes in the buffer with their stored paste content.
func resolve(buf []rune, pastes map[int]string) string {
	var b strings.Builder
	for _, r := range buf {
		if isPUA(r) {
			b.WriteString(pastes[puaToID(r)])
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// Textarea creates a styled multiline text input prompt.
func Textarea(ctx context.Context, opts TextareaOptions) string {
	if opts.Input != nil && opts.Output != nil {
		return textarea(ctx, opts)
	}

	return runWithTerminal(func(in Reader, out Writer) string {
		if opts.Input == nil {
			opts.Input = in
		}

		if opts.Output == nil {
			opts.Output = out
		}

		return textarea(ctx, opts)
	})
}

func textarea(ctx context.Context, opts TextareaOptions) string {
	// Local buffer state (track=false, same pattern as Autocomplete)
	var (
		buf          []rune
		cur          int
		pasteCounter int
		pasteBuffers = make(map[int]string)
	)

	// Enable bracketed paste mode
	if opts.Output != nil {
		_, _ = opts.Output.Write([]byte(bracketedPasteEnable))
	}

	p := NewPromptWithTracking(PromptOptions{
		Input:        opts.Input,
		Output:       opts.Output,
		InitialValue: opts.DefaultValue,
		Render: func(p *Prompt) string {
			s := p.StateSnapshot()

			title := gray(Bar) + "\n" + Symbol(s) + "  " + opts.Message + "\n"

			switch s {
			case StateSubmit:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				if value == "" {
					return title + gray(Bar)
				}

				lines := strings.Split(value, "\n")
				var parts []string

				for _, line := range lines {
					parts = append(parts, gray(Bar)+"  "+dim(line))
				}

				return title + strings.Join(parts, "\n")

			case StateCancel:
				value := ""
				if val, ok := p.ValueSnapshot().(string); ok {
					value = val
				}

				if value == "" {
					return title + gray(Bar)
				}

				lines := strings.Split(value, "\n")
				var parts []string

				for _, line := range lines {
					parts = append(parts, gray(Bar)+"  "+strikethrough(dim(line)))
				}

				return title + strings.Join(parts, "\n")

			default:
				// Active/Initial/Error state
				barColor := cyan

				if s == StateError {
					barColor = yellow
				}

				if len(buf) == 0 && opts.Placeholder != "" {
					placeholder := renderTextareaPlaceholder(opts.Placeholder)
					result := title + barColor(Bar) + "  " + placeholder + "\n" + barColor(BarEnd)

					if s == StateError {
						errMsg := p.ErrorSnapshot()
						result = title + barColor(Bar) + "  " + placeholder + "\n" + barColor(BarEnd) + "  " + yellow(errMsg)
					}

					return result
				}

				// Render buffer with PUA runes replaced by dim placeholders
				content := renderBufWithPlaceholders(buf)
				lines := strings.Split(content, "\n")
				var parts []string

				for i, line := range lines {
					display := renderTextareaLine(line, buf, cur, i, s)
					parts = append(parts, barColor(Bar)+"  "+display)
				}

				result := title + strings.Join(parts, "\n") + "\n" + barColor(BarEnd)

				if s == StateError {
					errMsg := p.ErrorSnapshot()
					result = title + strings.Join(parts, "\n") + "\n" + barColor(BarEnd) + "  " + yellow(errMsg)
				}

				return result
			}
		},
	}, false)

	// Disable bracketed paste mode on finalize
	p.On("finalize", func() {
		if opts.Output != nil {
			_, _ = opts.Output.Write([]byte(bracketedPasteDisable))
		}
	})

	// Initialize from InitialValue if provided
	if opts.InitialValue != "" {
		buf = []rune(opts.InitialValue)
		cur = len(buf)
		p.SetImmediateValue(string(buf))
	}

	// Key handling
	p.On("key", func(_ string, key Key) {
		switch {
		case key.Name == "paste":
			pasteCounter++
			pasteBuffers[pasteCounter] = key.Content
			buf = slices.Insert(buf, cur, idToPUA(pasteCounter))
			cur++

		case key.Name == "return" && key.Shift:
			// Shift+Enter: insert newline
			buf = slices.Insert(buf, cur, '\n')
			cur++

		case key.Name == "return":
			val := resolve(buf, pasteBuffers)
			if val == "" && opts.DefaultValue != "" {
				val = opts.DefaultValue
			}

			// Validate the resolved string
			if opts.Validate != nil {
				if err := opts.Validate(val); err != nil {
					// Set error state, keep the buffer for continued editing
					errMsg := err.Error()
					e := &ValidationError{}
					if errors.As(err, &e) {
						errMsg = e.Message
					}

					p.cur.Value = val
					p.cur.Error = errMsg
					p.cur.State = StateError
					p.cur.PrevFrame = "" // Force re-render
					p.SetImmediateValue(val)
					return
				}
			}

			p.cur.Value = val
			p.cur.State = StateSubmit

			return

		case key.Name == "left":
			if cur > 0 {
				cur--
				// If we landed right after a PUA rune, skip it
				if cur > 0 && isPUA(buf[cur-1]) {
					cur--
				}
			}

		case key.Name == "right":
			if cur < len(buf) {
				cur++
				// If we just moved onto a PUA rune, skip it
				if cur < len(buf) && isPUA(buf[cur]) {
					cur++
				}
			}

		case key.Name == "up":
			line, col := cursorToLineCol(buf, cur)
			if line > 0 {
				cur = lineColToCursor(buf, line-1, col)
			}

		case key.Name == "down":
			line, col := cursorToLineCol(buf, cur)
			lineCount := countBufferLines(buf)
			if line < lineCount-1 {
				cur = lineColToCursor(buf, line+1, col)
			}

		case key.Name == "home":
			line, _ := cursorToLineCol(buf, cur)
			cur = lineColToCursor(buf, line, 0)

		case key.Name == "end":
			line, _ := cursorToLineCol(buf, cur)
			cur = lineColToCursor(buf, line, len(buf))

		case key.Name == "backspace":
			if cur > 0 {
				target := cur - 1
				if isPUA(buf[target]) {
					delete(pasteBuffers, puaToID(buf[target]))
				}

				buf = slices.Delete(buf, target, target+1)
				cur = target
			}

		case key.Name == "delete":
			if cur < len(buf) {
				if isPUA(buf[cur]) {
					delete(pasteBuffers, puaToID(buf[cur]))
				}

				buf = slices.Delete(buf, cur, cur+1)
			}

		default:
			if key.Rune >= 32 && key.Rune <= 126 {
				buf = slices.Insert(buf, cur, key.Rune)
				cur++
			}
		}

		p.SetImmediateValue(resolve(buf, pasteBuffers))
	})

	v := p.Prompt(ctx)
	if s, ok := v.(string); ok {
		return s
	}

	return ""
}

// renderBufWithPlaceholders converts a buffer to a display string,
// replacing PUA runes with dim "[Text N]" placeholders.
func renderBufWithPlaceholders(buf []rune) string {
	var b strings.Builder
	for _, r := range buf {
		if isPUA(r) {
			b.WriteString(dim(fmt.Sprintf("[Text %d]", puaToID(r))))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// renderTextareaPlaceholder renders placeholder text with inverse first char + dim rest.
func renderTextareaPlaceholder(placeholder string) string {
	runes := []rune(placeholder)
	if len(runes) == 0 {
		return inverse(" ")
	}

	return inverse(string(runes[0])) + dim(string(runes[1:]))
}

// renderTextareaLine renders a single line of textarea content with cursor display.
// The line parameter contains rendered text (with ANSI for PUA placeholders).
// Cursor positioning uses the raw buf to map cursor index to display position.
func renderTextareaLine(line string, buf []rune, cursor, lineIdx int, state ClackState) string {
	if state != StateActive && state != StateInitial {
		return line
	}

	// Calculate cursor position within this line in the raw buffer
	lineStart := 0
	for i := 0; i < lineIdx; i++ {
		idx := indexRune(buf[lineStart:], '\n')
		if idx < 0 {
			break
		}

		lineStart += idx + 1
	}

	// Find raw line end
	lineEnd := lineStart
	for lineEnd < len(buf) && buf[lineEnd] != '\n' {
		lineEnd++
	}

	if cursor < lineStart || cursor > lineEnd {
		return line
	}

	// Cursor is on this line. Map raw buffer position to display position.
	// Walk the raw buffer from lineStart to cursor, tracking display offset
	// in the rendered string (which has ANSI sequences for PUA runes).
	posInLine := cursor - lineStart

	// Count how many runes (non-PUA) and PUA placeholders are before cursor position
	// We need to map raw cursor offset to display string offset
	displayOffset := 0
	for i := lineStart; i < lineStart+posInLine && i < len(buf); i++ {
		if isPUA(buf[i]) {
			// PUA rune renders as dim("[Text N]") which includes ANSI codes
			placeholder := dim(fmt.Sprintf("[Text %d]", puaToID(buf[i])))
			displayOffset += len(placeholder)
		} else {
			displayOffset += len(string(buf[i]))
		}
	}

	// Now split the rendered line at displayOffset (byte offset)
	if displayOffset >= len(line) {
		return line + inverse(" ")
	}

	before := line[:displayOffset]
	after := line[displayOffset:]

	// Extract the first rune/token at cursor position for inverse rendering
	// If cursor is on a PUA rune, the placeholder is already styled, just add cursor after
	if posInLine < lineEnd-lineStart && isPUA(buf[lineStart+posInLine]) {
		// Cursor is at a PUA rune position — render cursor after the placeholder
		placeholder := dim(fmt.Sprintf("[Text %d]", puaToID(buf[lineStart+posInLine])))
		rest := line[displayOffset+len(placeholder):]

		return before + placeholder + inverse(" ") + rest
	}

	// Regular rune at cursor — apply inverse to the first rune
	firstRune := []rune(after)
	if len(firstRune) == 0 {
		return line + inverse(" ")
	}

	charStr := string(firstRune[0])
	remaining := after[len(charStr):]

	return before + inverse(charStr) + remaining
}

// indexRune returns the index of the first occurrence of r in s, or -1.
func indexRune(s []rune, r rune) int {
	for i, v := range s {
		if v == r {
			return i
		}
	}

	return -1
}

// cursorToLineCol converts a flat cursor index into line and column numbers.
func cursorToLineCol(buf []rune, cursor int) (line, col int) {
	for i := 0; i < cursor && i < len(buf); i++ {
		if buf[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	return line, col
}

// lineColToCursor converts a line and column back to a flat cursor index.
// If col exceeds the line's length, cursor is clamped to end of line.
func lineColToCursor(buf []rune, targetLine, targetCol int) int {
	line := 0
	lineStart := 0

	for i, r := range buf {
		if line == targetLine {
			lineStart = i
			break
		}

		if r == '\n' {
			line++
			lineStart = i + 1
		}
	}

	if line < targetLine {
		return len(buf)
	}

	pos := lineStart
	col := 0

	for pos < len(buf) && buf[pos] != '\n' && col < targetCol {
		pos++
		col++
	}

	return pos
}

// countBufferLines returns the number of lines in the buffer (1-based count).
func countBufferLines(buf []rune) int {
	lines := 1
	for _, r := range buf {
		if r == '\n' {
			lines++
		}
	}

	return lines
}
