# Performance Notes (Agent D)

## Benchmarks

- No `Benchmark*` functions found in repo (search across Go files).

## Hot Path Evidence

- Prompt render loop recomputes ANSI-stripped widths and wrapping each update (`renderIfNeeded` -> `countPhysicalLines` -> `visibleWidth`) and calls `getColumns` each render. Evidence: `prompt.go:526`, `prompt.go:536`, `prompt.go:542`, `prompt.go:570`.
- Input edits rebuild rune slices on each keystroke (O(n) per key for long inputs). Evidence: `prompt.go:404`.
- Spinner/progress render loops allocate per tick (`fmt.Sprintf`, `strings.Repeat`, `strings.Join`) and call regex stripping each render. Evidence: `spinner.go:216`, `spinner.go:272`, `spinner.go:289`, `progress.go:206`, `progress.go:277`.
- Table rendering does multiple passes with `visibleWidth` per cell and truncation scans. Evidence: `table.go:68`, `table.go:269`, `table.go:345`.
- Stream stores all lines and repaints them on stop, O(n) memory and tail latency. Evidence: `stream.go:134`.
- `Reader.emit` copies handler slices under mutex per keypress (allocations + contention). Evidence: `internal/terminal/terminal.go:323`.
- Autocomplete rebuilds suggestion slices each keystroke and uses rune-slice inserts. Evidence: `autocomplete.go:64`, `autocomplete.go:187`.

## Minimal Benchmark Plan

- `BenchmarkVisibleWidth` for long ANSI strings (`prompt.go:542`).
- `BenchmarkUpdateUserInputWithCursor` for long inputs (`prompt.go:404`).
- `BenchmarkSpinnerRender` and `BenchmarkProgressRender` with `MockWritable` (`spinner.go:216`, `progress.go:206`, `mock.go:70`).
- `BenchmarkTableRender` for large grids (`table.go:68`).
