---
description: Show task overview with status counts, blockers, and next actions
---

# Task Status

Gather task data using Grep before reading individual files.

1. Run `grep "^status:" .taskboard/*.yaml` to get all statuses in one pass.
2. Run `grep "^title:" .taskboard/*.yaml` to get all titles.
3. Only `Read` individual files when you need full details (e.g., to check `blocked-by` refs or priority).

## Sections

### Status Counts

Show counts for each status value:

```
open: 4 | in-progress: 1 | review: 0 | done: 3 | cancelled: 0
```

### Available Tasks

Tasks where `status: open` and all `blocked-by` refs are `done` or `cancelled`. Sort by priority, then by `created`.

```
ID       | Title                        | Priority
---------|------------------------------|--------
Xr8m2v   | Add widget component         | high
Nf3a5w   | Update documentation         | low
```

### In-Progress Tasks

Tasks where `status: in-progress`:

```
ID       | Title                        | Priority
---------|------------------------------|--------
Hv5n8r   | Review error handling        | medium
```

### Blocked Tasks

Tasks where `status: open` but one or more `blocked-by` refs are not `done`/`cancelled`. Show which specific tasks are blocking:

```
ID       | Title                        | Blocked By
---------|------------------------------|----------
Sd2k7x   | Deploy to staging            | Hv5n8r (in-progress)
```

## Suggested Next Action

- If there are available tasks: "Run `/next-task` to pick up the next task."
- If all remaining tasks are blocked: "All open tasks are blocked. Complete in-progress work first."
- If everything is done: "All tasks are complete!"
