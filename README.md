# Taskboard

A repo-native task tracker. Tasks are flat YAML files in `.tasks/`, served through a read-only Go web UI with HTMX live-updating.

## Quick Start

```bash
# Initialize a .tasks/ directory with default config
./bin/taskboard --init

# Start the web UI (default: http://localhost:9746)
./bin/taskboard
```

## Building

```bash
make build    # Build binary to ./bin/taskboard
make test     # Run tests
make lint     # Run linter
```

## Task Files

Tasks live as individual YAML files in `.tasks/`. Each file defines one task:

```yaml
title: Implement user authentication
status: in-progress
priority: high
tags: [backend, security]
description: |
  Add JWT-based authentication to all API endpoints.
```

Column definitions (like `status` and `priority`) are configured in `.tasks/config.yaml`.

## Configuration

`.tasks/config.yaml` defines the project and available columns:

```yaml
project: My Project
columns:
  status:
    order: 1
    values:
      - name: backlog
        color: gray
      - name: in-progress
        color: blue
      - name: done
        color: green
  priority:
    order: 2
    values:
      - name: low
        color: gray
      - name: medium
        color: yellow
      - name: high
        color: red
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `.tasks` | Path to tasks directory |
| `--port` | `9746` | HTTP server port |
| `--init` | | Initialize tasks directory and exit |
| `--version` | | Print version and exit |

## Tech Stack

- **Backend**: Go 1.25, `net/http` (no framework)
- **Frontend**: HTMX, Alpine.js, Tailwind CSS
- **Task Storage**: YAML files on disk
- **Live Updates**: fsnotify file watching + HTMX polling
