# TAP Progress Components

## Spinner

**Constructor**: `NewSpinner(SpinnerOptions) *Spinner`

**Behavior**:

- Animated indicator for indefinite-duration tasks
- Customizable animation frames and frame delay
- Two built-in indicators: "dots" (default) and "timer"

**Methods**:

- `Start(message string)`: Begin animation with message
- `Stop(message string, duration time.Duration, ...StopOptions) error`: End animation, display completion message
  - Optional `StopOptions.Hint` shows gray secondary text below main message
  - Duration currently unused; passed for future enhancement

**Implementation** (`spinner.go`):

- Ticker-based animation loop
- Frame rotation at configured delay
- Stops on first `Stop()` call
- Multiline cleanup via cursor positioning and EraseDown

**Example**:

```go
spin := tap.NewSpinner(tap.SpinnerOptions{})
spin.Start("Connecting...")
// ... work ...
spin.Stop("Connected", 0, tap.StopOptions{Hint: "Ready to proceed"})
```

## Progress Bar

**Constructor**: `NewProgress(ProgressOptions) *Progress`

**Behavior**:

- Linear progress indicator for bounded tasks (0 to Max)
- Three styles: "heavy", "light", "block"
- Shows percentage and visual bar

**Methods**:

- `Start(message string)`: Begin tracking with message
- `Advance(amount int, message string)`: Increment progress and update message
- `Stop(message string, duration time.Duration, ...StopOptions) error`: Complete, show final message
  - Optional `StopOptions.Hint` displays secondary info

**Implementation** (`progress.go`):

- Tracks current value and max
- Bar width configurable via Size
- Percentage calculated as (current / max) \* 100
- Multiline rendering with proper cleanup on re-render

**Example**:

```go
p := tap.NewProgress(tap.ProgressOptions{
    Style: "heavy",
    Max:   100,
    Size:  40,
})
p.Start("Downloading...")
for i := 0; i <= 100; i += 10 {
    p.Advance(10, fmt.Sprintf("Downloading... %d%%", i))
}
p.Stop("Complete!", 0, tap.StopOptions{Hint: "Saved to ~/Downloads/file.zip"})
```

## Stream

**Constructor**: `NewStream(StreamOptions) *Stream`

**Behavior**:

- Display streaming output with optional elapsed-time timer
- Line-buffered output (one line per Write call)
- Timer shows seconds elapsed or off-screen message

**Methods**:

- `Start(message string)`: Begin stream with header
- `Write(message string) error`: Append line to stream
- `Stop(message string, duration time.Duration, ...StopOptions) error`: Finalize stream
  - Optional `StopOptions.Hint` shows gray secondary text

**Implementation** (`stream.go`):

- Appends lines to internal buffer
- Renders with Bar prefix per line
- Timer ticker optional; can be disabled

**Example**:

```go
s := tap.NewStream(tap.StreamOptions{ShowTimer: true})
s.Start("Processing files...")
s.Write("Processed file1.txt")
s.Write("Processed file2.txt")
s.Stop("All done!", 0, tap.StopOptions{Hint: "2 files processed"})
```

## StopOptions

**Fields**:

- **Hint**: Optional second line displayed in gray below the main message

Used by `Spinner.Stop()`, `Progress.Stop()`, and `Stream.Stop()` for supplementary information.

## Timing and Cleanup

- All progress components show cursor during operation (no hide like prompts)
- Multiline cleanup: on re-render, move cursor up and EraseDown before writing new frame
- Duration parameter in `Stop()` reserved for future use; currently ignored
