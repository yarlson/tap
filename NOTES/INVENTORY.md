# Repo Inventory (Agent A)

## Module

- `go.mod:1` module `github.com/yarlson/tap`
- `go.mod:3` Go version `1.24.0`

## Public package: `tap`

### Core prompt system

- `prompt.go:16` `PromptOptions`, `EventHandler`, `Prompt`
- `prompt.go:118` `NewPrompt`
- `prompt.go:167` `NewPromptWithTracking`
- `prompt.go:60` prompt methods: `StateSnapshot`, `UserInputSnapshot`, `CursorSnapshot`, `ErrorSnapshot`, `ValueSnapshot`, `SetValue`, `SetImmediateValue`, `On`, `Emit`
- `prompt.go:235` `Prompt.Prompt` event loop
- `prompt.go:649` `SetTermIO`

### Interactive prompts

- `text.go:9` `Text`
- `password.go` (exported `Password`, see file)
- `confirm.go` (exported `Confirm`, see file)
- `select.go:16` `Select`
- `multiselect.go:17` `MultiSelect`
- `autocomplete.go` (exported `Autocomplete`, see file)

### Utilities

- `spinner.go:12` `SpinnerOptions`, `Spinner`
- `spinner.go:57` `NewSpinner`
- `spinner.go:95` `Spinner.Start`, `Message`, `Stop`, `IsCancelled`
- `progress.go:13` `ProgressOptions`, `Progress`
- `progress.go:48` `NewProgress`
- `progress.go:90` `Progress.Start`, `Advance`, `Message`, `Stop`
- `stream.go:12` `StreamOptions`, `Stream`
- `stream.go:31` `NewStream`
- `stream.go:43` `Stream.Start`, `WriteLine`, `Pipe`, `Stop`
- `table.go:10` `Table`
- `box.go:11` `BoxAlignment`, `BoxOptions`, border helper fns
- `box.go:44` `Box`
- `messages.go:9` `MessageOptions`, `Cancel`, `Intro`, `Outro`, `Message`

### Types and constants

- `types.go:12` options types: `TextOptions`, `PasswordOptions`, `ConfirmOptions`, `SelectOption`, `SelectOptions`, `MultiSelectOptions`, `AutocompleteOptions`, `TableOptions`
- `types.go:84` `ClackState`, `Key`, state constants (`StateInitial`, `StateActive`, `StateCancel`, `StateSubmit`, `StateError`), cursor constants
- `types.go:136` table style constants
- `symbols.go:6` symbols and ANSI style constants; `Symbol` (`symbols.go:64`)

### Test utilities

- `mock.go:8` `MockReadable`, `NewMockReadable`, `EmitKeypress`, `SendKey`
- `mock.go:70` `MockWritable`, `NewMockWritable`, `Emit`, `GetFrames`

## Internal package: `internal/terminal`

- `internal/terminal/terminal.go:15` `Key`, `Reader`, `Writer`, `Terminal`, `New`
- `internal/terminal/terminal_unix.go`, `terminal_windows.go` for platform-specific signal handling

## Examples

- `examples/*/main.go` show usage for `Text`, `Select`, `MultiSelect`, `Confirm`, `Password`, `Spinner`, `Progress`, `Stream`, `Table`, `Autocomplete`.
