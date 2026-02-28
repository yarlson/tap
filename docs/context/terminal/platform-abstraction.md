# Terminal Abstraction in TAP

## Reader and Writer Interfaces

**Reader**: Source of keyboard input

```go
type Reader interface {
	io.Reader
	On(event string, handler func(string, Key))
}
```

- Emits "keypress" events with character and Key metadata
- Called with `On("keypress", func(char string, key Key) { ... })`
- Goroutine-safe; runs independent of prompt event loop

**Writer**: Destination for rendering and resize events

```go
type Writer interface {
	io.Writer
	On(event string, handler func())
	Emit(event string)
}
```

- Emits "resize" events when terminal size changes
- Called with `On("resize", func() { ... })`
- `Emit()` used internally by output utilities

## Terminal Singleton

**Module**: `internal/terminal/`

**Singleton pattern** (`terminal.go`):

- One terminal instance per application lifetime
- Opened via `terminal.New()` (goroutine-safe singleton pattern)
- Ensures exclusive access to raw keyboard input

**Fields**:

- `Reader`: Keyboard handler emitting "keypress" events
- `Writer`: Output and resize handler
- Platform-specific signal handlers (Unix/Windows)

## Platform-Specific Signal Handling

### Unix (`signal_unix.go`)

- Catches SIGWINCH for terminal resize events
- Catches SIGTSTP/SIGCONT for suspend/resume
- Uses `syscall` for signal registration
- Restores terminal settings on exit via `tcsetattr()`

### Windows (`signal_windows.go`)

- Uses Windows Console API for resize detection
- Polls console buffer info in goroutine
- No POSIX signal equivalents; custom event loop
- Console mode configured for raw input

## Key Struct and Keyboard Protocol Support

```go
type Key struct {
	Name  string // "a", "return", "escape", "left", "up", etc.
	Rune  rune   // Printable character if applicable
	Ctrl  bool   // Ctrl modifier
	Shift bool   // Shift modifier
}
```

Built by keyboard handler; emitted on keypress event.

**Keyboard Protocol Detection**:

TAP enables **extended keyboard mode** via kitty keyboard protocol (CSI sequence `\x1b[>4m`) to receive modifier information from terminals that support it. This allows detection of Shift+Enter, Shift+arrows, etc.

**Supported protocols**:
- **Kitty protocol**: `ESC[keycode;modifiersu` format
- **xterm modifyOtherKeys**: `ESC[27;modifier;keycodeu` format (older xterm variants)
- **Basic ANSI**: Fallback to standard arrow keys and control codes

**Parser functions** (`internal/terminal/terminal.go`):
- `parseCSI()` — Collects CSI parameters (supports both `;` and `:` separators)
- `resolveCSI()` — Maps CSI terminator and parameters to Key events
- `resolveModifiedKey()` — Decodes modifier bitmask to Shift/Ctrl flags

Modifier encoding: `modifier = 1 + bitmask` where bit 0 = Shift, bit 1 = Alt, bit 2 = Ctrl.

## I/O Override System

**Functions** (`prompt.go`):

- `SetTermIO(in Reader, out Writer)`: Override for testing
- `getOverrideWriter()`: Retrieve override writer without creating terminal
- `resolveWriter()`: Get stdout writer (output-only, no full terminal)

**Pattern for testing**:

```go
in := NewMockReadable()
out := NewMockWritable()
SetTermIO(in, out)
defer SetTermIO(nil, nil)
// Now prompts use mock I/O
```

**stdoutWriter** (`prompt.go`):

- Lightweight writer for output-only operations (Intro, Outro, Message)
- Writes to stdout directly
- No TTY opening, no goroutines, no event loop
- Prevents zombie "readKeys" goroutine that steals keypresses

## TTY Management

**Opening a terminal** (`prompt.go`):

- `terminal.New()` called by `runWithTerminal()` if I/O not overridden
- Opens `/dev/tty` (Unix) or Windows console
- Configures raw mode (no echo, no line buffering)
- Starts keyboard read loop in background goroutine
- Starts resize event monitoring

**Cleanup**:

- Prompt `finalize()` writes trailing newline and shows cursor
- Terminal remains open for subsequent operations (singleton pattern)
- User responsible for explicit cleanup if needed (not automatic)

## Multiline Handling

Terminal abstractions used for render cleanup:

- `CursorUp`: Move cursor up N lines
- `EraseLine` / `EraseDown`: Clear screen content
- `CursorHide` / `CursorShow`: Manage cursor visibility

`prompt.renderIfNeeded()` uses these to replace previous multiline frames without terminal artifacts.

## Event Loop Coordination

- Terminal Reader emits "keypress" events asynchronously
- Prompt's event loop processes them sequentially via unbounded queue
- Terminal Writer emits "resize" events in parallel
- Both events are non-blocking sends to `prompt.evCh`
