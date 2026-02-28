## Slash Commands

- `/workflow` — branch creation, incremental change process, PR steps.
- `/go-standards` — full Go code quality standards (12 subsections).
- `/testing` — testing conventions, lint rules, make targets.
- `/checklist` — pre-commit code generation checklist with artifact status.
- `/tdd` — red-green-refactor test-driven development cycle.
- `/review-branch` — review all branch changes, find gaps, fix them.

## Project-Specific Standards

- Maximum 70 lines per function. `funlen` lint limit is 80.
- Standard `testing` package only — no testify. Use `if` + `t.Errorf`/`t.Fatalf`.
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`.
- `gosec G602`: avoid `arr[variable]` — use direct conditional returns.

## Debugging

- `make test`, `make test-all`, and `make lint` persist output to `artifacts/test.log`, `artifacts/test-all.log`, and `artifacts/lint.log` respectively.
- Grep these files to check results without re-running: e.g., `grep "FAIL" artifacts/test.log` or `grep "error" artifacts/lint.log`.
- The files are overwritten on each run, so they always reflect the latest results.

## Rules

- Never push broken builds.
- No TODOs unless explicitly approved.
- No behavior changes without tests.
- Fix lint errors by improving the code — never use `//nolint` unless it is provably the only option, with an inline comment explaining why.
- All code must be formatted. No exceptions. `make lint` applies formatting.
- All code must pass linting with zero warnings. Zero means zero — restructure code to eliminate false positives rather than suppressing.
