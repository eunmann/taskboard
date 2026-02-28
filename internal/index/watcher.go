package index

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/fsnotify/fsnotify"
)

const debounceDelay = 100 * time.Millisecond

// Watcher watches the tasks directory for changes and updates the index.
type Watcher struct {
	idx     *Index
	watcher *fsnotify.Watcher
}

// NewWatcher creates a filesystem watcher for the index's directory.
func NewWatcher(idx *Index) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	if err := w.Add(idx.Dir()); err != nil {
		_ = w.Close()

		return nil, fmt.Errorf("watch directory: %w", err)
	}

	return &Watcher{idx: idx, watcher: w}, nil
}

// debounceState tracks pending file changes for debounced processing.
type debounceState struct {
	mu      sync.Mutex
	timer   *time.Timer
	pending map[string]bool
	rebuild bool
}

// Start begins watching for file changes. It blocks until ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) {
	state := &debounceState{
		pending: make(map[string]bool),
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			w.handleEvent(event, state)

		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	err := w.watcher.Close()
	if err != nil {
		return fmt.Errorf("close watcher: %w", err)
	}

	return nil
}

func (w *Watcher) handleEvent(event fsnotify.Event, state *debounceState) {
	if !isRelevantEvent(event) {
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if filepath.Base(event.Name) == config.ConfigFile {
		state.rebuild = true
	} else {
		state.pending[event.Name] = true
	}

	if state.timer != nil {
		state.timer.Stop()
	}

	state.timer = time.AfterFunc(debounceDelay, func() {
		w.processPending(state)
	})
}

func (w *Watcher) processPending(state *debounceState) {
	state.mu.Lock()
	doRebuild := state.rebuild

	files := make([]string, 0, len(state.pending))
	for f := range state.pending {
		files = append(files, f)
	}

	state.pending = make(map[string]bool)
	state.rebuild = false
	state.mu.Unlock()

	if doRebuild {
		_ = w.idx.Reload()
	} else {
		for _, f := range files {
			w.idx.ReloadFile(f)
		}
	}
}

func isRelevantEvent(event fsnotify.Event) bool {
	return event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0
}
