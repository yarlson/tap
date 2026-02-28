# Input Handling in TAP Prompts

## Keyboard Processing

All prompts route keypresses through `prompt.handleKey()`:

1. **Error state recovery**: Any key except plain Return/Cancel clears error and returns to Active. Shift+Return does NOT clear error (allows continued editing on validation failure)
2. **User input tracking**: For text-based prompts (`track=true`), cursor-based updates via `updateUserInputWithCursor()`
3. **Movement key emission**: Cursor keys emit "cursor" events for select/multiselect/list components
4. **Vim alias support**: hjkl mapped to arrow keys when `track=false`
5. **Modifier detection**: Key struct includes `Shift` and `Ctrl` flags from terminal protocols
6. **Paste handling** (Textarea): Paste events insert PUA (Private Use Area) rune placeholders at cursor
7. **Special key handling**:
   - **Paste**: Insert content via PUA placeholder; resolved to original text on submit
   - **Return** (unmodified): Set value, run validation, transition to Submit/Error. In error state: re-validate
   - **Shift+Return**: Component-specific (Textarea: insert newline and clear error; others: may not apply)
   - **Escape/Ctrl+C**: Transition to Cancel immediately
   - **Backspace**: Delete character before cursor (or atomic PUA placeholder)
   - **Delete**: Delete character at cursor (or atomic PUA placeholder)
   - **Home**: Move cursor to start of current line (Textarea only)
   - **End**: Move cursor to end of current line (Textarea only)
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
2. Component (e.g., Textarea) may handle validation first:
   - If component sets state to StateError/StateSubmit/StateCancel, prompt skips validation
   - Otherwise, prompt runs `p.opts.Validate(s.Value)` if validator present
3. **Success**: State → Submit, prompt returns resolved value
4. **Failure**: Capture error message (string or `ValidationError`), State → Error, message visible in render
5. User can retry or cancel:
   - Pressing Return re-validates
   - Pressing Shift+Return (Textarea) clears error and allows continued editing
   - Pressing any other key clears error and returns to Active

Note: Validation function receives the _value_ (not user input), so custom validators can coerce/process before checking rules. For Textarea, validation receives the fully-resolved string with all paste placeholders expanded.

## Placeholder Behavior

Text-based prompts (Text, Password, Autocomplete) display placeholder when:

1. UserInput is empty
2. Placeholder is configured

Rendering shows inverse-video first character + dim remaining:

```
inverse(placeholder[0]) + dim(placeholder[1:])
```

Placeholder disappears once user types first character.

## Paste Buffer Handling (Textarea)

When a paste event occurs, the textarea:

1. **Store content**: Each paste increments a counter and stores content in `pasteBuffers` map
2. **Insert placeholder**: A PUA (Private Use Area) rune encodes the paste ID at cursor position
3. **Display as placeholder**: Renders as dim `[Text N]` where N is the paste ID
4. **Cursor behavior**:
   - Left/Right arrow keys skip over PUA placeholders atomically (move before/after the entire placeholder)
   - Backspace/Delete remove the entire PUA placeholder and clean its content from `pasteBuffers`
5. **Resolve on submit**: When user presses Return, `resolve()` expands all PUA runes to their original content
6. **Bracketed paste mode**: Textarea enables `ESC[?2004h` on init and disables `ESC[?2004l` on finalize

This design allows viewing multiple pastes visually while avoiding rendering complexity from long pasted text.

## Multiline Navigation (Textarea)

Textarea supports vertical navigation with Up/Down arrow keys:

- **Up**: Move cursor to previous line at same column position (or end of shorter line)
- **Down**: Move cursor to next line at same column position (or end of shorter line)
- Column position is tracked during vertical moves to maintain intuitive editing

Implementation uses helper functions:
- `cursorToLineCol(buf, cursor)` — converts flat cursor index to (line, column)
- `lineColToCursor(buf, line, col)` — converts (line, column) back to flat index
- `countBufferLines(buf)` — counts total lines in buffer

## Multi-key Sequences

No chord binding yet; each keypress is independent. Movement keys (up/down/left/right) are special-cased:
- Select/MultiSelect: Emit "cursor" events for menu navigation
- Textarea: Move cursor within text (vertical with Up/Down, horizontal with Left/Right)

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
