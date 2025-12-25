package tap

import (
	"fmt"
	"math"
	"strings"
)

// Table renders a formatted table with headers and rows
func Table(headers []string, rows [][]string, opts TableOptions) {
	out := opts.Output
	if out == nil {
		out, _ = resolveWriter()
	}

	if out == nil {
		return
	}

	if len(headers) == 0 {
		return
	}

	// Set defaults
	if opts.MaxWidth <= 0 {
		opts.MaxWidth = 80
	}

	formatBorder := opts.FormatBorder
	if formatBorder == nil {
		formatBorder = defaultBorderFormat
	}

	linePrefix := ""
	if opts.IncludePrefix {
		linePrefix = gray(Bar) + " "
	}

	// Calculate column widths
	// Use visible width for prefix and subtract border glyphs so final width fits MaxWidth
	available := opts.MaxWidth - visibleWidth(linePrefix)
	if available < 1 {
		available = 1
	}

	if opts.ShowBorders {
		borderGlyphs := len(headers) + 1 // number of vertical borders per line

		available -= borderGlyphs
		if available < 1 {
			available = 1
		}
	}

	columnWidths := calculateColumnWidths(headers, rows, available)

	// Normalize rows to have same number of columns as headers
	normalizedRows := normalizeRows(rows, len(headers))

	// Render table
	if opts.ShowBorders {
		renderTableWithBorders(out, headers, normalizedRows, columnWidths, linePrefix, formatBorder, opts)
	} else {
		renderTableWithoutBorders(out, headers, normalizedRows, columnWidths, linePrefix, opts)
	}
}

// calculateColumnWidths determines the optimal width for each column
func calculateColumnWidths(headers []string, rows [][]string, maxWidth int) []int {
	numCols := len(headers)

	// Compute natural content width (no padding) for each column
	natural := make([]int, numCols)
	for i, header := range headers {
		natural[i] = visibleWidth(header)
	}

	for _, row := range rows {
		for i := 0; i < numCols && i < len(row); i++ {
			if w := visibleWidth(row[i]); w > natural[i] {
				natural[i] = w
			}
		}
	}

	// Desired full widths (content + padding) and minimum widths
	desiredFull := make([]int, numCols)

	minWidth := make([]int, numCols)
	for i := 0; i < numCols; i++ {
		desiredFull[i] = natural[i] + 2 // 1 space padding each side
		if natural[i] <= 3 {
			m := natural[i] + 2 // show small values fully
			if m < 3 {
				m = 3
			}

			minWidth[i] = m
		} else {
			minWidth[i] = 5 // 2 padding + 3 for ellipsis
		}
	}

	// If even minimums exceed maxWidth, distribute fairly from 3 per column
	sumMin := 0
	for _, m := range minWidth {
		sumMin += m
	}

	if sumMin >= maxWidth {
		widths := make([]int, numCols)
		for i := range widths {
			widths[i] = 3
		}

		rem := maxWidth - 3*numCols
		for i := 0; rem > 0 && numCols > 0; i++ {
			idx := i % numCols
			widths[idx]++
			rem--
		}

		return widths
	}

	// Start from minimums and distribute remaining space toward desiredFull
	widths := make([]int, numCols)
	copy(widths, minWidth)

	remaining := maxWidth - sumMin

	// Compute wishes (how much each column wants to reach desiredFull)
	wishes := make([]int, numCols)
	totalWish := 0

	for i := 0; i < numCols; i++ {
		w := desiredFull[i] - minWidth[i]
		if w < 0 {
			w = 0
		}

		wishes[i] = w
		totalWish += w
	}

	if totalWish == 0 {
		return widths
	}

	// Proportional allocation by wish
	assigned := 0

	for i := 0; i < numCols; i++ {
		share := int(math.Floor(float64(remaining) * float64(wishes[i]) / float64(totalWish)))
		if share > wishes[i] {
			share = wishes[i]
		}

		widths[i] += share
		assigned += share
	}

	leftover := remaining - assigned

	// Distribute leftover one-by-one to columns still below desiredFull
	for leftover > 0 {
		progressed := false

		for i := 0; i < numCols && leftover > 0; i++ {
			if widths[i] < desiredFull[i] {
				widths[i]++
				leftover--
				progressed = true
			}
		}

		if !progressed {
			break
		}
	}

	return widths
}

// normalizeRows ensures all rows have the same number of columns as headers
func normalizeRows(rows [][]string, numCols int) [][]string {
	normalized := make([][]string, len(rows))
	for i, row := range rows {
		normalized[i] = make([]string, numCols)
		for j := 0; j < numCols; j++ {
			if j < len(row) {
				normalized[i][j] = row[j]
			} else {
				normalized[i][j] = ""
			}
		}
	}

	return normalized
}

