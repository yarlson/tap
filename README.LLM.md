# Tap: LLM Usage Guide

This document gives language models a compact, precise map of Tap’s public API so you can confidently generate working code that uses this library.

Tap is a Go library for building interactive, clack-style terminal prompts (text/password/confirm/select/multiselect), spinners, progress bars, streams, and message boxes.

- Module path: `github.com/yarlson/tap`
- Runtime model: each helper opens a terminal internally, runs, and restores the TTY
- Return values are typed (not `any`) and cancel maps to a sensible zero value
- A small test-only override is available to inject I/O without a real terminal

## Import

```go
import (
  "context"
  "fmt"
  "github.com/yarlson/tap"
)
```

## Core helpers and types

All helpers create and close a terminal per call, unless I/O is overridden in tests.

- `func tap.Text(ctx context.Context, opts tap.TextOptions) string`

  - Options:
    - `Message string`
    - `Placeholder string`
    - `DefaultValue string`
    - `InitialValue string`
    - `Validate func(string) error` (return non-nil to block submission and show error)
    - `Input tap.Reader` (optional)
    - `Output tap.Writer` (optional)

- `func tap.Autocomplete(ctx context.Context, opts tap.AutocompleteOptions) string`

  - Options:
    - `Message string`
    - `Placeholder string`
    - `DefaultValue string`
    - `InitialValue string`
    - `Validate func(string) error`
    - `Suggest func(string) []string` (return candidate suggestions for current input)
    - `MaxResults int` (max suggestions to display; default 5)
    - `Input tap.Reader` (optional)
    - `Output tap.Writer` (optional)
  - Behavior:
    - Arrow Up/Down navigates suggestions; `Tab` accepts the highlighted suggestion
    - `Enter` submits the current input (or accepted suggestion)
    - `Ctrl+C`/`Esc` cancel and return an empty string

- `func tap.Password(ctx context.Context, opts tap.PasswordOptions) string`

  - Same options as `TextOptions` (input is masked in the UI)
  - Includes `Input tap.Reader` and `Output tap.Writer`

- `func tap.Confirm(ctx context.Context, opts tap.ConfirmOptions) bool`

  - Options:
    - `Message string`
    - `Active string` (label for true)
    - `Inactive string` (label for false)
    - `InitialValue bool`
    - `Input tap.Reader`
    - `Output tap.Writer`

- `type tap.SelectOption[T any] struct { Value T; Label, Hint string }`
- `type tap.SelectOptions[T any] struct { Message string; Options []tap.SelectOption[T]; InitialValue *T; MaxItems *int; Input tap.Reader; Output tap.Writer }`
- `func tap.Select[T any](ctx context.Context, opts tap.SelectOptions[T]) T`

- `type tap.MultiSelectOptions[T any] struct { Message string; Options []tap.SelectOption[T]; InitialValues []T; MaxItems *int; Input tap.Reader; Output tap.Writer }`
- `func tap.MultiSelect[T any](ctx context.Context, opts tap.MultiSelectOptions[T]) []T`

- Spinner

  - `type tap.SpinnerOptions struct { Indicator string; Frames []string; Delay time.Duration; Output tap.Writer; CancelMessage, ErrorMessage string }`
  - `type tap.Spinner struct { /* unexported */ }`
  - `func tap.NewSpinner(opts tap.SpinnerOptions) *tap.Spinner`
  - `func (s *tap.Spinner) Start(msg string)`
  - `func (s *tap.Spinner) Message(msg string)`
  - `func (s *tap.Spinner) Stop(msg string, code int)` // 0=success, 1=cancel, >1=error
  - `func (s *tap.Spinner) IsCancelled() bool`

- Progress

  - `type tap.ProgressOptions struct { Style string; Max, Size int; Output tap.Writer }`
  - `type tap.Progress struct { /* unexported */ }`
  - `func tap.NewProgress(opts tap.ProgressOptions) *tap.Progress`
  - `func (p *tap.Progress) Start(msg string)`
  - `func (p *tap.Progress) Advance(step int, msg string)`
  - `func (p *tap.Progress) Message(msg string)`
  - `func (p *tap.Progress) Stop(msg string, code int)` // 0=success, 1=cancel, >1=error

