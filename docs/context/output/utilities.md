# TAP Output Utilities

## Message, Intro, Outro

**Functions**:

- `Message(message, ...MessageOptions)`
- `Intro(title, ...MessageOptions)`
- `Outro(message, ...MessageOptions)`

**Behavior**:

- One-way text output; no interaction
- Intro is green, Outro is blue (color hints)
- Message uses configured color/styling
- Optional hint parameter displays gray secondary text below main message

**Implementation** (`messages.go`):

- Use `resolveWriter()` to get stdout writer (no full terminal needed)
- Emit styled string to output
- No event loop; synchronous writes

## Table

**Function**: `Table(headers []string, rows [][]string, TableOptions) error`

**Behavior**:

- Render formatted ASCII table with optional borders
- Column alignment (left/center/right)
- Header styling (bold/dim/normal + color)
- MaxWidth respected with truncation

**Features**:

- ShowBorders: Render box-drawing characters
- IncludePrefix: Add leading bar (│) per row
- ColumnAlignments: Per-column justify
- HeaderStyle/HeaderColor: Formatting for header row
- FormatBorder: Custom border rendering function

**Implementation** (`table.go`):

- Calculates column widths based on content and MaxWidth
- Handles ANSI escape sequences in width calculations via `visibleWidth()`
- Truncates overlong cells with ellipsis

## Box

**Function**: `Box(message string, title string, BoxOptions) error`

**Behavior**:

- Render message in bordered box with optional title
- Multiline message support with auto-wrap
- Configurable border style and padding

**Features**:

- ShowBorders: Draw box-drawing characters
- IncludePrefix: Add bar prefix to content lines
- Alignment: Center/left/right for title and message
- Padding: Space around content
- BorderStyle: Box-drawing character set (normal/bold/rounded)

**Implementation** (`box.go`):

- Wraps message to MaxWidth
- Centers title above box
- Handles multiline layout with proper spacing

## Styling

**Core functions** (`ansi_utils.go`):

- `gray(s)` / `dim(s)`: Lighter text (different gray intensities)
- `bold(s)`: Bright/bold weight
- `inverse(s)`: Swap foreground/background
- `strikethrough(s)`: Crossed-out text
- `foreground(color, s)`: Colored text (red, green, cyan, etc.)
- `background(color, s)`: Colored background

All wrap with ANSI codes; nested styles work (each adds codes).

## Symbols

**Constants** (`symbols.go`):

- **Bar**: `│` (vertical line prefix)
- **StepActive**: `◆` (active/current state)
- **StepSubmit**: `✓` (submitted/completed)
- **StepCancel**: `✕` (cancelled)
- **StepError**: `✘` (error state)

**Function**: `Symbol(state ClackState) string` returns the symbol for a given state.

## Terminal Control

**Control characters**:

- `CursorHide`: Hide terminal cursor
- `CursorShow`: Show terminal cursor
- `EraseLine`: Clear current line
- `CursorUp`: Move cursor up one line
- `EraseDown`: Clear from cursor to end of screen

Used by `prompt.renderIfNeeded()` for multiline cleanup and re-render.