// renderTableWithBorders renders a table with full borders
func renderTableWithBorders(out Writer, headers []string, rows [][]string, columnWidths []int, linePrefix string, formatBorder func(string) string, opts TableOptions) {
	numCols := len(headers)

	// Top border
	_, _ = fmt.Fprint(out, linePrefix)

	_, _ = fmt.Fprint(out, formatBorder(TableTopLeft))
	for i, width := range columnWidths {
		_, _ = fmt.Fprint(out, strings.Repeat(formatBorder(TableHorizontal), width))
		if i < numCols-1 {
			_, _ = fmt.Fprint(out, formatBorder(TableTopTee))
		}
	}

	_, _ = fmt.Fprint(out, formatBorder(TableTopRight))
	_, _ = fmt.Fprint(out, "\n")

	// Header row
	renderTableRow(out, headers, columnWidths, linePrefix, formatBorder, opts, true)

	// Header separator
	_, _ = fmt.Fprint(out, linePrefix)

	_, _ = fmt.Fprint(out, formatBorder(TableLeftTee))
	for i, width := range columnWidths {
		_, _ = fmt.Fprint(out, strings.Repeat(formatBorder(TableHorizontal), width))
		if i < numCols-1 {
			_, _ = fmt.Fprint(out, formatBorder(TableCross))
		}
	}

	_, _ = fmt.Fprint(out, formatBorder(TableRightTee))
	_, _ = fmt.Fprint(out, "\n")

	// Data rows
	for _, row := range rows {
		renderTableRow(out, row, columnWidths, linePrefix, formatBorder, opts, false)
	}

	// Bottom border
	_, _ = fmt.Fprint(out, linePrefix)

	_, _ = fmt.Fprint(out, formatBorder(TableBottomLeft))
	for i, width := range columnWidths {
		_, _ = fmt.Fprint(out, strings.Repeat(formatBorder(TableHorizontal), width))
		if i < numCols-1 {
			_, _ = fmt.Fprint(out, formatBorder(TableBottomTee))
		}
	}

	_, _ = fmt.Fprint(out, formatBorder(TableBottomRight))
	_, _ = fmt.Fprint(out, "\n")
}

// renderTableWithoutBorders renders a table without borders
func renderTableWithoutBorders(out Writer, headers []string, rows [][]string, columnWidths []int, linePrefix string, opts TableOptions) {
	// Header row
	renderTableRow(out, headers, columnWidths, linePrefix, nil, opts, true)

	// Data rows
	for _, row := range rows {
		renderTableRow(out, row, columnWidths, linePrefix, nil, opts, false)
	}
}

// renderTableRow renders a single table row
func renderTableRow(out Writer, row []string, columnWidths []int, linePrefix string, formatBorder func(string) string, opts TableOptions, isHeader bool) {
	_, _ = fmt.Fprint(out, linePrefix)

	for i, cell := range row {
		if i >= len(columnWidths) {
			break
		}

		width := columnWidths[i]
		alignment := TableAlignLeft

		// Get column alignment if specified
		if i < len(opts.ColumnAlignments) {
			alignment = opts.ColumnAlignments[i]
		}

		// Calculate available width for content (subtract padding only)
		contentWidth := width - 2 // 1 space on each side

		// Truncate the raw text first to avoid cutting ANSI sequences later
		cellContent := cell
		if visibleWidth(cellContent) > contentWidth {
			cellContent = truncateTableText(cellContent, contentWidth)
		}

		// Apply header styling after truncation so resets remain intact
		styledCell := cellContent
		if isHeader {
			styledCell = tableStyle(cellContent, opts.HeaderStyle, opts.HeaderColor)
		}

		// Apply alignment
		alignedCell := alignText(styledCell, contentWidth, alignment)

		// Add borders if needed
		if formatBorder != nil {
			_, _ = fmt.Fprint(out, formatBorder(TableVertical))
		}

		_, _ = fmt.Fprint(out, " ")
		_, _ = fmt.Fprint(out, alignedCell)
		_, _ = fmt.Fprint(out, " ")
	}

	// Right border
	if formatBorder != nil {
		_, _ = fmt.Fprint(out, formatBorder(TableVertical))
	}

	_, _ = fmt.Fprint(out, "\n")
}

// alignText aligns text within the given width, accounting for ANSI escape sequences
func alignText(text string, width int, alignment TableAlignment) string {
	// Get the visible width (excluding ANSI escape sequences)
	textWidth := visibleWidth(text)
	if textWidth >= width {
		return text
	}

	padding := width - textWidth

	switch alignment {
	case TableAlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad

		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	case TableAlignRight:
		return strings.Repeat(" ", padding) + text
	default: // TableAlignLeft
		return text + strings.Repeat(" ", padding)
	}
}

// truncateTableText truncates text to fit within the given width
func truncateTableText(text string, width int) string {
	if visibleWidth(text) <= width {
		return text
	}

	if width <= 0 {
		return ""
	}

	if width <= 3 {
		var b strings.Builder

		w := 0
		sawANSI := false
		sawReset := false

		for i := 0; i < len(text); {
			token, tw, next := scanANSIToken(text, i)
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

			if w+tw > width {
				break
			}

			b.WriteString(token)

			w += tw
		}

		if sawANSI && !sawReset {
			b.WriteString(Reset)
		}

		return b.String()
	}

	target := width - 3

	var b strings.Builder

	w := 0
	sawANSI := false
	sawReset := false

	for i := 0; i < len(text); {
		token, tw, next := scanANSIToken(text, i)
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