- Stream (live output)

  - `type tap.StreamOptions struct { ShowTimer bool; Output tap.Writer }`
  - `type tap.Stream struct { /* unexported */ }`
  - `func tap.NewStream(opts tap.StreamOptions) *tap.Stream`
  - `func (s *tap.Stream) Start(msg string)`
  - `func (s *tap.Stream) WriteLine(line string)`
  - `func (s *tap.Stream) Pipe(r io.Reader)`
  - `func (s *tap.Stream) Stop(msg string, code int)` // 0=success, 1=cancel, >1=error

- Messages, Box, and Table
  - `func tap.Intro(title string)`
  - `func tap.Outro(message string)`
  - `func tap.Message(message string)`
  - `type tap.BoxOptions struct { Columns int; WidthFraction float64; WidthAuto bool; TitlePadding, ContentPadding int; TitleAlign, ContentAlign tap.BoxAlignment; Rounded, IncludePrefix bool; FormatBorder func(string) string }`
  - `func tap.Box(message string, title string, opts tap.BoxOptions)`
  - `func tap.GrayBorder(s string) string`
  - `func tap.CyanBorder(s string) string`
  - `type tap.TableOptions struct { Output tap.Writer; ShowBorders, IncludePrefix bool; MaxWidth int; ColumnAlignments []tap.TableAlignment; HeaderStyle tap.TableStyle; HeaderColor tap.TableColor; FormatBorder func(string) string }`
  - `func tap.Table(headers []string, rows [][]string, opts tap.TableOptions)`

## Behavior and conventions

- **Context Support**

  - All interactive prompt functions (Text, Password, Confirm, Select, MultiSelect) require a `context.Context` as the first parameter
  - Use `context.Background()` for basic usage, or pass custom contexts for cancellation/timeouts
  - If context is cancelled, prompts return zero values (empty string, false, etc.)

- **Typed returns**

  - `Text`/`Password` → `string`
  - `Confirm` → `bool`
  - `Select[T]` → `T`
  - `MultiSelect[T]` → `[]T`
  - If the user cancels, helpers return a reasonable zero value (`""`, `false`, `var zero T`).

- **Keybindings**

  - Navigate: Arrow keys or `h`/`j`/`k`/`l`
  - Submit: `Enter`
  - Cancel: `Ctrl+C` or `Esc`
  - Toggle (MultiSelect): `Space`
  - Accept suggestion (Autocomplete): `Tab`

- **Validation errors** (Text/Password)

  - Provide `Validate: func(string) error { ... }`
  - If non-nil error is returned on submit, the prompt stays active and shows a yellow error line below the input:
    - Yellow left bar for the input line
    - Yellow bottom-left corner (└) for the error line and a yellow error message

- **Lifecycle**
  - Helpers open and close a terminal per call; you do not need to manage sessions
  - `Spinner`/`Progress` create a terminal under the hood and close it on `Stop` (unless I/O is overridden in tests)

## Basic usage

```go
ctx := context.Background()
name := tap.Text(ctx, tap.TextOptions{Message: "What is your name?"})
lang := tap.Text(ctx, tap.TextOptions{Message: fmt.Sprintf("Hi %s! Favorite language?", name)})
proceed := tap.Confirm(ctx, tap.ConfirmOptions{Message: "Proceed?", InitialValue: true})
if proceed {
  tap.Outro("Let's go! 🎉")
}
```

### Validation

```go
email := tap.Text(ctx, tap.TextOptions{
  Message: "Enter email:",
  Validate: func(s string) error {
    if !strings.Contains(s, "@") { return errors.New("please enter a valid email") }
    return nil
  },
})
```

### Autocomplete

```go
// Suggest function to filter a static list
suggest := func(input string) []string {
  all := []string{"Go", "Golang", "Python", "Rust", "Java"}
  if input == "" { return all }
  low := strings.ToLower(input)
  var out []string
  for _, s := range all {
    if strings.Contains(strings.ToLower(s), low) {
      out = append(out, s)
    }
  }
  return out
}

val := tap.Autocomplete(ctx, tap.AutocompleteOptions{
  Message:     "Search language:",
  Placeholder: "Start typing...",
  Suggest:     suggest,
  MaxResults:  6,
})
```

