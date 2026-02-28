# Tap

[![CI](https://github.com/yarlson/tap/actions/workflows/ci.yml/badge.svg)](https://github.com/yarlson/tap/actions/workflows/ci.yml)

A Go library for building beautiful, interactive command-line prompts and terminal UI components.

## Features

- **Interactive Prompts**: Text input, password, confirmation, single/multi-select, autocomplete, and textarea
- **Progress Indicators**: Spinners with customizable frames and progress bars with multiple styles
- **Output Utilities**: Styled tables, boxed messages, and streaming output
- **Modern API**: Context-aware, generic types for select options, functional options pattern
- **Cross-Platform**: Unix and Windows terminal support

## Installation

```bash
go get github.com/yarlson/tap
```

**Requirements**: Go 1.24+

## Quick Start

### Text Input

```go
package main

import (
    "context"
    "fmt"

    "github.com/yarlson/tap"
)

func main() {
    name := tap.Text(context.Background(), tap.TextOptions{
        Message:     "Enter your name:",
        Placeholder: "Type something...",
    })
    fmt.Printf("Hello, %s!\n", name)
}
```

### Select Menu

```go
colors := []tap.SelectOption[string]{
    {Value: "red", Label: "Red", Hint: "The color of passion"},
    {Value: "blue", Label: "Blue", Hint: "The color of calm"},
    {Value: "green", Label: "Green", Hint: "The color of nature"},
}

result := tap.Select[string](context.Background(), tap.SelectOptions[string]{
    Message: "What's your favorite color?",
    Options: colors,
})
fmt.Printf("You selected: %s\n", result)
```

### Spinner

```go
spin := tap.NewSpinner(tap.SpinnerOptions{})
spin.Start("Connecting")
// ... do work ...
spin.Stop("Connected", 0)

// With hint (optional second line in gray)
spin.Stop("Installed 42 packages", 0, tap.StopOptions{
    Hint: "Run 'npm start' to begin",
})
```

### Progress Bar

```go
progress := tap.NewProgress(tap.ProgressOptions{
    Style: "heavy",
    Max:   100,
    Size:  40,
})

progress.Start("Downloading...")
for i := 0; i <= 100; i += 10 {
    progress.Advance(10, fmt.Sprintf("Downloading... %d%%", i))
}
progress.Stop("Download complete!", 0)

// With hint (optional second line in gray)
progress.Stop("Download complete!", 0, tap.StopOptions{
    Hint: "Saved to ~/Downloads/file.zip (128 MB)",
})
```

### Messages with Hints

```go
// Message, Outro, and Cancel support optional hints
tap.Message("Task completed", tap.MessageOptions{
    Hint: "Processed 42 items in 1.2s",
})

tap.Outro("All done!", tap.MessageOptions{
    Hint: "Run 'app --help' for more options",
})
```

## API Reference

### Interactive Prompts

| Function                                     | Description                 | Return Type |
| -------------------------------------------- | --------------------------- | ----------- |
| `Text(ctx, TextOptions)`                     | Single-line text input      | `string`    |
| `Password(ctx, PasswordOptions)`             | Masked password input       | `string`    |
| `Confirm(ctx, ConfirmOptions)`               | Yes/No confirmation         | `bool`      |
| `Select[T](ctx, SelectOptions[T])`           | Single-choice selection     | `T`         |
| `MultiSelect[T](ctx, MultiSelectOptions[T])` | Multiple-choice selection   | `[]T`       |
| `Textarea(ctx, TextareaOptions)`             | Multiline text input        | `string`    |
| `Autocomplete(ctx, AutocompleteOptions)`     | Text input with suggestions | `string`    |

### Progress Components

| Function                       | Description                          |
| ------------------------------ | ------------------------------------ |
| `NewSpinner(SpinnerOptions)`   | Animated spinner indicator           |
| `NewProgress(ProgressOptions)` | Progress bar with percentage         |
| `NewStream(StreamOptions)`     | Streaming output with optional timer |

### Output Utilities

| Function                              | Description             |
| ------------------------------------- | ----------------------- |
| `Table(headers, rows, TableOptions)`  | Render formatted tables |
| `Box(message, title, BoxOptions)`     | Render boxed messages   |
| `Intro(title, ...MessageOptions)`     | Display intro message   |
| `Outro(message, ...MessageOptions)`   | Display outro message   |
| `Message(message, ...MessageOptions)` | Display styled message  |

### Options Structs

#### TextOptions

```go
type TextOptions struct {
    Message      string
    Placeholder  string
    DefaultValue string
    InitialValue string
    Validate     func(string) error
    Input        Reader
    Output       Writer
}
```

#### TextareaOptions

```go
type TextareaOptions struct {
    Message      string
    Placeholder  string
    DefaultValue string
    InitialValue string
    Validate     func(string) error
    Input        Reader
    Output       Writer
}
```

#### SelectOptions

```go
type SelectOptions[T any] struct {
    Message      string
    Options      []SelectOption[T]
    InitialValue *T
    MaxItems     *int
    Input        Reader
    Output       Writer
}
```

#### SpinnerOptions

```go
type SpinnerOptions struct {
    Indicator string        // "dots" (default) or "timer"
    Frames    []string      // custom animation frames
    Delay     time.Duration // frame delay
}
```

#### ProgressOptions

```go
type ProgressOptions struct {
    Style string  // "heavy", "light", or "block"
    Max   int     // maximum value
    Size  int     // bar width in characters
}
```

#### TableOptions

```go
type TableOptions struct {
    ShowBorders      bool
    IncludePrefix    bool
    MaxWidth         int
    ColumnAlignments []TableAlignment
    HeaderStyle      TableStyle
    HeaderColor      TableColor
    FormatBorder     func(string) string
}
```

#### MessageOptions

```go
type MessageOptions struct {
    Output Writer
    Hint   string // Optional second line displayed in gray
}
```

#### StopOptions

Used with `Spinner.Stop()` and `Progress.Stop()` to add an optional hint line:

```go
type StopOptions struct {
    Hint string // Optional second line displayed in gray below the message
}
```

## Examples

Run interactive examples:

```bash
go run ./examples/text/main.go
go run ./examples/textarea/main.go
go run ./examples/password/main.go
go run ./examples/select/main.go
go run ./examples/multiselect/main.go
go run ./examples/confirm/main.go
go run ./examples/autocomplete/main.go
go run ./examples/spinner/main.go
go run ./examples/progress/main.go
go run ./examples/messages/main.go
go run ./examples/table/main.go
go run ./examples/stream/main.go
```

## Compatibility

### Platform Support

Terminal signal handling differs between Unix and Windows. The library includes platform-specific implementations:

- `internal/terminal/terminal_unix.go` - Unix signal handling
- `internal/terminal/terminal_windows.go` - Windows signal handling

### Environment Variables

| Variable | Required | Description                                    |
| -------- | -------- | ---------------------------------------------- |
| `TERM`   | No       | Terminal type for ANSI escape sequence support |

## Development

### Running Tests

```bash
go test ./...
```

With race detection:

```bash
go test -race ./...
```

### Linting

The project uses golangci-lint with configuration in `.golangci.yml`:

```bash
golangci-lint run
```

### Testing Patterns

Override terminal I/O for deterministic tests:

```go
in := tap.NewMockReadable()
out := tap.NewMockWritable()
tap.SetTermIO(in, out)
defer tap.SetTermIO(nil, nil)

// Simulate keypresses
go func() {
    in.EmitKeypress("h", tap.Key{Name: "h"})
    in.EmitKeypress("i", tap.Key{Name: "i"})
    in.EmitKeypress("", tap.Key{Name: "return"})
}()

result := tap.Text(ctx, tap.TextOptions{Message: "Enter:"})
```

## Troubleshooting

### Platform-specific signal handling

**Symptom**: Terminal signal handling differs between Unix and Windows

**Solution**: The library includes platform-specific implementations. Ensure you're building for the correct target platform. Check `internal/terminal/terminal_unix.go` and `internal/terminal/terminal_windows.go` for platform-specific behavior.

### ANSI sequence handling for width calculations

**Symptom**: Text width calculations may produce unexpected results with styled text

**Solution**: The library handles ANSI escape sequences in width calculations via `visibleWidth()` and `truncateToWidth()` functions in `ansi_utils.go`. Ensure styled text is properly formatted with reset sequences.

## Contributing

Contributions are welcome. When contributing:

- Follow existing code patterns and naming conventions
- Add unit tests using the mock I/O pattern
- Add examples under `examples/<feature>/main.go` for new features
- Run `go test ./...` and `golangci-lint run` before submitting

## License

MIT License - see [LICENSE](LICENSE) for details.
