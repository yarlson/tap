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
  "fmt"
  "github.com/yarlson/tap"
)
```

## Core helpers and types

All helpers create and close a terminal per call, unless I/O is overridden in tests.

- `func tap.Text(opts tap.TextOptions) string`

  - Options:
    - `Message string`
    - `Placeholder string`
    - `DefaultValue string`
    - `InitialValue string`
    - `Validate func(string) error` (return non-nil to block submission and show error)

- `func tap.Password(opts tap.PasswordOptions) string`

  - Same options as `TextOptions` (input is masked in the UI)

- `func tap.Confirm(opts tap.ConfirmOptions) bool`

  - Options:
    - `Message string`
    - `Active string` (label for true)
    - `Inactive string` (label for false)
    - `InitialValue bool`

- `type tap.SelectOption[T any] struct { Value T; Label, Hint string }`
- `type tap.SelectOptions[T any] struct { Message string; Options []tap.SelectOption[T]; InitialValue *T; MaxItems *int }`
- `func tap.Select[T any](opts tap.SelectOptions[T]) T`

- `type tap.MultiSelectOptions[T any] struct { Message string; Options []tap.SelectOption[T]; InitialValues []T; MaxItems *int }`
- `func tap.MultiSelect[T any](opts tap.MultiSelectOptions[T]) []T`

- Spinner

  - `type tap.SpinnerOptions struct { Indicator string; Frames []string; Delay time.Duration; CancelMessage, ErrorMessage string }`
  - `type tap.Spinner struct { /* unexported */ }`
  - `func tap.NewSpinner(opts tap.SpinnerOptions) *tap.Spinner`
  - `func (s *tap.Spinner) Start(msg string)`
  - `func (s *tap.Spinner) Message(msg string)`
  - `func (s *tap.Spinner) Stop(msg string, code int)` // 0=success, 1=cancel, >1=error
  - `func (s *tap.Spinner) IsCanceled() bool` (idiomatic) — `IsCancelled()` is kept for backward compat

- Progress
- Stream (live output)

  - `type tap.StreamOptions struct { ShowTimer bool }`
  - `type tap.Stream struct { /* unexported */ }`
  - `func tap.NewStream(opts tap.StreamOptions) *tap.Stream`
  - `func (s *tap.Stream) Start(msg string)`
  - `func (s *tap.Stream) WriteLine(line string)`
  - `func (s *tap.Stream) Pipe(r io.Reader)`
  - `func (s *tap.Stream) Stop(msg string, code int)` // 0=success, 1=cancel, >1=error

  - `type tap.ProgressOptions struct { Style string; Max, Size int }`
  - `type tap.Progress struct { /* unexported */ }`
  - `func tap.NewProgress(opts tap.ProgressOptions) *tap.Progress`
  - `func (p *tap.Progress) Start(msg string)`
  - `func (p *tap.Progress) Advance(step int, msg string)`
  - `func (p *tap.Progress) Message(msg string)`
  - `func (p *tap.Progress) Stop(msg string, code int)` // 0=success, 1=cancel, >1=error

- Messages and Box
  - `func tap.Intro(title string)`
  - `func tap.Outro(message string)`
  - `type tap.BoxOptions struct { Columns int; WidthFraction float64; WidthAuto bool; TitlePadding, ContentPadding int; TitleAlign, ContentAlign tap.BoxAlignment; Rounded, IncludePrefix bool; FormatBorder func(string) string }`
  - `func tap.Box(message string, title string, opts tap.BoxOptions)`
  - `func tap.GrayBorder(s string) string`
  - `func tap.CyanBorder(s string) string`

## Behavior and conventions

- **Typed returns**

  - `Text`/`Password` → `string`
  - `Confirm` → `bool`
  - `Select[T]` → `T`
  - `MultiSelect[T]` → `[]T`
  - If the user cancels, helpers return a reasonable zero value (`""`, `false`, `var zero T`).

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
name := tap.Text(tap.TextOptions{Message: "What is your name?"})
lang := tap.Text(tap.TextOptions{Message: fmt.Sprintf("Hi %s! Favorite language?", name)})
proceed := tap.Confirm(tap.ConfirmOptions{Message: "Proceed?", InitialValue: true})
if proceed {
  tap.Outro("Let's go! 🎉")
}
```

### Validation

```go
email := tap.Text(tap.TextOptions{
  Message: "Enter email:",
  Validate: func(s string) error {
    if !strings.Contains(s, "@") { return errors.New("please enter a valid email") }
    return nil
  },
})
```

### Select with typed values

```go
type Env string
env := tap.Select(tap.SelectOptions[Env]{
  Message: "Choose environment:",
  Options: []tap.SelectOption[Env]{
    {Value: "dev", Label: "Development", Hint: "Local"},
    {Value: "prod", Label: "Production", Hint: "Live"},
  },
})
```

### Multi-Select with typed values

```go
langs := tap.MultiSelect(tap.MultiSelectOptions[string]{
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

## Testing (override terminal I/O)

In tests, route I/O to mocks:

```go
in := core.NewMockReadable()
out := core.NewMockWritable()

tap.SetTermIO(in, out)
defer tap.SetTermIO(nil, nil)

_ = tap.Text(tap.TextOptions{Message: "Your name:"})
// feed input
in.EmitKeypress("A", core.Key{Name: "a"})
in.EmitKeypress("", core.Key{Name: "return"})
// assert frames in out.Buffer
```

## Gotchas and guidance for code generation

- For `Select[T]`, supply `[]tap.SelectOption[T]`. Labels/hints are optional.
- Always call `Stop` on `Spinner`/`Progress` to restore the terminal when used.
- For validation, return `error` (not `bool`) from the provided function; non-nil blocks submit and shows an error line.
- The library uses ANSI; ensure output is sent to a TTY when running examples.

## Minimal end-to-end example

```go
package main

import (
  "fmt"
  "strings"
  "errors"
  "github.com/yarlson/tap"
)

func main() {
  name := tap.Text(tap.TextOptions{Message: "Name:"})
  email := tap.Text(tap.TextOptions{
    Message: "Email:",
    Validate: func(s string) error {
      if !strings.Contains(s, "@") { return errors.New("please enter a valid email") }
      return nil
    },
  })
  ok := tap.Confirm(tap.ConfirmOptions{Message: fmt.Sprintf("Submit for %s?", name)})
  if ok { tap.Outro(fmt.Sprintf("Saved %s", email)) }
}
```
