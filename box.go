package tap

import (
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
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

func GrayBorder(s string) string   { return gray(s) }
func CyanBorder(s string) string   { return cyan(s) }
func GreenBorder(s string) string  { return green(s) }
func YellowBorder(s string) string { return yellow(s) }
func RedBorder(s string) string    { return red(s) }

// Box renders a framed message with optional title.
func Box(message string, title string, opts BoxOptions) {
	out := opts.Output
	if out == nil {
		out, _ = resolveWriter()
	}

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

	titlePadding := max(opts.TitlePadding, 0)
	contentPadding := max(opts.ContentPadding, 0)

	linePrefix := ""
	if opts.IncludePrefix {
		linePrefix = gray(Bar) + " "
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

	maxBoxWidth := columns - visibleWidth(linePrefix)

	// Determine box width
	var boxWidth int

	if opts.WidthAuto {
		// start from fraction if provided else full width
		frac := opts.WidthFraction
		if frac <= 0 {
			frac = 1.0
		}

		boxWidth = int(math.Floor(float64(columns)*frac)) - visibleWidth(linePrefix)
		if boxWidth <= 0 {
			boxWidth = maxBoxWidth
		}
	} else {
		frac := opts.WidthFraction
		if frac <= 0 {
			frac = 1.0
		}

		boxWidth = int(math.Floor(float64(columns)*frac)) - visibleWidth(linePrefix)
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
		longest := visibleWidth(title) + titlePadding*2
		for _, line := range strings.Split(message, "\n") {
			if l := visibleWidth(line) + contentPadding*2; l > longest {
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

// truncateToWidth trims s to fit width columns and appends "..." if trimmed.
func truncateToWidth(s string, width int) string {
	if visibleWidth(s) <= width {
		return s
	}

	if width <= 3 {
		return ""
	}

	target := width - 3

	var b strings.Builder

	w := 0
	sawANSI := false
	sawReset := false

	for i := 0; i < len(s); {
		token, tw, next := scanANSIToken(s, i)
		i = next

		if tw == 0 {
			if len(token) > 0 && token[0] == '\x1b' {
				sawANSI = true
				if token == Reset {
					sawReset = true
				}
			}

			b.WriteString(token)

			continue
		}

		if w+tw > target {
			break
		}

		b.WriteString(token)

		w += tw
	}

	b.WriteString("...")

	if sawANSI && !sawReset {
		b.WriteString(Reset)
	}

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

	type segment struct {
		text       string
		width      int
		breakAfter bool
	}

	build := func(segs []segment) string {
		if len(segs) == 0 {
			return ""
		}

		var b strings.Builder
		for _, seg := range segs {
			b.WriteString(seg.text)
		}

		return b.String()
	}

	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			result = append(result, "")
			continue
		}

		var tokens []segment

		currentWidth := 0
		lastBreak := -1

		recalc := func() {
			currentWidth = 0
			lastBreak = -1

			for i, seg := range tokens {
				currentWidth += seg.width
				if seg.breakAfter {
					lastBreak = i
				}
			}
		}

		appendToken := func(seg segment) {
			tokens = append(tokens, seg)

			currentWidth += seg.width
			if seg.breakAfter {
				lastBreak = len(tokens) - 1
			}
		}

		dropTokens := func(n int) {
			if n <= 0 {
				return
			}

			tokens = append([]segment{}, tokens[n:]...)

			recalc()
		}

		flush := func(end int) {
			if end < 0 {
				return
			}

			result = append(result, build(tokens[:end]))
		}

		for i := 0; i < len(line); {
			token, tw, next := scanANSIToken(line, i)
			i = next

			breakAfter := false

			if tw > 0 {
				r, _ := utf8.DecodeRuneInString(token)
				if unicode.IsSpace(r) {
					breakAfter = true
				}
			}

			appendToken(segment{token, tw, breakAfter})

			for width > 0 && currentWidth > width {
				if lastBreak >= 0 {
					flush(lastBreak)
					dropTokens(lastBreak + 1)

					continue
				}

				hardIdx := len(tokens) - 1
				if hardIdx <= 0 {
					flush(len(tokens))
					dropTokens(len(tokens))
				} else {
					flush(hardIdx)
					dropTokens(hardIdx)
				}
			}
		}

		if len(tokens) > 0 {
			result = append(result, build(tokens))
		}
	}

	return result
}
