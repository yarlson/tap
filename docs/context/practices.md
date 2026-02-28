# TAP Development Practices

## Component Patterns

### Public Prompt Components

Every interactive prompt has a public wrapper function:

```go
func Text(ctx context.Context, opts TextOptions) string {
	if opts.Input != nil && opts.Output != nil {
		return text(ctx, opts)  // Use provided I/O
	}
	return runWithTerminal(func(in Reader, out Writer) string {
		// Fill missing I/O, delegate to private implementation
		if opts.Input == nil { opts.Input = in }
		if opts.Output == nil { opts.Output = out }
		return text(ctx, opts)
	})
}
```

Private implementation creates a `Prompt` with custom render logic, context, and validation.

### Render Functions

All render functions:

1. Snapshot current state via `p.StateSnapshot()`, `p.UserInputSnapshot()`, `p.ValueSnapshot()`
2. Format output using Bar prefix and Symbol(state)
3. Return styled ANSI string; changes trigger re-render

### Multiline Rendering

Multiline prompts (Textarea, Autocomplete, Table, Box):

- Use `countPhysicalLines()` to track frame height
- On re-render, move cursor up and `EraseDown` before writing new frame
- Prevents terminal artifacts when content shrinks

### Input Tracking

Text input components set `NewPrompt(...)` with `track=true` (default):

- Cursor position updates automatically for left/right/backspace/delete
- User input stored in `UserInputSnapshot()`
- Non-tracking components (Autocomplete, Textarea): manually manage input in render logic

### Event Loop Safety

- `unboundedQueue()` prevents channel deadlocks by buffering events in a goroutine slice
- Re-entrant calls (`inEventLoop` flag) avoid channel sends during event processing
- All state updates happen in the event loop, then rendered and snapshotted

## Testing Patterns

### Mock I/O Setup

```go
in := NewMockReadable()
out := NewMockWritable()
SetTermIO(in, out)
defer SetTermIO(nil, nil)
```

### Simulating Keypresses

```go
go func() {
	in.EmitKeypress("h", Key{Name: "h", Rune: 'h'})
	in.EmitKeypress("i", Key{Name: "i", Rune: 'i'})
	in.EmitKeypress("", Key{Name: "return"})
}()
result := Text(ctx, TextOptions{...})
```

### Simulating Paste Events

```go
go func() {
	in.EmitKeypress("t", Key{Name: "t", Rune: 't'})
	in.EmitPaste("large content from clipboard")
	in.EmitKeypress("", Key{Name: "return"})
}()
result := Textarea(ctx, TextareaOptions{...})
```

`EmitPaste()` emits a Key with Name "paste" and Content set. Textarea collects paste content into a buffer and renders as a `[Text N]` placeholder, resolving the original content on submit.

### Frame Verification

Render functions must emit frames:

- Check `out.GetFrames()` for state-dependent output
- Verify bars, symbols, placeholder text present
- Use frame count for timing (wait for final state before asserting)

## Naming Conventions

- **Prompt components**: Verb(Noun) — `Text()`, `Select()`, `Textarea()`
- **Progress components**: Noun constructor — `NewSpinner()`, `NewProgress()`
- **Output utilities**: Verb or noun — `Table()`, `Box()`, `Message()`, `Intro()`, `Outro()`
- **Options structs**: `<Feature>Options` — `TextOptions`, `SelectOptions[T]`, `SpinnerOptions`
- **Private implementations**: lowercase — `text()`, `textarea()`, `select_[T]()`

## Error Handling

### Validation

- User-supplied validation function may return `ValidationError` (custom) or standard error
- Validation can occur at two levels:
  1. **Component-level**: Complex components (Textarea) validate before submit and set state directly
  2. **Prompt-level**: Prompt engine calls `opts.Validate()` only if component hasn't set state
- Validation errors trigger `StateError`; user can retry or cancel
- Components should validate the fully-resolved value (e.g., Textarea expands paste placeholders before validating)

### Component State Control

Components like Textarea can set state to StateError, StateSubmit, or StateCancel directly:

- If component sets state, prompt skips validation and proceeds with finalization
- This allows components to implement custom validation logic or state management
- Example: Textarea validates after resolving paste placeholders, sets error state with custom message

### Type Safety

- Generic select components use `SelectOptions[T]` and return `T`
- Concrete returns (Text, Password) are `string`; Confirm is `bool`
- Mock assertions use type assertion to extract values from `ValueSnapshot()`

## Multiline Handling

### ANSI Width Calculations

- `visibleWidth()` strips ANSI escape sequences before calculating width
- Uses `go-runewidth` for emoji/wide-character support
- `countPhysicalLines()` splits by newline, estimates wraps per line

### Terminal Soft-wrap Awareness

- Default terminal width detected via `getColumns()`, falls back to 80
- Frame sizing accounts for ANSI codes not occupying screen space
- Cursor positioning corrected during re-renders

## State Transitions

Valid transitions:

```
Initial → Active → (Submit | Cancel | Error)
Active → Error (validation failure)
Error → Active (user continues typing)
```

Cleanup happens when state is Submit or Cancel:

- `finalize()` called to emit events and restore terminal
- Result returned to caller
- Channel-based coordination ensures no re-entry
