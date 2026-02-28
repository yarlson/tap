# TAP — Terminal UI Library

## What

TAP is a Go library providing a unified API for building beautiful, interactive command-line prompts and terminal UI components. It abstracts terminal I/O, keyboard handling, rendering, and cleanup across Unix and Windows platforms.

**Current Version**: Production-grade with cross-platform support

## Architecture

TAP uses an **event-driven prompt engine** at its core:

- **`Prompt` struct** (`prompt.go`): Central event loop managing state transitions, rendering, and input handling
- **Unbounded event queue**: Prevents deadlocks from rapid keypress sequences
- **Terminal abstraction**: `Reader`/`Writer` interfaces decouple I/O from prompt logic
- **Platform-specific terminal**: `internal/terminal/` handles Unix signals and Windows console API
- **Mock I/O**: Test utilities (`mock.go`) for deterministic testing

## Core Flow

1. **Prompt creation**: User calls `Text()`, `Select()`, `Textarea()`, etc.
2. **Terminal setup**: `runWithTerminal()` opens a real terminal or uses override I/O for tests
3. **Event loop**: `prompt.loop()` processes keyboard/resize events and re-renders
4. **State machine**: State transitions (Initial → Active → Submit/Cancel/Error)
5. **Cleanup**: Cursor shown, terminal restored, result returned to caller

## System State

- **Interactive prompts**: Text, Password, Confirm, Select, MultiSelect, Autocomplete, Textarea
- **Progress indicators**: Spinner, Progress (with stop hints)
- **Output utilities**: Table, Box, Messages (Intro/Outro/Message with hints)
- **Stream component**: Streaming output with optional timer
- **I/O override system**: For testing and programmatic flows
- **Terminal singleton pattern**: One active terminal per session

## Capabilities

- **Styled prompts** with ANSI colors and formatting
- **Cursor-based input** with movement (left/right), deletion, backspace
- **Placeholder text** with inverse-video first character
- **Validation** with custom error messages (supports `ValidationError`)
- **Default/initial values** for pre-filling prompts
- **Keyboard shortcuts**: vim-style movement (hjkl), Escape/Ctrl+C for cancel, Return to submit
- **Multiline rendering** with soft-wrap detection and cursor repositioning
- **Event emission**: for custom subscribers (on state, value, cursor, key events)
- **Cross-platform**: Unix (signals) and Windows (console API) terminal handling

## Tech Stack

- **Go**: 1.24+ (uses `slices`, `strings`, generic types)
- **Terminal**: `golang.org/x/term` (standard TTY operations)
- **Keyboard**: `github.com/mattn/go-tty` (raw key reading)
- **Text width**: `github.com/mattn/go-runewidth` (ANSI-aware width calculations)
- **Testing**: `testing` package + mock I/O utilities
- **CI/CD**: GitHub Actions with golangci-lint (v9)