### Select with typed values

```go
type Env string
env := tap.Select(ctx, tap.SelectOptions[Env]{
  Message: "Choose environment:",
  Options: []tap.SelectOption[Env]{
    {Value: "dev", Label: "Development", Hint: "Local"},
    {Value: "prod", Label: "Production", Hint: "Live"},
  },
})
```

### Multi-Select with typed values

```go
langs := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
  Message: "Choose languages:",
  Options: []tap.SelectOption[string]{
    {Value: "go", Label: "Go"},
    {Value: "py", Label: "Python"},
    {Value: "js", Label: "JavaScript"},
  },
  InitialValues: []string{"go"},
})
fmt.Println(langs) // []string{"go", ...}
```

### Spinner/Progress/Stream

```go
sp := tap.NewSpinner(tap.SpinnerOptions{Indicator: "dots"})
sp.Start("Connecting")
// ... do work ...
sp.Stop("Connected", 0)

pr := tap.NewProgress(tap.ProgressOptions{Style: "heavy", Max: 100, Size: 40})
pr.Start("Processing…")
for i := 0; i < 10; i++ { pr.Advance(10, fmt.Sprintf("%d/10", i+1)) }
pr.Stop("Complete", 0)

st := tap.NewStream(tap.StreamOptions{ShowTimer: true})
st.Start("Building project")
st.WriteLine("step 1: deps")
st.WriteLine("step 2: compile")
st.Stop("Done", 0)
```

### Messages

```go
tap.Message("Here's a summary table:")
```

### Table

```go
headers := []string{"Field", "Value"}
rows := [][]string{{"Name", name}, {"Languages", strings.Join(langs, ", ")}}

tap.Table(headers, rows, tap.TableOptions{
  ShowBorders:   true,
  IncludePrefix: true,
  HeaderStyle:   tap.TableStyleBold,
  HeaderColor:   tap.TableColorCyan,
})
```

## Testing (override terminal I/O)

In tests, route I/O to mocks:

```go
in := tap.NewMockReadable()
out := tap.NewMockWritable()

tap.SetTermIO(in, out)
defer tap.SetTermIO(nil, nil)

_ = tap.Text(ctx, tap.TextOptions{Message: "Your name:"})
// feed input
in.EmitKeypress("A", tap.Key{Name: "a"})
in.EmitKeypress("", tap.Key{Name: "return"})
// assert frames in out.Buffer
```

Alternatively, pass per-call I/O via options: set `Input` and `Output` on the options (e.g., `TextOptions`, `ConfirmOptions`) to avoid global overrides.

## Gotchas and guidance for code generation

- For `Select[T]`, supply `[]tap.SelectOption[T]`. Labels/hints are optional.
- Always call `Stop` on `Spinner`/`Progress`/`Stream` to restore the terminal when used.
- For validation, return `error` (not `bool`) from the provided function; non-nil blocks submit and shows an error line.
- The library uses ANSI; ensure output is sent to a TTY when running examples.
- Use `context.Background()` for basic usage, or create custom contexts for cancellation/timeout behavior.
- Generics inference: you can omit type parameters on `Select`/`MultiSelect` if the compiler can infer them from `SelectOptions[T]`/`MultiSelectOptions[T]`.

## Minimal end-to-end example

```go
package main

import (
  "context"
  "fmt"
  "strings"
  "errors"
  "github.com/yarlson/tap"
)

func main() {
  ctx := context.Background()
  name := tap.Text(ctx, tap.TextOptions{Message: "Name:"})
  email := tap.Text(ctx, tap.TextOptions{
    Message: "Email:",
    Validate: func(s string) error {
      if !strings.Contains(s, "@") { return errors.New("please enter a valid email") }
      return nil
    },
  })
  ok := tap.Confirm(ctx, tap.ConfirmOptions{Message: fmt.Sprintf("Submit for %s?", name)})
  if ok { tap.Outro(fmt.Sprintf("Saved %s", email)) }
}
```
