# Architecture

## Overview

Taskboard is a three-layer system:

1. **YAML files on disk** — tasks and configuration live in `.taskboard/` as plain files
2. **In-memory index** — loaded at startup, kept current via filesystem watching
3. **HTTP server + templates** — stdlib `net/http` serves HTML pages with embedded static assets

There is no database. Git provides version history. The binary is self-contained — all templates, CSS, and JavaScript are embedded via `go:embed`.

## Directory Structure

```
cmd/
  taskboard/         → main entry point, CLI flag parsing, server startup
  seed/              → generates sample task YAML files from templates

internal/
  config/            → YAML config loading and validation (Config, Column, Value structs)
  index/             → in-memory task index, file watcher, search
  scaffold/          → --init scaffolding: creates .taskboard/, config, AI commands
    commands/        → Claude Code slash command templates (plan-work, next-task, task-status)
  seed/              → sample task YAML templates used by cmd/seed
    samples/         → individual sample task YAML files
  server/            → HTTP handlers, routing, template rendering, graph
  task/              → Task model, YAML parsing, validation, ID generation, refs
  web/               → embedded static assets and HTML templates
    static/          → CSS, JS, favicon
    templates/       → Go html/template files
      layout.html    → base layout (head, nav, body wrapper)
      pages/         → full page templates (list, detail, graph)
      partials/      → reusable content blocks (table rows, detail content)
      components/    → small UI elements (badge, warning, filter, refs)
```

## Data Model

### Task

The `Task` struct (`internal/task/task.go`) holds both fixed fields and a flexible field map:

```go
type Task struct {
    ID          string
    Title       string
    Description string
    Tags        []string
    Refs        []Ref
    Fields      map[string]string   // custom columns: status, priority, type, etc.
    Created     time.Time
    Updated     time.Time
    Warnings    []Warning
    FileName    string
    SkippedRefs int
}
```

Core fields (`id`, `title`, `description`, `tags`, `refs`, `created`, `updated`) are parsed into typed struct fields. Everything else lands in `Fields` and is validated against column definitions in the config.

### References

Three reference types connect tasks:

| Type | Constant | Meaning |
|------|----------|---------|
| `parent` | `RefParent` | Groups subtask under a parent (max one per task) |
| `blocked-by` | `RefBlockedBy` | Dependency — this task waits on the target |
| `relates-to` | `RefRelatesTo` | Informational link |

Reverse references are computed at index time. If task A declares `parent: B`, then B's detail page lists A as a child without B needing to declare anything.

### Config

`config.Config` (`internal/config/config.go`) defines:

- **Project name** — displayed in the UI header
- **Columns** — ordered map of column definitions, each with a name, display order, and list of allowed values with hex colors

Reserved field names (`id`, `title`, `description`, `tags`, `refs`, `created`, `updated`) cannot be used as column names.

### Validation

Validation runs in two phases:

1. **Parse-time** (`task.Validate()`) — checks a single task in isolation:
   - Required fields: `title`, `created`, `updated`
   - ID/filename consistency (prefix before first `-` must match `id` field)
   - Field values must be defined in the config's column list
   - Ref structure: max one `parent`, warns on malformed refs

2. **Index-time** (`task.ValidateDanglingRefs()`) — requires the full task set:
   - Every ref target ID must exist in the index
   - Missing targets produce warnings (not errors)

The system prefers warnings over hard errors, so tasks with issues still appear in the UI with inline warning indicators.

## Index

The in-memory index (`internal/index/index.go`) is the central data store:

```go
type Index struct {
    mu      sync.RWMutex
    dir     string
    tasks   map[string]*task.Task
    cfg     *config.Config
    version atomic.Uint64
}
```

- **Thread safety** — read operations take `RLock`, writes take full `Lock`
- **Version counter** — `atomic.Uint64` incremented on every reload, used for ETag-based caching
- **`Reload()`** — full reload: re-reads config and all task files from disk
- **`ReloadFile(path)`** — incremental: re-parses a single changed file without touching the rest
- **Dangling ref detection** — after each reload, `addDanglingRefWarnings()` scans all refs and adds warnings for targets that don't exist

## File Watching

The watcher (`internal/index/watcher.go`) uses `fsnotify` to monitor `.taskboard/`:

- **Events** — write, create, remove, rename
- **Debounce** — 100ms delay accumulates rapid changes into a single reload
- **Config vs task** — changes to `config.yaml` trigger a full `Reload()` (since column definitions may have changed); changes to task files trigger incremental `ReloadFile()`
- **State tracking** — a `debounceState` struct holds a map of pending file paths and a rebuild flag, protected by a mutex

