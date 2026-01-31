# DX Pain Points (Agent B)

## High Priority

- Cancel/abort returns zero values with no signal; callers cannot distinguish cancel vs valid zero input. Evidence: `prompt.go:634` finalizes cancel by returning `nil`, wrappers coerce `nil` to zero values (`select.go:91`, `confirm.go:87`, `multiselect_test.go:151`).
- Context cancellation drops `ctx.Err()`. Evidence: `prompt.go:235` and `prompt.go:264` check `ctx.Done()` and return `nil` without returning error; tests assert zero value on cancel (`prompt_test.go:206`).
- Terminal init failure returns zero values without error. Evidence: `prompt.go:658` returns `nil` on terminal init error.

## Medium Priority

- `SelectOptions.MaxItems` exists but unused in `Select`, which is confusing. Evidence: `types.go:50`, `select.go:15` (no usage).
- `MultiSelect` MaxItems silently ignores extra selections with no feedback. Evidence: `multiselect.go:120`.
- Validation only runs on submit; invalid initial value is accepted. Evidence: `prompt.go:370`, `prompt_test.go:667`.

## Low Priority

- Utilities use magic integer status codes on `Stop` without exported constants. Evidence: `progress.go:139`, `spinner.go:132`, `stream.go:91`.
- Global I/O overrides via `SetTermIO` are process-wide and can clash in concurrent apps. Evidence: `types.go:125`, `prompt.go:647`.
