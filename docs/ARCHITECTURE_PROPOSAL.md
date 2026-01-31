# Architecture Proposal: Explicit Prompt Outcomes and Typed Errors

## BLUF

Tap’s prompt APIs currently collapse cancellation, context aborts, and terminal-init failures into zero values, making it impossible for callers to distinguish “user canceled” from valid empty inputs. I propose adding additive, result-returning APIs that surface `ClackState` and `error`, plus light option validation for empty selections. This preserves backward compatibility while materially improving correctness, reliability, and observability.

## Problem Statement (Evidence-Backed)

- `Prompt.Prompt` returns `nil` immediately when `ctx.Done()` is already closed and returns `nil` on cancel finalization, which downstream wrappers coerce into zero values (`""`, `false`, `nil`, zero `T`). This loses information about cancellation and context errors. Evidence: `prompt.go:235-239`, `prompt.go:634-638`, `text.go:112-117`, `confirm.go:87-92`, `select.go:91-98`, `multiselect.go:152-158`.
- Terminal initialization errors (`terminal.New`) return zero values silently, with no error surfaced to callers. Evidence: `prompt.go:658-662`.
- `Select` and `MultiSelect` assume non-empty options and can index `len(options)-1`, which can panic when the options list is empty. Evidence: `select.go:71-88`, `multiselect.go:88-104`.

## Goals

- Surface explicit prompt outcomes (`submit`, `cancel`, `error`) and context/terminal errors to callers without breaking existing APIs.
- Provide an additive path for callers who need to distinguish cancel vs. empty input.
- Add minimal validation for empty selection options in the new APIs.

## Non-Goals

- No breaking change to existing `Text`, `Confirm`, `Select`, etc. signatures.
- No re-architecture of rendering, terminal I/O, or performance changes.
- No behavioral changes to the existing APIs beyond documentation updates.

## Current State (APIs, Invariants, Architecture)

**Public prompt APIs**

- `Text`, `Password`, `Confirm`, `Select`, `MultiSelect`, `Autocomplete` are entry points that call into `Prompt` and return typed values (`string`, `bool`, `T`, `[]T`). Evidence: `text.go:8-117`, `password.go:8-106`, `confirm.go:5-92`, `select.go:15-98`, `multiselect.go:16-158`, `autocomplete.go:9-238`.
- `Prompt` is the core event/render loop; it manages state, validation, and terminal render diffing. Evidence: `prompt.go:16-86`, `prompt.go:321-399`, `prompt.go:570-614`.

**Key invariants**

- `ClackState` drives rendering and outcome transitions (`initial`, `active`, `cancel`, `submit`, `error`). Evidence: `types.go:82-90`, `prompt.go:616-639`.
- On cancel, `Prompt.finalize` emits `cancel` and returns `nil`. Evidence: `prompt.go:620-639`.
- `runWithTerminal` returns zero values if terminal init fails. Evidence: `prompt.go:651-662`.

**High-level architecture**

```
Caller
  │
  ▼
Public APIs (Text/Select/…)
  │
  ▼
Prompt (event loop + render)
  │
  ▼
internal/terminal (Reader/Writer + signals)
```

Evidence: `prompt.go:16-399`, `internal/terminal/terminal.go:44-305`.

## Proposed Change

### 1) Add explicit result types and additive entrypoints

Introduce a small result type and additive APIs that preserve existing behavior while allowing callers to opt into error-aware outcomes.

**API sketch (additive)**

```go
// New result type
// State: StateSubmit, StateCancel, or StateError
// Err: ctx.Err(), ErrTerminalUnavailable, ErrEmptyOptions, or validation error
// Value: populated only on submit or when explicitly set

type PromptResult[T any] struct {
    Value T
    State ClackState
    Err   error
}

func (r PromptResult[T]) Canceled() bool { return r.State == StateCancel }
func (r PromptResult[T]) Submitted() bool { return r.State == StateSubmit }

var (
    ErrTerminalUnavailable = errors.New("tap: terminal unavailable")
    ErrEmptyOptions        = errors.New("tap: empty options")
)

// Additive prompt APIs
func TextResult(ctx context.Context, opts TextOptions) PromptResult[string]
func PasswordResult(ctx context.Context, opts PasswordOptions) PromptResult[string]
func ConfirmResult(ctx context.Context, opts ConfirmOptions) PromptResult[bool]
func SelectResult[T any](ctx context.Context, opts SelectOptions[T]) PromptResult[T]
func MultiSelectResult[T any](ctx context.Context, opts MultiSelectOptions[T]) PromptResult[[]T]
func AutocompleteResult(ctx context.Context, opts AutocompleteOptions) PromptResult[string]
```

**Behavioral rules**

- On cancel (escape/Ctrl+C or context cancellation), return `StateCancel` and `Err = ctx.Err()` when cancellation is context-driven. Evidence that cancel is currently silent: `prompt.go:235-239`, `prompt.go:634-638`.
- On terminal init failure, return `StateError` with `ErrTerminalUnavailable`. Evidence for silent zero values: `prompt.go:658-662`.
- For `SelectResult`/`MultiSelectResult`, if `len(opts.Options) == 0`, return `StateError` with `ErrEmptyOptions` to avoid panics. Evidence of current panic risk: `select.go:71-88`, `multiselect.go:88-104`.
- Validation errors continue using `ValidationError` for user feedback, but are surfaced in `PromptResult.Err` when in `StateError`. Evidence: `prompt.go:380-393`, `types.go:94-103`.

