# TAP Prompt Types

## Text Input

**Function**: `Text(ctx, TextOptions) string`

**Behavior**:

- Single-line text input with cursor positioning
- Supports placeholder, default value, initial value, validation
- User input tracked in real-time
- Renders placeholder with inverse-video first char when empty

**Key implementation** (`text.go`):

- Uses `NewPrompt()` with `track=true`
- Render shows message, placeholder, cursor, and input text
- Validation via custom function

## Password

**Function**: `Password(ctx, PasswordOptions) string`

**Behavior**:

- Masked single-line input
- No placeholder support
- Returns plaintext (caller responsible for security)

**Key implementation** (`password.go`):

- Similar to Text but masks input display
- Cursor tracking enabled; input hidden

## Confirm

**Function**: `Confirm(ctx, ConfirmOptions) bool`

**Behavior**:

- Binary yes/no choice
- Default to inactive (N)
- Auto-submit on Y/N keypress (no Return needed)
- Configurable Active/Inactive labels

**Key implementation** (`confirm.go`):

- `track=false` (no cursor position)
- Emits "confirm" event on y/n; sets value and state
- Renders active/inactive choice with toggle

## Select

**Function**: `Select[T](ctx, SelectOptions[T]) T`

**Behavior**:

- Single-choice menu from generic options
- Cursor-based navigation (up/down arrows or j/k)
- Configurable max visible items (scrolling)
- Returns selected value

**Key implementation** (`select.go`):

- `track=false`
- Movement keys emit "cursor" events
- Highlights current selection
- Return key submits

## MultiSelect

**Function**: `MultiSelect[T](ctx, MultiSelectOptions[T]) []T`

**Behavior**:

- Multiple-choice menu
- Toggle items with Space or Return
- Cursor navigation with up/down
- Returns slice of selected values

**Key implementation** (`multiselect.go`):

- `track=false`
- Space/Return toggles current item
- Shows checkmarks for selected
- Return with no items marked still submits (empty slice)

## Autocomplete

**Function**: `Autocomplete(ctx, AutocompleteOptions) string`

**Behavior**:

- Text input with live suggestions
- `Suggest` callback provides completion list based on current input
- Up/down navigate suggestions; Return selects or submits
- MaxResults limits shown suggestions (default 5)

**Key implementation** (`autocomplete.go`):

- `track=false` (manual buffer management)
- Render calls `Suggest()` on each keystroke
- Shows filtered suggestions below input
- Movement keys navigate suggestions; other keys edit input

## Textarea

**Function**: `Textarea(ctx, TextareaOptions) string`

**Behavior**:

- Multiline text input with cursor-based editing
- **Shift+Return**: Insert newline within text (multiline editing)
- **Return**: Submit the entire textarea content
- **Up/Down arrows**: Navigate between lines, maintaining column position
- **Left/Right arrows**: Move cursor within current line
- **Backspace/Delete**: Delete characters across line boundaries
- Placeholder when empty
- Default and initial values supported

**Key implementation** (`textarea.go`):

- `track=false` (manual input buffer management)
- Render multiline text with bar prefix per line
- Key handlers:
  - Shift+Return calls `buf = slices.Insert(buf, cur, '\n')`
  - Regular Return calls `p.SetValue(string(buf))` to submit
  - Up/Down use `cursorToLineCol()` and `lineColToCursor()` for smart vertical navigation
  - Left/Right adjust cursor within bounds
  - Backspace/Delete modify buffer at cursor position

## All Prompt Components

Common fields in options:

- **Message**: The prompt text displayed to user
- **Input/Output**: Optional I/O override (for testing)
- **Validate**: Optional validation function

Selection prompts additionally:

- **InitialValue**: Pre-selected value
- **MaxItems**: Max visible items (select/multiselect)
- **Hint**: Used in SelectOption for submenu descriptions

Text-based prompts additionally:

- **Placeholder**: Gray help text when empty
- **DefaultValue**: Used if user submits without input
- **InitialValue** (Text) / **InitialValues** (MultiSelect): Pre-filled text
