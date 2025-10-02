package main

import (
	"fmt"

	"github.com/yarlson/tap"
)

func main() {
	tap.Intro("📊 Table Writer Example")

	// Example 1: Basic table with borders
	fmt.Println("\n1. Basic Table with Borders:")

	headers := []string{"Name", "Age", "City", "Status"}
	rows := [][]string{
		{"Alice Johnson", "28", "New York", "Active"},
		{"Bob Smith", "34", "London", "Active"},
		{"Charlie Brown", "22", "Tokyo", "Inactive"},
		{"Diana Prince", "31", "Paris", "Active"},
	}

	tap.Table(headers, rows, tap.TableOptions{
		ShowBorders: true,
	})

	// Example 2: Table without borders
	fmt.Println("\n2. Table without Borders:")

	simpleHeaders := []string{"ID", "Product", "Price"}
	simpleRows := [][]string{
		{"1", "Laptop", "$999"},
		{"2", "Mouse", "$25"},
		{"3", "Keyboard", "$75"},
	}

	tap.Table(simpleHeaders, simpleRows, tap.TableOptions{
		ShowBorders: false,
	})

	// Example 3: Table with custom styling and alignment
	fmt.Println("\n3. Styled Table with Custom Alignment:")

	styledHeaders := []string{"Metric", "Value", "Trend"}
	styledRows := [][]string{
		{"Users", "1,234", "↗ +12%"},
		{"Revenue", "$45,678", "↗ +8%"},
		{"Errors", "23", "↘ -5%"},
	}

	tap.Table(styledHeaders, styledRows, tap.TableOptions{
		ShowBorders: true,
		HeaderStyle: tap.TableStyleBold,
		HeaderColor: tap.TableColorCyan,
		ColumnAlignments: []tap.TableAlignment{
			tap.TableAlignLeft,   // Metric
			tap.TableAlignRight,  // Value
			tap.TableAlignCenter, // Trend
		},
		FormatBorder: tap.GrayBorder,
	})

	// Example 4: Table with prefix (like in a box)
	fmt.Println("\n4. Table with Prefix:")

	prefixedHeaders := []string{"Service", "Status", "Uptime"}
	prefixedRows := [][]string{
		{"API Gateway", "🟢 Running", "99.9%"},
		{"Database", "🟢 Running", "99.8%"},
		{"Cache", "🟡 Warning", "95.2%"},
		{"Queue", "🔴 Down", "0%"},
	}

	tap.Table(prefixedHeaders, prefixedRows, tap.TableOptions{
		ShowBorders:   true,
		IncludePrefix: true,
		HeaderStyle:   tap.TableStyleBold,
		HeaderColor:   tap.TableColorGreen,
	})

	// Example 5: Table with constrained width
	fmt.Println("\n5. Table with Max Width Constraint:")

	longHeaders := []string{"Very Long Column Name", "Another Long Header", "Short"}
	longRows := [][]string{
		{"This is a very long piece of data that should be truncated", "More long data here", "OK"},
		{"Short", "Also short", "Fine"},
	}

	tap.Table(longHeaders, longRows, tap.TableOptions{
		ShowBorders: true,
		MaxWidth:    60,
		HeaderStyle: tap.TableStyleBold,
		HeaderColor: tap.TableColorYellow,
	})

	// Example 6: Empty table
	fmt.Println("\n6. Empty Table (headers only):")

	emptyHeaders := []string{"No Data", "Available"}
	emptyRows := [][]string{}

	tap.Table(emptyHeaders, emptyRows, tap.TableOptions{
		ShowBorders: true,
		HeaderStyle: tap.TableStyleDim,
		HeaderColor: tap.TableColorGray,
	})

	// Example 7: Table with uneven rows
	fmt.Println("\n7. Table with Uneven Rows:")

	unevenHeaders := []string{"A", "B", "C"}
	unevenRows := [][]string{
		{"1", "2"},                // missing third column
		{"3", "4", "5", "6", "7"}, // extra columns
		{"8", "9", "10"},          // correct number
	}

	tap.Table(unevenHeaders, unevenRows, tap.TableOptions{
		ShowBorders: true,
		HeaderStyle: tap.TableStyleBold,
		HeaderColor: tap.TableColorRed,
	})

	tap.Outro("Thanks for exploring the table writer! 🚀")
}
