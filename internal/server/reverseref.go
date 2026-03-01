package server

import (
	"sort"

	"github.com/eunmann/taskboard/internal/task"
)

// Reverse label constants displayed for inbound references.
const (
	reverseLabelChild   = "child"
	reverseLabelBlocks  = "blocks"
	reverseLabelRelated = "related"
)

// ReverseRef represents an inbound reference from another task.
type ReverseRef struct {
	Label  string
	Source *task.Task
}

// reverseLabel maps a forward ref type to its reverse display label.
func reverseLabel(refType string) string {
	switch refType {
	case task.RefParent:
		return reverseLabelChild
	case task.RefBlockedBy:
		return reverseLabelBlocks
	case task.RefRelatesTo:
		return reverseLabelRelated
	default:
		return ""
	}
}

// buildReverseRefs scans all tasks for refs pointing at taskID and returns
// sorted reverse references (by label, then source ID).
func buildReverseRefs(taskID string, allTasks map[string]*task.Task) []ReverseRef {
	var refs []ReverseRef

	for _, t := range allTasks {
		if t.ID == taskID {
			continue
		}

		for _, r := range t.Refs {
			if r.ID != taskID {
				continue
			}

			label := reverseLabel(r.Type)
			if label == "" {
				continue
			}

			refs = append(refs, ReverseRef{Label: label, Source: t})
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Label != refs[j].Label {
			return refs[i].Label < refs[j].Label
		}

		return refs[i].Source.ID < refs[j].Source.ID
	})

	return refs
}
