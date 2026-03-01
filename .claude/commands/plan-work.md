---
description: Decompose an objective into taskboard tasks with dependencies
---

# Plan Work

Break a large objective into 3-8 right-sized taskboard tasks.

## Process

1. **Understand the objective** — read the user's description carefully.
2. **Check existing tasks** — use `grep "^title:" .taskboard/*.yaml` to see existing work and avoid duplicates. Use `grep "^id:" .taskboard/*.yaml` to avoid ID collisions.
3. **Explore the codebase** — identify affected files, packages, and patterns.
4. **Decompose** — split the work into discrete tasks, each representing 30-90 minutes of focused AI work.
5. **Define dependencies** — use `blocked-by` refs so tasks can be executed in order.
6. **Create task files** — write YAML files to `.taskboard/`.

## Task File Format

Each task is a YAML file in `.taskboard/` named `<id>-<slug>.yaml`:

```yaml
id: Xr8m2v
title: "Short imperative description"
status: open
priority: medium
type: feature
tags: [area-one, area-two]
created: 2025-01-15T10:00:00Z
updated: 2025-01-15T10:00:00Z
refs:
  - type: blocked-by
    id: Hv5n8r
  - type: relates-to
    id: Nf3a5w
description: |
  Detailed description of what needs to be done.

  ## Acceptance Criteria

  - Criterion one
  - Criterion two
```

## ID Generation

- 6 characters, alphanumeric.
- Exclude visually ambiguous characters: `0`, `O`, `o`, `1`, `l`, `L`, `I`, `i`.
- Valid alphabet: `23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ`

## Rules

- Every task must have `status: open` and a `created` timestamp.
- Use `blocked-by` refs to express ordering constraints.
- Use `parent` refs to group subtasks under a larger task.
- Use `relates-to` for informational cross-references.
- Priority values: `critical`, `high`, `medium`, `low`.
- Type values: `feature`, `fix`, `chore`, `spike`, `docs`, `refactor`.
- Each task should be completable independently once its blockers are resolved.

## Output

After creating all tasks, print a summary table:

```
ID       | Title                        | Priority | Blocked By
---------|------------------------------|----------|----------
Xr8m2v   | Add widget component         | high     | —
Hv5n8r   | Write widget tests           | high     | Xr8m2v
```

Then suggest: "Run `/next-task` to start working."
