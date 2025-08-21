package prompts

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

type BoxAlignment string

const (
	BoxAlignLeft   BoxAlignment = "left"
	BoxAlignCenter BoxAlignment = "center"
	BoxAlignRight  BoxAlignment = "right"
)

type BoxOptions struct {
	Output         Writer
	Columns        int          // terminal columns; if 0, default to 80
	WidthFraction  float64      // 0..1 fraction of Columns; ignored if WidthAuto
	WidthAuto      bool         // compute width to content automatically (capped by Columns)
	TitlePadding   int          // spaces padding inside borders around title
	ContentPadding int          // spaces padding inside borders around content lines
	TitleAlign     BoxAlignment // left|center|right
	ContentAlign   BoxAlignment // left|center|right
	Rounded        bool
	IncludePrefix  bool
	FormatBorder   func(string) string // formatter for border glyphs (e.g., color)
}

func defaultBorderFormat(s string) string { return s }

// Common border formatters for examples
func GrayBorder(s string) string { return gray(s) }
func CyanBorder(s string) string { return cyan(s) }

// Box renders a framed message with optional title.
func Box(message string, title string, opts BoxOptions) {
	out := opts.Output
	if out == nil {
		return
	}

	columns := opts.Columns
	if columns <= 0 {
		columns = 80
	}

	formatBorder := opts.FormatBorder
	if formatBorder == nil {
		formatBorder = defaultBorderFormat
	}

	borderWidth := 1
	borderTotal := borderWidth * 2

	titlePadding := opts.TitlePadding
	if titlePadding < 0 {
		titlePadding = 0
	}

	contentPadding := opts.ContentPadding
	if contentPadding < 0 {
		contentPadding = 0
	}

	linePrefix := ""
	if opts.IncludePrefix {
		linePrefix = formatBorder(Bar) + " "
	}

	var symbols [4]string
	if opts.Rounded {
		symbols[0] = formatBorder(CornerTopLeft)
		symbols[1] = formatBorder(CornerTopRight)
		symbols[2] = formatBorder(CornerBottomLeft)
		symbols[3] = formatBorder(CornerBottomRight)
	} else {
		symbols[0] = formatBorder(BarStart)
		symbols[1] = formatBorder(BarStartRight)
		symbols[2] = formatBorder(BarEnd)
		symbols[3] = formatBorder(BarEndRight)
	}

	hSymbol := formatBorder(BarH)
	vSymbol := formatBorder(Bar)

	maxBoxWidth := columns - len(linePrefix)

	// Determine box width
	var boxWidth int
	if opts.WidthAuto {
		// start from fraction if provided else full width
		frac := opts.WidthFraction
		if frac <= 0 {
			frac = 1.0
		}
		boxWidth = int(math.Floor(float64(columns)*frac)) - len(linePrefix)
		if boxWidth <= 0 {
			boxWidth = maxBoxWidth
		}
		// ensure big enough for content once inner width computed; we will shrink if needed
	} else {
		frac := opts.WidthFraction
		if frac <= 0 {
			frac = 1.0
		}
		boxWidth = int(math.Floor(float64(columns)*frac)) - len(linePrefix)
		if boxWidth <= 0 {
			boxWidth = maxBoxWidth
		}
	}

	if boxWidth%2 != 0 {
		if boxWidth < maxBoxWidth {
			boxWidth++
		} else if boxWidth > 1 {
			boxWidth--
		}
	}

	innerWidth := boxWidth - borderTotal
	if innerWidth < 1 {
		innerWidth = 1
		boxWidth = innerWidth + borderTotal
	}

	// Auto width: shrink to content size if possible
	if opts.WidthAuto {
		longest := len(title) + titlePadding*2
		for _, line := range strings.Split(message, "\n") {
			if l := len(line) + contentPadding*2; l > longest {
				longest = l
			}
		}
		want := longest + borderTotal
		if want < boxWidth {
			boxWidth = want
			if boxWidth%2 != 0 {
				boxWidth++
			}
			innerWidth = boxWidth - borderTotal
		}
	}

	// Title alignment and truncation
	maxTitle := innerWidth - titlePadding*2
	truncatedTitle := title
	if maxTitle < 0 {
		maxTitle = 0
	}
	if visibleWidth(truncatedTitle) > maxTitle && maxTitle >= 3 {
		// naive truncate by runes while tracking width
		truncatedTitle = truncateToWidth(title, maxTitle)
	}

	leftTitlePad, rightTitlePad := getPaddingForLine(visibleWidth(truncatedTitle), innerWidth, titlePadding, opts.TitleAlign)

	// Write top border with title
	_, _ = fmt.Fprintf(out, "%s%s%s%s%s%s\n",
		linePrefix,
		symbols[0],
		strings.Repeat(hSymbol, leftTitlePad),
		truncatedTitle,
		strings.Repeat(hSymbol, rightTitlePad),
		symbols[1],
	)

	// Wrap content to inner width - content paddings
	wrapWidth := innerWidth - contentPadding*2
	if wrapWidth < 0 {
		wrapWidth = 0
	}
	wrappedLines := wrapTextHardWidth(message, wrapWidth)

	for _, line := range wrappedLines {
		leftPad, rightPad := getPaddingForLine(visibleWidth(line), innerWidth, contentPadding, opts.ContentAlign)
		_, _ = fmt.Fprintf(out, "%s%s%s%s%s%s\n",
			linePrefix,
			vSymbol,
			strings.Repeat(" ", leftPad),
			line,
			strings.Repeat(" ", rightPad),
			vSymbol,
		)
	}

	// Bottom border
	_, _ = fmt.Fprintf(out, "%s%s%s%s\n",
		linePrefix,
		symbols[2],
		strings.Repeat(hSymbol, innerWidth),
		symbols[3],
	)
}

// getPaddingForLine mirrors the TS logic.
func getPaddingForLine(lineLength int, innerWidth int, padding int, align BoxAlignment) (int, int) {
	left := padding
	var right int
	switch align {
	case BoxAlignCenter:
		left = int(math.Floor(float64(innerWidth-lineLength) / 2.0))
		if left < padding {
			left = padding
		}
	case BoxAlignRight:
		left = innerWidth - lineLength - padding
		if left < padding {
			left = padding
		}
	}
	right = innerWidth - left - lineLength
	if right < 0 {
		right = 0
	}
	if left < 0 {
		left = 0
	}
	return left, right
}

// visibleWidth returns the display cell width, accounting for wide runes and ignoring ANSI.
func visibleWidth(s string) int {
	// strip ANSI
	ansi := regexp.MustCompile("\x1b\\[[0-9;?]*[ -/]*[@-~]")
	clean := ansi.ReplaceAllString(s, "")
	return runewidth.StringWidth(clean)
}

// truncateToWidth trims s to fit width columns and appends "..." if trimmed.
func truncateToWidth(s string, width int) string {
	if visibleWidth(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:0]
	}
	target := width - 3
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > target {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteString("...")
	return b.String()
}

// wrapTextHardWidth performs hard wrapping at the given display width preserving newlines.
func wrapTextHardWidth(s string, width int) []string {
	if width <= 0 {
		parts := strings.Split(s, "\n")
		for i := range parts {
			parts[i] = ""
		}
		return parts
	}
	var result []string
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			result = append(result, "")
			continue
		}
		var b strings.Builder
		w := 0
		for _, r := range line {
			rw := runewidth.RuneWidth(r)
			if w+rw > width {
				result = append(result, b.String())
				b.Reset()
				w = 0
			}
			b.WriteRune(r)
			w += rw
		}
		result = append(result, b.String())
	}
	return result
}
