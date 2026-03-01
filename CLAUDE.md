## Slash Commands

- `/workflow` — branch creation, incremental change process, PR steps.
- `/go-standards` — full Go code quality standards (12 subsections).
- `/testing` — testing conventions, lint rules, make targets.
- `/checklist` — pre-commit code generation checklist with artifact status.
- `/tdd` — red-green-refactor test-driven development cycle.
- `/review-branch` — review all branch changes, find gaps, fix them.
- `/verify-task` — per-task implementation verification procedure.

## Architecture

```
cmd/taskboard/   → main entry point, CLI flags, server startup
cmd/seed/        → sample task data generator
internal/config/ → YAML config loading and validation
internal/index/  → full-text search indexing
internal/seed/   → sample task YAML templates
internal/server/ → HTTP handlers, routing, middleware (stdlib net/http)
internal/task/   → task model, YAML file operations, validation
internal/web/    → HTML templates, static assets (CSS/JS)
```

- No database — tasks are YAML files on disk in `.taskboard/`.
- No external router — stdlib `net/http.ServeMux` only.
- Templates use Go `html/template` with partials.

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

## Execution Loop

When implementing a task, follow this loop:

1. **Decompose** — break the task into small, testable steps. Create a `/verify-task` checklist.
2. **Implement** — one step at a time. Write tests first (TDD) when adding behavior.
3. **Verify** — after each step: `make test`, `make lint`. Check artifact logs.
4. **Commit** — atomic commit per step. Clear message describing "why".
5. **Repeat** — next step until done.
6. **Final check** — `make test-all`, `make lint`, review all changes with `/review-branch`.

## Verification Task Template

When decomposing work, create a verification checklist:

```
- [ ] Tests pass: `make test` → check artifacts/test.log
- [ ] Lint passes: `make lint` → check artifacts/lint.log
- [ ] No unintended changes: `git diff` is clean
- [ ] Commit message describes the "why"
```

## Parallel Work

- **Reads are safe to parallelize**: file reads, grep, glob, git log.
- **Writes must be serialized**: file edits, git commits, make targets.
- Use subagents for independent research tasks (e.g., reading multiple files).
- Never run two `make` targets concurrently — they share artifact files.

## Error Recovery

- **Test failure**: read `artifacts/test.log`, find the failing test, fix the code (not the test unless the test is wrong).
- **Lint failure**: read `artifacts/lint.log`, fix the code to satisfy the linter. Never suppress with `//nolint`.
- **Build failure**: read the error, fix the code. Check `go.mod` if import issues.
- **Blocked**: if stuck, explain the blocker clearly and ask the user for guidance.
<!-- taskboard:begin -->

## Taskboard

Tasks are YAML files in `.taskboard/`. Use these commands to plan and execute work:

- `/plan-work` — Decompose an objective into 3-8 taskboard tasks with dependencies.
- `/next-task` — Pick the next unblocked task, implement it, mark it done.
- `/task-status` — Show overview: status counts, available/blocked tasks, next actions.

### Task YAML Schema

```yaml
id: Xr8m2v
title: "Short imperative description"
status: open
priority: medium
type: feature
tags: [area]
created: 2025-01-15T10:00:00Z
updated: 2025-01-15T10:00:00Z
refs:
  - type: blocked-by
    id: Hv5n8r
description: |
  What needs to be done and acceptance criteria.
```

### Status Values

`open` | `in-progress` | `review` | `done` | `cancelled`

### Ref Types

- `blocked-by` — this task cannot start until the referenced task is done.
- `parent` — groups subtasks under a larger task.
- `relates-to` — informational cross-reference.

### Workflow

1. `/plan-work` — break the objective into tasks.
2. `/next-task` — pick and complete one task at a time.
3. `/task-status` — check progress, find blockers.
4. Repeat until done.

### Quick Access

Use Grep to search task files — don't read every file when you only need specific info.

- All statuses at a glance: `grep "^status:" .taskboard/*.yaml`
- Find open tasks: `grep -l "^status: open" .taskboard/*.yaml`
- Find blocked tasks: `grep -l "blocked-by" .taskboard/*.yaml`
- All titles: `grep "^title:" .taskboard/*.yaml`
- Count by status: `grep -c "^status: open" .taskboard/*.yaml` (per-file), or `grep -rl "^status: open" .taskboard/*.yaml | wc -l` (total)

<!-- taskboard:end -->