### 2) New internal helper to preserve behavior

- Introduce internal helpers (e.g., `runWithTerminalResult`) used only by the new result APIs. Existing APIs remain unchanged and continue returning zero values on cancel/error to preserve compatibility.
- `Prompt` gains a `Result(ctx)` method returning `PromptResult[any]`, while `Prompt.Prompt` remains unchanged and can be a thin wrapper. Evidence of current `Prompt.Prompt` return type: `prompt.go:234-286`.

### 3) Documentation additions

- Document the new result APIs and the error semantics in README and examples; add a result example to `examples/` to demonstrate cancellation handling.

## Alternatives Considered

1. **Change existing APIs to return `(T, error)`** — Rejected because it breaks all call sites and violates library backward compatibility expectations. Evidence of widespread public API usage: `README.md:89-115` and `examples/*/main.go`.
2. **Keep status quo and only document cancellation behavior** — Rejected because the behavior is observable but still forces ambiguous zero-value handling; it does not resolve the panic risk on empty options. Evidence: `prompt.go:634-638`, `select.go:71-88`.
3. **Emit cancel/error via global hooks** — Rejected because it adds global state and does not scale in concurrent callers; the repo already has global I/O overrides that can clash (`types.go:125-129`, `prompt.go:647-649`).

## Risks & Mitigations

- **Risk: API surface expansion increases maintenance** → Mitigate by keeping a single `PromptResult` type and implementing result wrappers in each prompt file; no changes to existing API signatures.
- **Risk: Mixed usage creates confusion** → Mitigate with clear README section “Result APIs (opt‑in)” and examples showing both patterns.
- **Risk: Additional errors could be misinterpreted** → Mitigate by using Go standard `context` errors (return `ctx.Err()`), and only two new sentinel errors.

## Rollout Plan (Phased)

**Phase 1 (v0.x patch)**

- Add `PromptResult[T]`, `ErrTerminalUnavailable`, `ErrEmptyOptions`.
- Add `Prompt.Result(ctx)` and new `*Result` entrypoints for each prompt.
- Update README and add one example demonstrating cancel/error handling.

**Phase 2 (v0.x patch)**

- Add option-validation checks inside result APIs for `Select`/`MultiSelect` to prevent empty options from panic.
- Add tests for result APIs (cancel, ctx timeout, empty options, terminal error).

**Phase 3 (v0.x minor)**

- Encourage adoption by updating examples to show `*Result` variants in at least one flow (keep original examples intact).

**Success metrics**

- ≥90% of new tests for result APIs cover cancel/ctx/terminal failure paths.
- Zero changes in existing API test behavior.
- No new lints; existing lints remain unchanged unless addressed separately.

## Testing Strategy

- Add unit tests for `*Result` functions using `MockReadable/MockWritable` and `SetTermIO` (per existing patterns). Evidence: `mock.go:8-86`, `INSTRUCTIONS.md:73-91`.
- Add tests for empty options returning `ErrEmptyOptions` (new behavior only in result APIs).
- Add tests for context cancellation returning `ctx.Err()` and `StateCancel`.
- Baseline tests currently pass (`go test ./...`, `go test -race ./...`). Lint currently fails with `wsl_v5` warnings in `box.go` and `table.go` (not changed by this proposal). Evidence: lint output; files `box.go:263`, `table.go:368`, `table.go:409`.

## Operational Concerns

- **Versioning**: Additive APIs only; no breaking changes. Safe for patch/minor release in `v0.x`.
- **Compatibility**: Existing functions remain unchanged and continue returning zero values on cancel/error.
- **Performance**: No performance work proposed; no benchmarks required. (Benchmarks currently absent: no `Benchmark*` functions found.)
- **Security/Reliability**: New empty-options validation prevents potential panics; explicit errors improve failure visibility. Evidence: `select.go:71-88`, `multiselect.go:88-104`.
- **Docs**: README and a new example should document the result APIs and cancel semantics.

## Open Questions / Unknowns

- Is there a preferred naming scheme (`TextResult` vs. `TextWithResult`)? No evidence of naming conventions beyond current functions (`Text`, `Password`, `Confirm`).
- Should we propagate `ValidationError` as-is or wrap it? Current behavior uses `ValidationError` in the prompt loop; error strategy in result APIs is a design choice. Evidence: `prompt.go:380-393`, `types.go:94-103`.
- Should result APIs also cover utility components (Spinner/Progress/Stream)? No evidence of existing error reporting patterns for utilities beyond `Stop(code int)`.

## Appendix: Evidence Anchors

- Module/version: `go.mod:1-3`.
- Public prompt APIs: `text.go:8-117`, `password.go:8-106`, `confirm.go:5-92`, `select.go:15-98`, `multiselect.go:16-158`, `autocomplete.go:9-238`.
- Prompt loop and cancellation: `prompt.go:234-286`, `prompt.go:616-639`.
- Terminal init failure fallback: `prompt.go:651-662`.
- `ClackState` and `ValidationError`: `types.go:82-103`.
- Empty options cursor logic: `select.go:71-88`, `multiselect.go:88-104`.
- Mock testing utilities: `mock.go:8-86`.
- Lint failures: `box.go:263`, `table.go:368`, `table.go:409`.
