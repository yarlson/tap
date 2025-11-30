# Introduction

This repository provides Tap, a Go library for building beautiful, interactive command‑line prompts and utilities (Text, Password, Confirm, Select/MultiSelect, Spinner, Progress, Stream, Box, Table). Contributions should maintain a clean public API, robust terminal handling, and predictable tests with mocked I/O.

Integration guidance:

- Start with design: decide whether your feature is an interactive prompt (uses Prompt) or a utility (Spinner/Progress/Stream/Table/Box).
- Follow existing patterns:
  - Interactive prompts open a temporary terminal per call via runWithTerminal, render using NewPrompt and snapshots, and finalize automatically.
  - Utilities resolve a Writer with resolveWriter and manage their own lifecycle (Start/Message/Stop).
- Keep dependencies minimal; prefer stdlib and existing third‑party deps already in go.mod.
- Ensure Go version compatibility; the module targets Go 1.24 and uses modern constructs (e.g., range over integers). Use Go 1.22+ at minimum; align any README references if needed.
- Add examples under examples/<feature>/main.go to document usage.
- Add unit tests mirroring existing patterns: override I/O with mocks, simulate keypresses, and assert rendered frames.

# Directory/File Structure

Key files and their roles (examples taken from this project):

- Root package (tap):
  - text.go, password.go, confirm.go: Interactive prompts using Prompt with Render closures and state snapshots.
  - select.go, multiselect.go: Typed selections with generics, cursor navigation via “cursor” events, and styled rendering.
  - spinner.go, progress.go, stream.go: Long‑running UI components with Start/Message/Stop, frame rendering, and OSC 9;4 integrations (osc.go).
  - table.go, box.go: Non‑interactive renderers for tabular output and boxed messages; handle width/alignments with runewidth.
  - prompt.go: Core prompt event loop, render diffing, state tracking, terminal session lifecycle, and helpers to integrate input/output.
  - types.go: Public option structs, shared enums/types (ClackState, Key), table types, I/O interfaces, and terminal control constants.
  - symbols.go: Unicode drawing glyphs and ANSI color/style helpers; Symbol(state) mapping.
  - osc.go: Terminal progress OSC 9;4 helpers (spin/set/clear).
  - messages.go: Intro/Message/Outro/Cancel convenience helpers.
  - mock.go: MockReadable/MockWritable and SetTermIO test overrides.
  - LICENSE, README.md, README.LLM.md: Project docs and API guide.
  - assets/demo.gif: Visual demo asset.

- Internal terminal layer:
  - internal/terminal/terminal.go: Keyboard input, ESC sequence handling, resize notifications, signal cleanup, and Reader/Writer implementations.

- Examples:
  - examples/text, password, confirm, select, multiselect: Basic interactive prompts.
  - examples/spinner, progress, stream: Long‑running and streaming examples.
  - examples/table, multiple: Output formatting and an end‑to‑end flow.

Placement guidelines for new code:

- New interactive prompt: add a new <feature>.go in root (package tap). If it requires new public options, add a <Feature>Options struct to types.go.
- New non‑interactive utility (e.g., formatter or UI helper): add a new file in root if it’s public API; keep terminal glue in resolveWriter pattern.
- Low‑level terminal changes: only under internal/terminal; preserve the public surface area.

# Naming Conventions

- Packages and files:
  - Package name: tap for root, internal/terminal for internals.
  - File names: match the component (text.go, progress.go). Tests use <file>\_test.go.
- Exported API:
  - Functions: UpperCamelCase for public entry points (Text, Password, Confirm, Select, MultiSelect, NewSpinner, NewProgress, NewStream, Table, Box, Intro/Outro/Message).
  - Option structs: <Type>Options (TextOptions, ConfirmOptions, TableOptions).
  - Types/enums: UpperCamelCase (ClackState, TableAlignment, TableStyle, TableColor).
  - Generics: Use descriptive type parameters where needed (SelectOption[T], Select[T]).
- Unexported internals:
  - Helpers and state: lowerCamelCase (renderStyledSelect, styledMultiSelectState).
  - Constants and vars scoped to file: lowerCamelCase unless shared public constants (StepActive, Bar, etc.).
- Events and state:
  - Event names: "keypress" (internal), and public Prompt events like "cursor", "key", "userInput", "confirm", "submit", "cancel", "finalize".
  - State constants: StateInitial, StateActive, StateError, StateSubmit, StateCancel; Symbol(state) maps to colored glyphs.
- API shape patterns:
  - Interactive prompt entry functions accept context.Context and an Options struct, return a typed value, and delegate to a private implementation when Input/Output are explicitly set.
  - Utilities expose Start/Message/Stop with a status code convention: 0=success, 1=cancel, >1=error.
  - Color/styling helpers wrap text in ANSI sequences (gray, cyan, etc.) and must reset.

# Testing Patterns

Use mocks and deterministic rendering; avoid real TTYs in tests.

- Override terminal I/O:
  ```go
  in := tap.NewMockReadable()
  out := tap.NewMockWritable()
  tap.SetTermIO(in, out)
  defer tap.SetTermIO(nil, nil)
  ```
- Simulate user interaction:
  ```go
  ctx := context.Background()
  // Launch a prompt then feed keys
  go func() {
    in.EmitKeypress("t", tap.Key{Name: "t"})
    in.EmitKeypress("e", tap.Key{Name: "e"})
    in.EmitKeypress("",  tap.Key{Name: "return"})
  }()
  res := tap.Text(ctx, tap.TextOptions{Message: "Enter:"})
  // Assert result and frames in out.Buffer or out.GetFrames()
  ```
- Verify frames/output:
  - For prompts: inspect out.GetFrames() to ensure correct symbols, colors, and structure (gray(Bar), cyan(BarEnd), etc.).
  - For utilities (Spinner/Progress/Stream): call Start/Message/Stop and assert expected sequences; ensure Stop clears frames and prints final line with the appropriate status symbol.
- Table/Box testing:
  - Render with a MockWritable, assert border glyphs and alignment; use controlled MaxWidth to keep outputs stable.
- Race and full suite:
  ```bash
  go test ./...
  go test -race ./...
  ```
- Notes:
  - Use context cancellation tests to verify zero‑value returns on abort.
  - Keep tests isolated and deterministic (no time.Sleep unless specifically validating timers; prefer dependency injection or small delays).
  - Prefer checking structural markers (glyphs, ANSI presence) over exact frame counts when terminal width could affect wrapping.
