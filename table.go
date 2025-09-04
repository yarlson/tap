package tap

import (
	"fmt"
	"math"
	"strings"

	"github.com/mattn/go-runewidth"
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
	columnWidths := calculateColumnWidths(headers, rows, opts.MaxWidth-len(linePrefix))

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
	widths := make([]int, numCols)

	// Calculate minimum width needed for each column
	for i, header := range headers {
		widths[i] = visibleWidth(header)
	}

	for _, row := range rows {
		for i := 0; i < numCols && i < len(row); i++ {
			if visibleWidth(row[i]) > widths[i] {
				widths[i] = visibleWidth(row[i])
			}
		}
	}

	// Add padding
	for i := range widths {
		widths[i] += 2 // 1 space on each side
	}

	// If total width exceeds maxWidth, proportionally reduce
	totalWidth := 0
	for _, w := range widths {
		totalWidth += w
	}

	if totalWidth > maxWidth {
		scale := float64(maxWidth) / float64(totalWidth)
		for i := range widths {
			widths[i] = int(math.Floor(float64(widths[i]) * scale))
			if widths[i] < 3 { // minimum width
				widths[i] = 3
			}
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
	if width <= 3 {
		return "..."
	}
	target := width - 3
	var b strings.Builder
	w := 0
	for _, r := range text {
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
