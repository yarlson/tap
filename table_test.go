package tap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable_BasicTable(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"Alice", "25", "New York"},
		{"Bob", "30", "London"},
		{"Charlie", "35", "Tokyo"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
	})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	text := strings.Join(frames, "")

	// Should contain headers
	assert.Contains(t, text, "Name")
	assert.Contains(t, text, "Age")
	assert.Contains(t, text, "City")

	// Should contain data rows
	assert.Contains(t, text, "Alice")
	assert.Contains(t, text, "Bob")
	assert.Contains(t, text, "Charlie")

	// Should contain table borders
	assert.Contains(t, text, "│") // vertical borders
	assert.Contains(t, text, "─") // horizontal borders
}

func TestTable_WithBorders(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"ID", "Status"}
	rows := [][]string{
		{"1", "Active"},
		{"2", "Inactive"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
	})

	text := strings.Join(out.GetFrames(), "")

	// Should contain corner symbols for borders
	assert.Contains(t, text, "┌") // top-left corner
	assert.Contains(t, text, "┐") // top-right corner
	assert.Contains(t, text, "└") // bottom-left corner
	assert.Contains(t, text, "┘") // bottom-right corner
}

func TestTable_WithoutBorders(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Name", "Value"}
	rows := [][]string{
		{"Test", "123"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: false,
	})

	text := strings.Join(out.GetFrames(), "")

	// Should not contain corner symbols
	assert.NotContains(t, text, "┌")
	assert.NotContains(t, text, "┐")
	assert.NotContains(t, text, "└")
	assert.NotContains(t, text, "┘")
}

func TestTable_ColumnAlignment(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Left", "Center", "Right"}
	rows := [][]string{
		{"A", "B", "C"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
		ColumnAlignments: []TableAlignment{
			TableAlignLeft,
			TableAlignCenter,
			TableAlignRight,
		},
	})

	// Test passes if no errors occur - alignment is visual
	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
}

func TestTable_HeaderStyling(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Header1", "Header2"}
	rows := [][]string{
		{"Data1", "Data2"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
		HeaderStyle: TableStyleBold,
		HeaderColor: TableColorCyan,
	})

	text := strings.Join(out.GetFrames(), "")

	// Should contain ANSI color codes for cyan
	assert.Contains(t, text, "\033[96m") // cyan color
}

func TestTable_EmptyTable(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Empty"}
	rows := [][]string{}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
	})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	text := strings.Join(frames, "")

	// Should still show headers even with no data
	assert.Contains(t, text, "Empty")
}

func TestTable_UnevenRows(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"A", "B", "C"}
	rows := [][]string{
		{"1", "2"},           // missing third column
		{"3", "4", "5", "6"}, // extra column
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
	})

	// Should handle uneven rows gracefully
	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
}

func TestTable_MaxWidth(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Very Long Header Name", "Short"}
	rows := [][]string{
		{"Very long data that might exceed width", "OK"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
		MaxWidth:    30,
	})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)
	// Should truncate or wrap content to fit max width
}

func TestTable_WithPrefix(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Name"}
	rows := [][]string{
		{"Test"},
	}

	Table(headers, rows, TableOptions{
		Output:        out,
		ShowBorders:   true,
		IncludePrefix: true,
	})

	text := strings.Join(out.GetFrames(), "")
	lines := strings.Split(text, "\n")

	// First non-empty line should start with prefix
	var first string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			first = l
			break
		}
	}

	if first != "" {
		// Should start with bar symbol and space (prefix)
		assert.True(t, strings.HasPrefix(first, "│ ") || strings.Contains(first, "│"))
	}
}

func TestTable_FormatBorder(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Test"}
	rows := [][]string{
		{"Data"},
	}

	Table(headers, rows, TableOptions{
		Output:       out,
		ShowBorders:  true,
		FormatBorder: gray,
	})

	text := strings.Join(out.GetFrames(), "")
	// Should contain gray-colored border characters
	assert.Contains(t, text, gray("┌"))
}

func TestTable_ColumnWidths(t *testing.T) {
	out := NewMockWritable()

	headers := []string{"Short", "Very Long Column Name"}
	rows := [][]string{
		{"A", "B"},
	}

	Table(headers, rows, TableOptions{
		Output:      out,
		ShowBorders: true,
	})

	frames := out.GetFrames()
	assert.NotEmpty(t, frames)

	// Check that we have reasonable output
	text := strings.Join(frames, "")
	assert.Contains(t, text, "Short")
	assert.Contains(t, text, "Very Long Column Name")
	assert.Contains(t, text, "A")
	assert.Contains(t, text, "B")
}
