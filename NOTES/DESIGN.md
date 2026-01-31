# Design Outline (Agent E)

## Proposed Theme

Introduce a first-class theming system so apps can customize symbols, colors, and borders consistently across prompts, utilities, and message helpers. Evidence of hard-coded styling: `symbols.go:4`, `symbols.go:34`, `select.go:122`, `spinner.go:163`, `messages.go:28`.

## Additive API Sketch

- New `Theme` type (symbols + styles + colorizer) with helpers `DefaultTheme()` and `NoColorTheme()`.
- Add `Theme *Theme` to option structs across prompts/utilities and `MessageOptions` (optionally `TableOptions`).
- Package-level `SetTheme(theme Theme)` / `Theme()` accessors for global defaults.

## Migration Plan

- Phase 1: introduce `Theme` + defaults mirroring current symbols/colors; keep `Symbol` and color helpers delegating to default theme.
- Phase 2: wire components to prefer `opts.Theme` -> global theme -> default theme; update tests to use default theme or `NoColorTheme()`.
- Phase 3: document theming + add examples; no breaking change.
