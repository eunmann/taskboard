package index

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestIsRelevantEvent(t *testing.T) {
	tests := []struct {
		name string
		op   fsnotify.Op
		want bool
	}{
		{"write", fsnotify.Write, true},
		{"create", fsnotify.Create, true},
		{"remove", fsnotify.Remove, true},
		{"rename", fsnotify.Rename, true},
		{"chmod", fsnotify.Chmod, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := fsnotify.Event{Name: "test.yaml", Op: tt.op}
			if got := isRelevantEvent(event); got != tt.want {
				t.Errorf("isRelevantEvent(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestNewWatcherAndClose(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	w, err := NewWatcher(idx)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	if err := w.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestWatcherDetectsFileChange(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	initialVersion := idx.Version()

	w, err := NewWatcher(idx)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	}()

	ctx := t.Context()

	go w.Start(ctx)

	// Modify an existing task file to trigger the watcher.
	writeFile(t, dir, "abc12345.yaml", "title: Modified\nstatus: open\n")

	// Wait for debounce to complete.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if idx.Version() > initialVersion {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if idx.Version() <= initialVersion {
		t.Error("watcher did not update index version after file change")
	}

	task := idx.Get("abc12345")
	if task == nil {
		t.Fatal("task disappeared after watcher update")
	}

	if task.Title != "Modified" {
		t.Errorf("task title = %q, want %q", task.Title, "Modified")
	}
}

func TestWatcherDetectsNewFile(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	initialCount := len(idx.List())
	initialVersion := idx.Version()

	w, err := NewWatcher(idx)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	}()

	ctx := t.Context()

	go w.Start(ctx)

	// Create a new task file.
	writeFile(t, dir, "new12345.yaml", "title: New task\nstatus: open\n")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if idx.Version() > initialVersion {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if len(idx.List()) <= initialCount {
		t.Error("watcher did not detect new file")
	}
}

func TestWatcherDetectsDeletedFile(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	initialVersion := idx.Version()

	w, err := NewWatcher(idx)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	}()

	ctx := t.Context()

	go w.Start(ctx)

	if err := os.Remove(filepath.Join(dir, "abc12345.yaml")); err != nil {
		t.Fatalf("remove file: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if idx.Version() > initialVersion {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if idx.Get("abc12345") != nil {
		t.Error("watcher did not detect deleted file")
	}
}

func TestWatcherConfigChangeRebuilds(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	initialVersion := idx.Version()

	w, err := NewWatcher(idx)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	}()

	ctx := t.Context()

	go w.Start(ctx)

	// Rewrite config to trigger a full rebuild.
	writeFile(t, dir, "config.yaml", `
project: Updated Project
columns:
  status:
    order: 1
    values:
      - name: open
        color: green
      - name: closed
        color: red
`)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if idx.Version() > initialVersion {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	cfg := idx.Config()
	if cfg.Project != "Updated Project" {
		t.Errorf("config.Project = %q, want %q", cfg.Project, "Updated Project")
	}
}

func TestWatcherStartStopsOnCancel(t *testing.T) {
	dir := setupTestDir(t)

	idx, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	w, err := NewWatcher(idx)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		w.Start(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Start returned as expected.
	case <-time.After(2 * time.Second):
		t.Error("Start did not return after context cancellation")
	}
}
