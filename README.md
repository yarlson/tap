# Tap

**Beautiful, interactive command-line prompts for Go** — A Go port inspired by the TypeScript [Clack](https://clack.cc/) library.

<div>
  <img src="assets/demo.gif" alt="Tap Demo" width="1400">
</div>

## Why Tap?

Building CLI applications shouldn't require wrestling with terminal complexities. Tap provides elegant, type-safe prompts with beautiful Unicode styling, letting you focus on your application logic instead of terminal management.

## Features

- 🎯 **Type-safe prompts** with Go generics for strongly-typed selections
- 🎨 **Beautiful styling** with consistent Unicode symbols and colors
- ⚡ **Zero-config** terminal management with automatic cleanup
- 🧪 **Testing utilities** with built-in mocks for reliable testing
- 📦 **Minimal dependencies** — only essential terminal libraries

### Available Components

- **Text Input** — Single-line input with validation, placeholders, and defaults
- **Password Input** — Masked input for sensitive data
- **Confirm** — Yes/No prompts with customizable labels
- **Select** — Single selection from typed options with hints
- **MultiSelect** — Multiple selection with checkboxes
- **Context Support** — All interactive prompts support context cancellation and timeouts
- **Progress Bar** — Animated progress indicators (light, heavy, block styles)
- **Spinner** — Loading indicators with dots, timer, or custom frames
- **Stream** — Real-time output with start/write/stop lifecycle
- **Messages** — Intro, outro, and styled message boxes

## Installation

```bash
go get github.com/yarlson/tap@latest
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/yarlson/tap"
)

func main() {
    ctx := context.Background()

    tap.Intro("Welcome! 👋")

    name := tap.Text(ctx, tap.TextOptions{
        Message: "What's your name?",
        Placeholder: "Enter your name...",
    })

    confirmed := tap.Confirm(ctx, tap.ConfirmOptions{
        Message: fmt.Sprintf("Hello %s! Continue?", name),
    })

    if confirmed {
        tap.Outro("Let's go! 🎉")
    }
}
```

## API Examples

### Text Input with Validation

```go
email := tap.Text(ctx, tap.TextOptions{
    Message:      "Enter your email:",
    Placeholder:  "user@example.com",
    DefaultValue: "anonymous@example.com",
    Validate: func(input string) error {
        if !strings.Contains(input, "@") {
            return errors.New("Please enter a valid email")
        }
        return nil
    },
})
```

### Type-Safe Selection

```go
type Environment string

environments := []tap.SelectOption[Environment]{
    {Value: "dev", Label: "Development", Hint: "Local development"},
    {Value: "staging", Label: "Staging", Hint: "Pre-production testing"},
    {Value: "production", Label: "Production", Hint: "Live environment"},
}

env := tap.Select(ctx, tap.SelectOptions[Environment]{
    Message: "Choose deployment target:",
    Options: environments,
})

// env is strongly typed as Environment
```

### Progress Indicators

```go
// Progress Bar
progress := tap.NewProgress(tap.ProgressOptions{
    Style: "heavy",  // "light", "heavy", or "block"
    Max:   100,
    Size:  40,
})

progress.Start("Processing...")
for i := 0; i <= 100; i += 10 {
    time.Sleep(200 * time.Millisecond)
    progress.Advance(10, fmt.Sprintf("Step %d/10", i/10+1))
}
progress.Stop("Complete!", 0)

// Spinner
spinner := tap.NewSpinner(tap.SpinnerOptions{})
spinner.Start("Loading...")
// ... do work ...
spinner.Stop("Done!", 0)
```

## OSC 9;4 Integration (Terminal Progress)

Tap emits OSC 9;4 control sequences to signal progress/spinner state to compatible terminals. Unsupported terminals ignore these sequences (no-op), so visuals remain unchanged.

What’s emitted automatically:

- Spinner:
  - Start → indeterminate: `ESC ] 9 ; 4 ; 3 ST`
  - Stop → always clear: `ESC ] 9 ; 4 ; 0 ST`
- Progress:
  - On render when percent changes → `ESC ] 9 ; 4 ; 1 ; <PCT> ST`
  - Stop → always clear: `ESC ] 9 ; 4 ; 0 ST`

Notes:

- Terminator: Tap uses ST (`ESC \\`) for robustness. Some terminals also accept BEL (`\a`).
- Throttling: Progress only emits a new percentage when it changes to avoid spam.
- Multiplexers: tmux/screen may swallow OSC sequences unless configured to passthrough.

### Multiple Selection

```go
languages := []tap.SelectOption[string]{
    {Value: "go", Label: "Go"},
    {Value: "python", Label: "Python"},
    {Value: "javascript", Label: "JavaScript"},
}

selected := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
    Message: "Which languages do you use?",
    Options: languages,
})

fmt.Printf("You selected: %v\n", selected)
```

### Messages, Box, and Table

```go
tap.Intro("Summary")
tap.Message("Here's a table summary of your selections:")

headers := []string{"Field", "Value"}
rows := [][]string{
  {"Name", name},
  {"Languages", strings.Join(selected, ", ")},
}

tap.Table(headers, rows, tap.TableOptions{
  ShowBorders:   true,
  IncludePrefix: true,
  HeaderStyle:   tap.TableStyleBold,
  HeaderColor:   tap.TableColorCyan,
})
```

### Context Support and Cancellation

All interactive prompts support Go's context package for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

name := tap.Text(ctx, tap.TextOptions{
    Message: "Enter your name (30s timeout):",
})

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5*time.Second)
    cancel() // Cancel after 5 seconds
}()

