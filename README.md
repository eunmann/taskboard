# Taskboard

A repo-native task tracker that lives alongside your code. Tasks are plain YAML files in `.taskboard/`, tracked by Git like everything else, and served through a read-only web UI that updates in real time.

## Features

- **List, detail, and graph views** — browse tasks in a sortable table, drill into details, or visualize dependencies as a graph
- **Search, filters, and sort** — full-text search across all fields with multi-column filtering
- **Task references** — `parent`, `blocked-by`, and `relates-to` links between tasks with automatic reverse-ref computation
- **Live updates** — file system watching with HTMX polling; edits appear without refreshing
- **Dark mode** — system-aware theme toggle with persistent preference
- **Mobile responsive** — card layout on narrow screens, full table on desktop
- **Keyboard navigation** — `j`/`k` movement, `/` to search, `Enter` to open
- **Validation warnings** — malformed refs, unknown field values, and missing fields surfaced inline
- **Configurable columns** — define your own status, priority, type (or any column) with custom values and colors

## Quick Start

```bash
# Option 1: go run (no install)
go run github.com/eunmann/taskboard/cmd/taskboard@latest --init
go run github.com/eunmann/taskboard/cmd/taskboard@latest

# Option 2: go install
go install github.com/eunmann/taskboard/cmd/taskboard@latest
taskboard --init
taskboard

# Option 3: manual build
make build                    # Build binary to ./bin/taskboard
./bin/taskboard --init        # Create .taskboard/ with default config
./bin/taskboard               # Open http://localhost:9746
```

`--init` creates a `.taskboard/` directory containing `config.yaml` with default columns and scaffolds [Claude Code](https://claude.com/claude-code) slash commands for AI-assisted task management (see [AI Integration](#ai-integration)).

## Task Files

Each task is a single YAML file in `.taskboard/`. Filenames follow the pattern `<id>-<slug>.yaml`, where the ID is a 6-character alphanumeric string.

```yaml
id: Xr8m2v
title: "Add keyboard navigation shortcuts"
status: open
priority: high
type: feature
tags: [ui, accessibility]
created: 2025-01-14T10:00:00Z
updated: 2025-01-14T10:00:00Z
refs:
  - type: parent
    id: Nf3a5w
  - type: relates-to
    id: Wc4g6t
description: |
  Add keyboard shortcuts for common task list operations.

  ## Shortcuts

  - `j`/`k` to move up/down in the task list
  - `Enter` to open selected task
  - `/` to focus search input
  - `Escape` to clear search or go back
```

**Core fields:** `id`, `title`, `description`, `tags`, `refs`, `created`, `updated`

All other fields (like `status`, `priority`, `type` above) are custom columns defined in `config.yaml` and stored in a flexible fields map.

## References

Tasks can reference each other with typed links:

| Type | Meaning |
|------|---------|
| `parent` | Groups this task under a larger task (max one per task) |
| `blocked-by` | This task cannot start until the referenced task is done |
| `relates-to` | Informational cross-reference |

Reverse references are computed automatically — if task A has `parent: B`, then B's detail page shows A as a child. The graph view renders all references as a directed graph with distinct edge styles for each type.

## Configuration

`.taskboard/config.yaml` defines the project name and columns with ordered values and hex colors:

```yaml
project: Taskboard
columns:
  status:
    order: 1
    values:
      - name: open
        color: "#22c55e"
      - name: in-progress
        color: "#3b82f6"
      - name: review
        color: "#f59e0b"
      - name: done
        color: "#6b7280"
      - name: cancelled
        color: "#ef4444"
  priority:
    order: 2
    values:
      - name: critical
        color: "#dc2626"
      - name: high
        color: "#f97316"
      - name: medium
        color: "#eab308"
      - name: low
        color: "#6b7280"
  type:
    order: 3
    values:
      - name: feature
        color: "#8b5cf6"
      - name: fix
        color: "#ef4444"
      - name: chore
        color: "#64748b"
      - name: spike
        color: "#06b6d4"
      - name: docs
        color: "#22d3ee"
      - name: refactor
        color: "#f59e0b"
```

Colors are used for badges in the list view and node fills in the graph. Add your own columns by adding new entries under `columns:`.

## CLI

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `.taskboard` | Path to tasks directory |
| `--port` | `9746` | HTTP server port |
| `--init` | — | Initialize tasks directory and exit |
| `--version` | — | Print version and exit |

## AI Integration

When `--init` detects a `.claude/` directory (indicating [Claude Code](https://claude.com/claude-code) is in use), it scaffolds three slash commands:

| Command | Purpose |
|---------|---------|
| `/plan-work` | Decompose an objective into 3–8 taskboard tasks with dependencies |
| `/next-task` | Pick the next unblocked task, implement it, and mark it done |
| `/task-status` | Show project overview with status counts and blocked/available tasks |

These commands are written to `.claude/commands/` and a reference section is appended to `CLAUDE.md`.

## Tech Stack

- **Go 1.26** — stdlib `net/http`, no framework, no database
- **HTMX** — navigation, polling, partial page updates
- **Alpine.js** — theme toggle, filter dropdowns, density toggle
- **Tailwind CSS** — responsive layout, dark mode via class strategy
- **Mermaid** — dependency graph rendering
- **goldmark** — Markdown rendering in task descriptions
- **fsnotify** — file system watching with 100ms debounce

## Development

```bash
make build    # Build binary to ./bin/taskboard
make dev      # Init sample data + hot reload via air
make test     # Run tests (results in artifacts/test.log)
make lint     # Run linter with auto-fix (results in artifacts/lint.log)
make seed     # Generate sample task files in .taskboard/
```