## HTTP Server

### Routes

Five routes registered on a stdlib `http.ServeMux` (`internal/server/server.go`):

| Pattern | Handler | Description |
|---------|---------|-------------|
| `GET /{$}` | `handleList` | Task list with search, filters, sort |
| `GET /task/{id}` | `handleDetail` | Single task detail page |
| `GET /graph` | `handleGraph` | Dependency graph (Mermaid) |
| `GET /partials/table` | `handleTablePartial` | HTMX table partial for polling |
| `GET /static/` | `http.FileServer` | Embedded static assets |

### HTMX partial vs full page

Every page handler checks for the `HX-Request` header. Full page requests render the complete layout; HTMX requests return only the relevant partial, avoiding redundant re-rendering of the shell.

### Graceful shutdown

The server listens for `SIGINT` and `SIGTERM`, then calls `http.Server.Shutdown()` with a context deadline to drain in-flight requests.

## Templates

### Hierarchy

```
layout.html              → base HTML: <head>, nav bar, content slot
  pages/list.html         → task table with filters and search
  pages/detail.html       → single task view with refs and description
  pages/graph.html        → Mermaid graph container
partials/table.html       → table body rows (used by both list page and HTMX partial)
partials/detail.html      → detail content block
components/badge.html     → colored status/field badges
components/warning.html   → validation warning display
components/filter.html    → filter dropdown UI
components/refs.html      → reference link list
```

All templates are parsed together with a shared `FuncMap`.

### Template Functions

Key functions available in templates (`internal/server/render.go`):

| Function | Purpose |
|----------|---------|
| `formatTime` | Formats `time.Time` as "Jan 2, 2006 3:04 PM" |
| `relativeTime` | Converts to relative format ("2 hours ago") |
| `markdown` | Renders Markdown via goldmark |
| `dict` | Creates a map from key-value pairs for passing to sub-templates |
| `sortParam` | Builds query parameter for column sorting |
| `sortIndicator` | Renders sort direction arrow |
| `inSlice` | Checks slice membership |
| `queryParams` | Builds query string from current filters |
| `lower` / `upper` | Case conversion |
| `hasPrefix` | String prefix check |
| `joinComma` | Joins string slice with commas |

## Frontend

### HTMX

- **Navigation** — link clicks fetch pages via HTMX, swapping the content area
- **Polling** — the table partial is polled on an interval; 304 responses (via ETag) prevent unnecessary re-renders
- **Search** — input triggers HTMX requests with debounced keystrokes

### Alpine.js

- **Theme toggle** — dark/light mode switch with `localStorage` persistence
- **Filter dropdowns** — multi-select filter panels for each column
- **Density toggle** — compact/comfortable table row spacing

### Tailwind CSS

- **Responsive** — mobile card layout at narrow breakpoints, full table on wider screens
- **Dark mode** — class-based strategy (`dark:` variants), toggled by Alpine.js

### Mermaid

- **Graph** — server generates a Mermaid graph definition string; the client renders it
- **Node styling** — fill colors from status column config, text color computed for contrast (ITU-R BT.601 luminance)
- **Edge styles** — solid arrows for parent, thick arrows for blocked-by, dashed for relates-to
- **Theme-aware** — graph re-renders on theme change to match dark/light mode

## Caching

The server uses ETag-based caching to minimize data transfer during HTMX polling:

1. **ETag generation** — the index's atomic version counter combined with the query string produces a unique ETag per response state
2. **Conditional requests** — the client stores the ETag in a `<meta>` tag and sends `If-None-Match` on subsequent polls
3. **304 responses** — if the ETag matches, the server returns `304 Not Modified` with no body

This means polling requests are essentially free when no files have changed.

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **No database** | Git already provides version control; YAML files are human-readable and diff-friendly |
| **Read-only UI** | Tasks are meant to be edited in your text editor or by AI agents, not through a web form |
| **In-memory index** | Fast enough for typical project sizes (hundreds of tasks); avoids query language complexity |
| **Warnings over errors** | Graceful degradation — a task with a broken ref still appears in the UI with a warning badge |
| **Embedded assets** | Single binary deployment; `go:embed` bundles templates, CSS, and JS at compile time |
| **Stdlib router** | `net/http.ServeMux` covers the five routes without adding a dependency |
| **100ms debounce** | Batches rapid file saves (e.g., from `git checkout`) into one reload |