result := tap.Confirm(ctx, tap.ConfirmOptions{
    Message: "Quick decision needed:",
})
```

### Styled Messages

```go
// Message box with custom styling
tap.Box("This is important information!", "⚠️ Warning", tap.BoxOptions{
    Rounded:       true,
    FormatBorder:  tap.CyanBorder,
    TitleAlign:    tap.BoxAlignCenter,
    ContentAlign:  tap.BoxAlignCenter,
})
```

## Testing

Tap includes comprehensive testing utilities. Override terminal I/O in tests:

```go
func TestYourPrompt(t *testing.T) {
    // Create mock I/O
    mockInput := tap.NewMockReadable()
    mockOutput := tap.NewMockWritable()

    // Override terminal I/O for testing
    tap.SetTermIO(mockInput, mockOutput)
    defer tap.SetTermIO(nil, nil)

    // Simulate user input
    go func() {
        mockInput.EmitKeypress("test", tap.Key{Name: "t"})
        mockInput.EmitKeypress("", tap.Key{Name: "return"})
    }()

    result := tap.Text(ctx, tap.TextOptions{Message: "Enter text:"})
    assert.Equal(t, "test", result)
}
```

Run tests:

```bash
go test ./...
go test -race ./...  # with race detection
```

## Examples

Explore working examples in the [`examples/`](examples/) directory:

```bash
go run examples/text/main.go      # Text input
go run examples/select/main.go    # Selection menus
go run examples/progress/main.go  # Progress bars
go run examples/multiple/main.go  # Complete workflow
```

## Architecture

Tap uses an event-driven architecture with atomic state management for race-condition-free operation. The library automatically handles:

- Terminal raw mode setup/cleanup
- Keyboard input processing
- Cursor positioning and output formatting
- Cross-platform compatibility

The main package provides a clean API while internal packages handle terminal complexity.

## Status

Tap API is **stable** and production-ready. The library follows semantic versioning and maintains backward compatibility.

## Contributing

Contributions welcome! Please:

- Follow Go best practices and maintain test coverage
- Include examples for new features
- Update documentation as needed

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- **[Clack](https://clack.cc/)** — The original TypeScript library that inspired this project
- **[@eiannone/keyboard](https://github.com/eiannone/keyboard)** — Cross-platform keyboard input
- The Go community for excellent tooling and feedback

---

Built with ❤️ for developers who value simplicity and speed.
