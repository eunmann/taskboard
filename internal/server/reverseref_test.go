package server

import (
	"testing"

	"github.com/eunmann/taskboard/internal/task"
)

func TestReverseLabelParent(t *testing.T) {
	if got := reverseLabel(task.RefParent); got != reverseLabelChild {
		t.Errorf("reverseLabel(%q) = %q, want %q", task.RefParent, got, reverseLabelChild)
	}
}

func TestReverseLabelBlockedBy(t *testing.T) {
	if got := reverseLabel(task.RefBlockedBy); got != reverseLabelBlocks {
		t.Errorf("reverseLabel(%q) = %q, want %q", task.RefBlockedBy, got, reverseLabelBlocks)
	}
}

func TestReverseLabelRelatesTo(t *testing.T) {
	if got := reverseLabel(task.RefRelatesTo); got != reverseLabelRelated {
		t.Errorf("reverseLabel(%q) = %q, want %q", task.RefRelatesTo, got, reverseLabelRelated)
	}
}

func TestReverseLabelUnknown(t *testing.T) {
	if got := reverseLabel("unknown-type"); got != "" {
		t.Errorf("reverseLabel(%q) = %q, want empty", "unknown-type", got)
	}
}

func TestBuildReverseRefsEmpty(t *testing.T) {
	allTasks := map[string]*task.Task{
		"A": {ID: "A"},
		"B": {ID: "B"},
	}

	refs := buildReverseRefs("A", allTasks)
	if len(refs) != 0 {
		t.Errorf("buildReverseRefs() = %d refs, want 0", len(refs))
	}
}

func TestBuildReverseRefsParent(t *testing.T) {
	allTasks := map[string]*task.Task{
		"parent1": {ID: "parent1", Title: "Parent"},
		"child1":  {ID: "child1", Title: "Child", Refs: []task.Ref{{Type: task.RefParent, ID: "parent1"}}},
	}

	refs := buildReverseRefs("parent1", allTasks)
	if len(refs) != 1 {
		t.Fatalf("buildReverseRefs() = %d refs, want 1", len(refs))
	}

	if refs[0].Label != reverseLabelChild {
		t.Errorf("refs[0].Label = %q, want %q", refs[0].Label, reverseLabelChild)
	}

	if refs[0].Source.ID != "child1" {
		t.Errorf("refs[0].Source.ID = %q, want %q", refs[0].Source.ID, "child1")
	}
}

func TestBuildReverseRefsBlockedBy(t *testing.T) {
	allTasks := map[string]*task.Task{
		"blocker": {ID: "blocker", Title: "Blocker"},
		"blocked": {ID: "blocked", Title: "Blocked", Refs: []task.Ref{{Type: task.RefBlockedBy, ID: "blocker"}}},
	}

	refs := buildReverseRefs("blocker", allTasks)
	if len(refs) != 1 {
		t.Fatalf("buildReverseRefs() = %d refs, want 1", len(refs))
	}

	if refs[0].Label != reverseLabelBlocks {
		t.Errorf("refs[0].Label = %q, want %q", refs[0].Label, reverseLabelBlocks)
	}
}

func TestBuildReverseRefsRelatesTo(t *testing.T) {
	allTasks := map[string]*task.Task{
		"taskA": {ID: "taskA", Title: "Task A"},
		"taskB": {ID: "taskB", Title: "Task B", Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "taskA"}}},
	}

	refs := buildReverseRefs("taskA", allTasks)
	if len(refs) != 1 {
		t.Fatalf("buildReverseRefs() = %d refs, want 1", len(refs))
	}

	if refs[0].Label != reverseLabelRelated {
		t.Errorf("refs[0].Label = %q, want %q", refs[0].Label, reverseLabelRelated)
	}
}

func TestBuildReverseRefsSorted(t *testing.T) {
	allTasks := map[string]*task.Task{
		"target": {ID: "target"},
		"taskC":  {ID: "taskC", Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "target"}}},
		"taskA":  {ID: "taskA", Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "target"}}},
		"taskB":  {ID: "taskB", Refs: []task.Ref{{Type: task.RefBlockedBy, ID: "target"}}},
	}

	refs := buildReverseRefs("target", allTasks)
	if len(refs) != 3 {
		t.Fatalf("buildReverseRefs() = %d refs, want 3", len(refs))
	}

	// "blocks" < "related" by label, then within "related": taskA < taskC by ID.
	if refs[0].Label != reverseLabelBlocks || refs[0].Source.ID != "taskB" {
		t.Errorf("refs[0] = {%q, %q}, want {%s, taskB}", refs[0].Label, refs[0].Source.ID, reverseLabelBlocks)
	}

	if refs[1].Label != reverseLabelRelated || refs[1].Source.ID != "taskA" {
		t.Errorf("refs[1] = {%q, %q}, want {%s, taskA}", refs[1].Label, refs[1].Source.ID, reverseLabelRelated)
	}

	if refs[2].Label != reverseLabelRelated || refs[2].Source.ID != "taskC" {
		t.Errorf("refs[2] = {%q, %q}, want {%s, taskC}", refs[2].Label, refs[2].Source.ID, reverseLabelRelated)
	}
}

func TestBuildReverseRefsSelfExcluded(t *testing.T) {
	allTasks := map[string]*task.Task{
		"self": {ID: "self", Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "self"}}},
	}

	refs := buildReverseRefs("self", allTasks)
	if len(refs) != 0 {
		t.Errorf("buildReverseRefs() = %d refs, want 0 (self-ref excluded)", len(refs))
	}
}

func TestBuildReverseRefsUnknownTypeSkipped(t *testing.T) {
	allTasks := map[string]*task.Task{
		"target": {ID: "target"},
		"other":  {ID: "other", Refs: []task.Ref{{Type: "custom-type", ID: "target"}}},
	}

	refs := buildReverseRefs("target", allTasks)
	if len(refs) != 0 {
		t.Errorf("buildReverseRefs() = %d refs, want 0 (unknown type skipped)", len(refs))
	}
}
