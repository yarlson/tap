# 🚀 Tap

**Tap** is a Go port of the popular TypeScript [Clack](https://clack.cc/) library for building beautiful, interactive command-line applications.

<div align="center">
  <img src="assets/demo.gif" alt="Tap Demo" width="800">
</div>

> ⚠️ **Heavy Development**: This project is currently in heavy development. APIs may change, and some features are still being implemented. Use with caution in production environments.

## 🎯 About

Clack is a library that makes building interactive command-line applications effortless with beautiful, minimal, and opinionated CLI prompts. Tap brings this elegant experience to the Go ecosystem while maintaining the same design philosophy and user experience.

## ✅ What's Ported

### Core Functionality

- ✅ **Event-driven prompt system** - Complete with proper state management and event loop architecture
- ✅ **Terminal management** - Raw terminal mode, keyboard input handling, and cursor control
- ✅ **Mock testing utilities** - Full test coverage with mock input/output for reliable testing

### Prompts (Unstyled - Core Package)

- ✅ **Text Input** - Single-line text input with cursor navigation, validation, and default values
- ✅ **Confirm** - Yes/No prompts with keyboard navigation
- ✅ **Select** - Single selection from a list with cursor navigation and wrap-around

### Prompts (Styled - Prompts Package)

- ✅ **Text Input** - Beautifully styled text prompts with symbols, bars, placeholders, and error states
- ✅ **Confirm** - Styled confirmation prompts with radio button interface
- ✅ **Select** - Styled selection prompts with radio buttons, hints, and color-coded options
- ✅ **Progress Bar** - Animated progress with messages and final states
- ✅ **Symbols & Styling** - Unicode symbols, ANSI colors, and consistent visual design

### Still To Come

- 🔄 **Password Input** - Masked text input
- 🔄 **Multi-Select** - Multiple selection from a list
- 🔄 **Autocomplete** - Text input with autocomplete suggestions
- 🔄 **Spinner** - Loading indicators for long-running operations
- 🔄 **Group** - Grouped prompts for complex workflows
- 🔄 **Note/Log** - Informational messages and logging utilities
- 🔄 **Box** - Styled message boxes

## 🚀 Quick Start

### Installation

```bash
go get github.com/yarlson/tap
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/yarlson/tap/pkg/prompts"
    "github.com/yarlson/tap/pkg/terminal"
)

func main() {
    // Initialize terminal
    term, err := terminal.New()
    if err != nil {
        panic(err)
    }
    defer term.Close()

    // Text input
    name := prompts.Text(prompts.TextOptions{
        Message: "What's your name?",
        Input:   term,
        Output:  term,
    })

    // Confirmation
    confirmed := prompts.Confirm(prompts.ConfirmOptions{
        Message: fmt.Sprintf("Hello %s! Continue?", name),
        Input:   term,
        Output:  term,
    })

    if confirmed.(bool) {
        fmt.Println("Let's go! 🎉")
    }
}
```

### Advanced Features

```go
// Text with validation and default value
email := prompts.Text(prompts.TextOptions{
    Message:      "Enter your email:",
    Placeholder:  "user@example.com",
    DefaultValue: "anonymous@example.com",
    Validate: func(input string) error {
        if !strings.Contains(input, "@") {
            return errors.New("Please enter a valid email")
        }
        return nil
    },
    Input:  term,
    Output: term,
})

// Confirmation with custom labels
proceed := prompts.Confirm(prompts.ConfirmOptions{
    Message:      "Deploy to production?",
    Active:       "Deploy",
    Inactive:     "Cancel",
    InitialValue: false,
    Input:        term,
    Output:       term,
})
```

### Progress Bar

```go
// Progress bar with animated frames and messages
prog := prompts.NewProgress(prompts.ProgressOptions{
    Style:  "heavy",   // "light", "heavy", or "block"
    Max:    100,        // total units of work
    Size:   40,         // bar width in characters
    Output: term.Writer, // implements prompts.Writer
})

prog.Start("Processing...")

// Update progress and optionally the message
for i := 0; i <= 100; i += 10 {
    time.Sleep(200 * time.Millisecond)
    prog.Advance(10, fmt.Sprintf("Processing... %d%%", i))
}

// Stop with final status. code: 0=success, 1=cancel, other=error
prog.Stop("Done!", 0)
```

## 🏗️ Architecture

Tap follows a clean, event-driven architecture:

- **`pkg/core`** - Core prompt engine with unstyled, functional prompts
- **`pkg/prompts`** - Beautifully styled prompts built on top of core
- **`pkg/terminal`** - Terminal management and keyboard input handling

### Event Loop Design

Tap uses a pure event loop architecture (no mutexes or atomic operations) for excellent performance and race-condition-free operation:

```go
// Events flow through a single event loop
for event := range prompt.events {
    event(&state)           // Update state
    prompt.render(&state)   // Render changes
    prompt.updateSnapshot(&state) // Update atomic snapshot
}
```

## 🧪 Testing

All prompts include comprehensive test coverage with mock input/output:

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run specific package tests
go test ./pkg/prompts -v
```

## 📁 Project Structure

```
tap/
├── pkg/
│   ├── core/           # Core prompt engine (unstyled)
│   │   ├── prompt.go   # Main prompt implementation
│   │   ├── text.go     # Text input prompt
│   │   ├── confirm.go  # Confirmation prompt
│   │   ├── select.go   # Selection prompt
│   │   └── mock.go     # Testing utilities
│   ├── prompts/        # Styled prompts
│   │   ├── text.go     # Styled text input
│   │   ├── confirm.go  # Styled confirmation
│   │   ├── select.go   # Styled selection
│   │   └── symbols.go  # Unicode symbols & colors
│   └── terminal/       # Terminal management
│       └── terminal.go # Keyboard input & raw mode
└── examples/           # Usage examples
    ├── text/
    ├── confirm/
    ├── select/
    ├── progress/
    └── multiple/
```

## 🤝 Contributing

We welcome contributions! This project is in active development and there's lots to build.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yarlson/tap.git
cd tap

# Run tests
go test ./...

# Try examples
go run examples/text/main.go
go run examples/confirm/main.go
go run examples/select/main.go
go run examples/progress/main.go
go run examples/multiple/main.go
```

### What Needs Help

- **New Prompt Types**: Multi-Select, Password, Autocomplete
- **Enhanced Styling**: Better color support, themes, custom symbols
- **Documentation**: More examples, API documentation, tutorials
- **Testing**: Edge cases, cross-platform testing, performance tests
- **Bug Fixes**: Race conditions, rendering issues, keyboard handling

### Coding Standards

- Follow Go best practices and `gofmt` formatting
- Maintain test coverage above 80%
- Use event-driven architecture (no mutexes/atomics)
- Write clear, self-documenting code
- Add examples for new features

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Clack](https://clack.cc/)** - The original TypeScript library that inspired this project
- **[@eiannone/keyboard](https://github.com/eiannone/keyboard)** - Cross-platform keyboard input for Go
- The Go community for excellent tooling and libraries

## 🔗 Links

- **Original Clack**: https://clack.cc/
- **TypeScript Source**: https://github.com/bombshell-dev/clack
- **Go Port Issues**: https://github.com/yarlson/tap/issues
- **Documentation**: Coming soon!

---

Made with ❤️ for the Go community. Building interactive CLIs shouldn't be so hard!
