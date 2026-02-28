package tap

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// helper to strip ANSI codes.
func removeANSI(s string) string {
	ansi := regexp.MustCompile("\x1b\\[[0-9;?]*[\x20-\x2f]*[@-~]")
	return ansi.ReplaceAllString(s, "")
}

func TestBox_SquareBasic(t *testing.T) {
	out := NewMockWritable()

	Box("Hello world", "TITLE", BoxOptions{
		Output:        out,
		Columns:       40,
		WidthFraction: 1.0,
		Rounded:       false,
	})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	text := strings.Join(frames, "")

	// Top and bottom borders should use square symbols
	assert.Contains(t, text, BarStart)
	assert.Contains(t, text, BarStartRight)
	assert.Contains(t, text, BarEnd)
	assert.Contains(t, text, BarEndRight)
}

func TestBox_Rounded(t *testing.T) {
	out := NewMockWritable()

	Box("Rounded", "TITLE", BoxOptions{
		Output:        out,
		Columns:       30,
		WidthFraction: 1.0,
		Rounded:       true,
	})

	text := strings.Join(out.GetFrames(), "")
	assert.Contains(t, text, CornerTopLeft)
	assert.Contains(t, text, CornerTopRight)
	assert.Contains(t, text, CornerBottomLeft)
	assert.Contains(t, text, CornerBottomRight)
}

func TestBox_IncludePrefix(t *testing.T) {
	out := NewMockWritable()

	Box("Prefixed", "T", BoxOptions{
		Output:        out,
		Columns:       20,
		WidthFraction: 1.0,
		IncludePrefix: true,
	})

	lines := strings.Split(strings.Join(out.GetFrames(), ""), "\n")
	// First non-empty line should start with prefix "│ " (gray or plain)
	var first string

	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			first = l
			break
		}
	}

	if first == "" && len(lines) > 0 {
		first = lines[0]
	}
	// Strip potential ANSI to compare raw glyph
	raw := removeANSI(first)
	assert.True(t, strings.HasPrefix(raw, Bar+" "))
}

func TestBox_AutoWidth_WrapsContent(t *testing.T) {
	out := NewMockWritable()

	long := "This is a very long line that should wrap around the inner width"
	Box(long, "T", BoxOptions{
		Output:         out,
		Columns:        24,
		WidthAuto:      true,
		ContentPadding: 1,
		TitlePadding:   1,
	})

	lines := strings.Split(strings.Join(out.GetFrames(), ""), "\n")
	// Expect more than three lines: top border, at least two content lines, bottom border
	nonEmpty := 0

	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}

	assert.GreaterOrEqual(t, nonEmpty, 4)
}

func TestBox_FormatBorder_Applied(t *testing.T) {
	out := NewMockWritable()

	Box("X", "T", BoxOptions{
		Output:        out,
		Columns:       20,
		WidthFraction: 1.0,
		FormatBorder:  gray,
	})

	text := strings.Join(out.GetFrames(), "")
	// Should contain gray-colored border character
	assert.Contains(t, text, gray(BarStart))
}

func TestWrapTextHardWidth_IgnoresANSISequences(t *testing.T) {
	colored := green("ColoredContent")
	lines := wrapTextHardWidth(colored, 6)
	assert.Greater(t, len(lines), 1)

	for _, line := range lines {
		assert.LessOrEqual(t, visibleWidth(line), 6)
	}
	// First line should retain color prefix
	assert.Contains(t, lines[0], Green)
}

func TestWrapTextHardWidth_PrefersWhitespaceBreaks(t *testing.T) {
	lines := wrapTextHardWidth("Alpha Beta Gamma", 8)
	assert.Equal(t, []string{"Alpha", "Beta", "Gamma"}, lines)
}

func TestTruncateToWidth_PreservesColorSafety(t *testing.T) {
	colored := green("ColorfulText")
	truncated := truncateToWidth(colored, 8)

	assert.Equal(t, 8, visibleWidth(truncated))
	assert.Contains(t, truncated, Green)
	assert.True(t, strings.HasSuffix(truncated, Reset))
}
