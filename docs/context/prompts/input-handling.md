# Input Handling in TAP Prompts

## Keyboard Processing

All prompts route keypresses through `prompt.handleKey()`:

1. **Error state recovery**: Any key (except Return/Cancel) clears error and returns to Active
2. **User input tracking**: For text-based prompts (`track=true`), cursor-based updates via `updateUserInputWithCursor()`
3. **Movement key emission**: Cursor keys emit "cursor" events for select/multiselect/list components
4. **Vim alias support**: hjkl mapped to arrow keys when `track=false`
5. **Special key handling**:
   - **Return**: Set value, run validation, transition to Submit/Error
   - **Escape/Ctrl+C**: Transition to Cancel immediately
   - **Backspace**: Delete character before cursor
   - **Delete**: Delete character at cursor
   - **Tab**: Insert tab character at cursor
   - **Space**: Insert space at cursor
   - **Regular chars**: Insert at cursor position (ASCII 32-126)

## Cursor Positioning

`updateUserInputWithCursor()` maintains cursor position in rune space:

- **Bounds checking**: Cursor clamped to [0, len(runes)]
- **Left/Right**: Decrement/increment cursor within bounds
- **Backspace**: Delete rune before cursor, decrement position
- **Delete**: Delete rune at cursor, keep position
- **Insert**: Insert rune at cursor, increment position
- **Tab/Space**: Insert special characters at cursor
- **Regular char**: Insert printable ASCII at cursor

No multi-rune tracking for surrogate pairs yet; single-rune basis.

## User Input vs. Value

- **UserInput**: Raw string as typed by user; available via `p.UserInputSnapshot()`
- **Value**: Processed result; set by prompt logic (DefaultValue, validated user input, or boolean for confirm)
- **InitialValue**: Pre-populated value used if user submits without typing
- **InitialUserInput**: Pre-populated text shown at cursor start; allows edit-and-submit workflow

Prompts track user input to provide real-time feedback (placeholder replacement, cursor display) while value holds the final result.

## Validation Flow

1. User presses Return
2. `prompt.handleKey()` calls `p.opts.Validate(s.Value)` if validator present
3. **Success**: State → Submit
4. **Failure**: Capture error message (string or `ValidationError`), State → Error, message visible in render
5. User can retry or cancel; pressing any other key clears error and returns to Active

Note: Validation function receives the _value_ (not user input), so custom validators can coerce/process before checking rules.

## Placeholder Behavior

Text-based prompts (Text, Password, Autocomplete) display placeholder when:

1. UserInput is empty
2. Placeholder is configured

Rendering shows inverse-video first character + dim remaining:

```
inverse(placeholder[0]) + dim(placeholder[1:])
```

Placeholder disappears once user types first character.

## Multi-key Sequences

No chord binding yet; each keypress is independent. Movement keys (up/down/left/right) are special-cased for select/multiselect; other components ignore them.

Example: hjkl support for Vim users in Select menus:

```go
if alias := getMovementAlias(key.Name); alias != "" {
	p.Emit("cursor", alias)  // hjkl → arrow key equivalents
}
```

## Buffering and Race Conditions

- Keyboard events are non-blocking sends to `prompt.evCh` (unbounded queue)
- Rapid keypresses never block the reader goroutine
- Older fix: `prompt.evCh` is write-only, preventing re-entrant sends
- Modern fix: `unboundedQueue()` goroutine with slice buffer guarantees FIFO ordering and no blocking
