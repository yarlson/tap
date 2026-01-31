# Reliability & Security Findings (Agent C)

## Potential Panics / Hazards

- Empty `SelectOptions.Options` can panic when cursor indexes an empty slice. Evidence: `select.go:71`, `select.go:127`.
- Empty `MultiSelectOptions.Options` can panic on cursor wrap and selection. Evidence: `multiselect.go:89`, `multiselect.go:107`.
- `Spinner` reuse can panic due to closing `stopCh` twice; `Start` does not recreate the channel. Evidence: `spinner.go:90`, `spinner.go:154`.
- `Progress` reuse can panic due to `stopChan` double-close; `Start` does not recreate the channel. Evidence: `progress.go:82`, `progress.go:57`.

## Concurrency / Resource Leaks

- `unboundedQueue` goroutine in prompt loop never terminates (event channel never closed). Evidence: `prompt.go:122`, `prompt.go:501`.
- `terminal` ESC timer uses `time.AfterFunc` without synchronization; races on `escPending`, `escPrefix`, `escBuf`. Evidence: `internal/terminal/terminal.go:241`.
- Resize handler goroutine leak: `signal.Notify` with no `signal.Stop` on cleanup. Evidence: `internal/terminal/terminal_unix.go:31`.

## Data Loss / Error Handling

- `Stream.Pipe` uses `bufio.Scanner` without increasing buffer and ignores `scanner.Err()`, risking truncation of long lines and silent errors. Evidence: `stream.go:85`.

## Dependency Posture

- Direct deps: `keyboard`, `go-runewidth`, `x/term`, `testify` (`go.mod:5`).
- Indirect deps include `yaml.v3`, `x/sys` (`go.mod:12`).
- Vulnerability status: Unknown (no audit data present in repo).
