// Package index provides an in-memory task index backed by YAML files on disk.
package index

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/task"
)

// Index holds all tasks in memory with thread-safe access.
type Index struct {
	mu      sync.RWMutex
	dir     string
	tasks   map[string]*task.Task
	cfg     *config.Config
	version atomic.Uint64
}

// New creates and loads an Index from the given tasks directory.
func New(dir string) (*Index, error) {
	idx := &Index{
		dir:   dir,
		tasks: make(map[string]*task.Task),
	}

	if err := idx.Reload(); err != nil {
		return nil, fmt.Errorf("initial load: %w", err)
	}

	return idx, nil
}

// Dir returns the tasks directory path.
func (idx *Index) Dir() string {
	return idx.dir
}

// Reload reloads config and all task files from disk.
func (idx *Index) Reload() error {
	cfg, err := config.Load(idx.dir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	tasks := make(map[string]*task.Task)

	entries, err := os.ReadDir(idx.dir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !isTaskFile(entry.Name()) || entry.Name() == config.ConfigFile {
			continue
		}

		path := filepath.Join(idx.dir, entry.Name())

		t, parseErr := task.ParseFile(path, cfg.Columns)
		if parseErr != nil {
			continue
		}

		tasks[t.ID] = t
	}

	addDanglingRefWarnings(tasks)

	idx.mu.Lock()
	idx.cfg = cfg
	idx.tasks = tasks
	idx.mu.Unlock()
	idx.version.Add(1)

	return nil
}

// ReloadFile re-parses a single task file. If the file was removed, the task is deleted.
func (idx *Index) ReloadFile(path string) {
	name := filepath.Base(path)
	if !isTaskFile(name) {
		return
	}

	idx.mu.RLock()
	cfg := idx.cfg
	idx.mu.RUnlock()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		idx.removeByFilename(name)
		idx.version.Add(1)

		return
	}

	t, err := task.ParseFile(path, cfg.Columns)
	if err != nil {
		return
	}

	idx.mu.Lock()
	idx.tasks[t.ID] = t
	idx.mu.Unlock()
	idx.version.Add(1)
}

// Get returns a task by ID, or nil if not found.
func (idx *Index) Get(id string) *task.Task {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.tasks[id]
}

// List returns all tasks sorted by filename.
func (idx *Index) List() []*task.Task {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := make([]*task.Task, 0, len(idx.tasks))
	for _, t := range idx.tasks {
		result = append(result, t)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].FileName < result[j].FileName
	})

	return result
}

// Config returns the current configuration.
func (idx *Index) Config() *config.Config {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.cfg
}

// Version returns the current data version counter.
func (idx *Index) Version() uint64 {
	return idx.version.Load()
}

func (idx *Index) removeByFilename(name string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for id, t := range idx.tasks {
		if t.FileName == name {
			delete(idx.tasks, id)

			return
		}
	}
}

func addDanglingRefWarnings(tasks map[string]*task.Task) {
	allIDs := make(map[string]bool, len(tasks))
	for id := range tasks {
		allIDs[id] = true
	}

	for _, t := range tasks {
		danglingWarnings := task.ValidateDanglingRefs(t, allIDs)
		if len(danglingWarnings) > 0 {
			t.Warnings = append(t.Warnings, danglingWarnings...)
		}
	}
}

func isTaskFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))

	return ext == ".yaml" || ext == ".yml"
}
