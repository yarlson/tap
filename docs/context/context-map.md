# Context Map — TAP Documentation Index

## Core Documentation

- **[summary.md](summary.md)** — Project overview, architecture, system state, capabilities, tech stack
- **[terminology.md](terminology.md)** — Term definitions and concepts
- **[practices.md](practices.md)** — Development patterns, testing, naming conventions, state machines

## Domain Documentation

### Prompts — Interactive Input Components

- **[prompts/prompt-types.md](prompts/prompt-types.md)** — All prompt types: Text, Password, Confirm, Select, MultiSelect, Autocomplete, Textarea
- **[prompts/input-handling.md](prompts/input-handling.md)** — Keyboard processing, cursor positioning, validation flow, placeholder behavior

### Progress — Long-running Operation Indicators

- **[progress/indicators.md](progress/indicators.md)** — Spinner, Progress bar, Stream components with examples

### Output — One-way Display Utilities

- **[output/utilities.md](output/utilities.md)** — Message, Intro, Outro, Table, Box, Styling, Symbols, Terminal control

### Terminal — Platform Abstraction

- **[terminal/platform-abstraction.md](terminal/platform-abstraction.md)** — Reader/Writer interfaces, singleton pattern, platform-specific signal handling, I/O override for testing

## File Organization

```
docs/context/
├── summary.md
├── terminology.md
├── practices.md
├── context-map.md (this file)
├── prompts/
│   ├── prompt-types.md
│   └── input-handling.md
├── progress/
│   └── indicators.md
├── output/
│   └── utilities.md
└── terminal/
    └── platform-abstraction.md
```

## How to Use This Documentation

1. **New to TAP?** Start with [summary.md](summary.md) for architecture overview.
2. **Building a prompt?** See [prompts/prompt-types.md](prompts/prompt-types.md) for component APIs.
3. **Debugging input behavior?** Check [prompts/input-handling.md](prompts/input-handling.md).
4. **Adding progress feedback?** See [progress/indicators.md](progress/indicators.md).
5. **Testing in isolation?** See [terminal/platform-abstraction.md](terminal/platform-abstraction.md) for I/O override patterns.
6. **Understanding conventions?** Review [practices.md](practices.md) for patterns.

## Key Files in Codebase

| Concept              | Files                                                                                                   |
| -------------------- | ------------------------------------------------------------------------------------------------------- |
| Core prompt engine   | `prompt.go`                                                                                             |
| Prompt components    | `text.go`, `password.go`, `confirm.go`, `select.go`, `multiselect.go`, `autocomplete.go`, `textarea.go` |
| Test support         | `mock.go` (+ `*_test.go` files)                                                                         |
| Styling & symbols    | `ansi_utils.go`, `symbols.go`                                                                           |
| Output utilities     | `messages.go`, `table.go`, `box.go`                                                                     |
| Progress components  | `spinner.go`, `progress.go`, `stream.go`                                                                |
| Terminal abstraction | `internal/terminal/terminal.go`, `internal/terminal/signal_*.go`                                        |
| Type definitions     | `types.go`                                                                                              |
| Examples             | `examples/<component>/main.go`                                                                          |

## Cross-references

- [Architecture](summary.md#architecture) → [Core Flow](summary.md#core-flow) → [Practices](practices.md#event-loop-safety)
- [Textarea component](prompts/prompt-types.md#textarea) → [Input Handling](prompts/input-handling.md) → [Terminal Abstraction](terminal/platform-abstraction.md)
- [Spinner component](progress/indicators.md#spinner) → [Multiline Handling](terminal/platform-abstraction.md#multiline-handling)
- [Testing patterns](practices.md#testing-patterns) → [I/O Override System](terminal/platform-abstraction.md#io-override-system)
