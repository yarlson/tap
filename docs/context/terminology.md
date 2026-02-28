# TAP Terminology

## Core Concepts

**Prompt** — The central event loop (`Prompt` struct) managing input, rendering, and state for any interactive prompt component.

**Reader/Writer** — Go interfaces abstracting terminal I/O. Reader receives keypress events. Writer receives render calls and resize events.

**State** — A `ClackState` value representing the prompt's lifecycle: `StateInitial`, `StateActive`, `StateError`, `StateSubmit`, `StateCancel`.

**Event loop** — The goroutine (`prompt.loop()`) that processes event handlers, renders on state changes, and finalizes when Submit or Cancel occurs.

**Unbounded queue** — A channel-based event queue using a buffer slice to prevent deadlocks when keypresses arrive faster than they're processed.

**Render function** — A closure passed to `NewPrompt()` that formats the current prompt state as a string for terminal output.

**Finalize** — Terminal cleanup: show cursor, write newline, emit finalize/submit/cancel events, return result.

**Terminal singleton** — One active terminal per session (created by `terminal.New()`). Ensures exclusive access to raw keyboard input.

**Override I/O** — Test mechanism via `SetTermIO()` to inject mock readers/writers instead of opening a real terminal.

## Component Types

**Prompt component** — A public function (e.g., `Text()`, `Textarea()`) accepting a context and options struct, returning a result. Examples:

- `Text(ctx, TextOptions) string`
- `Textarea(ctx, TextareaOptions) string`
- `Select[T](ctx, SelectOptions[T]) T`

**Progress component** — A struct with `Start()`, `Advance()`, and `Stop()` methods for long-running operations.

- `NewSpinner(SpinnerOptions) *Spinner`
- `NewProgress(ProgressOptions) *Progress`
- `NewStream(StreamOptions) *Stream`

**Output utility** — A function that writes styled output to stdout without interactive input.

- `Table(headers, rows, opts)`
- `Box(message, title, opts)`
- `Intro()`, `Outro()`, `Message()`

## Options

**Options struct** — Configuration for a prompt/component. All include `Input` and `Output` fields for I/O override. Examples:

- `TextOptions`: Message, Placeholder, DefaultValue, InitialValue, Validate
- `TextareaOptions`: Message, Placeholder, DefaultValue, InitialValue
- `SelectOptions[T]`: Message, Options, InitialValue, MaxItems
- `SpinnerOptions`: Indicator, Frames, Delay
- `ProgressOptions`: Style, Max, Size
- `StopOptions`: Hint (for Spinner.Stop and Progress.Stop)

**SelectOption[T]** — A key-value pair with label and hint for select menu items.

**MessageOptions** — Configuration for output messages: Output, Hint.

## Symbols and Styling

**Bar** — A vertical line symbol (│) used as a prefix in prompts and output.

**Symbol** — A function returning state-specific symbols: `StepActive` (◆), `StepSubmit` (✓), `StepCancel` (✕), `StepError` (✘).

**Styling functions** — ANSI color/formatting helpers: `gray()`, `dim()`, `inverse()`, `strikethrough()`, `bold()`, etc.

## Testing

**Mock readable** — `NewMockReadable()` simulates keyboard input via `EmitKeypress()`.

**Mock writable** — `NewMockWritable()` captures rendered frames via `GetFrames()`.

**Mock key** — A `Key` struct with name (e.g., "return", "escape", "left") and optional rune/modifiers.
