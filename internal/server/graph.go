package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/index"
	"github.com/eunmann/taskboard/internal/task"
)

// Luminance coefficients for perceived brightness (ITU-R BT.601).
const (
	luminanceThreshold = 186
	lumR               = 0.299
	lumG               = 0.587
	lumB               = 0.114
)

// LegendEntry holds a status name and its display color for the legend.
type LegendEntry struct {
	Name  string
	Color string
}

// GraphData holds template data for the graph page.
type GraphData struct {
	Title         string
	Project       string
	TaskCount     int
	MermaidDefs   []string
	FilteredCount int
	StatusLegend  []LegendEntry
}

func handleGraph(idx *index.Index, renderer *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := idx.Config()
		allTasks := idx.List()
		components := findConnectedComponents(allTasks)

		var defs []string

		var filtered int

		for _, comp := range components {
			if isTerminalComponent(comp) {
				filtered++

				continue
			}

			defs = append(defs, buildMermaidDef(comp, cfg.Columns))
		}

		data := GraphData{
			Title:         "Graph",
			Project:       cfg.Project,
			TaskCount:     len(allTasks),
			MermaidDefs:   defs,
			FilteredCount: filtered,
			StatusLegend:  buildStatusLegend(cfg.Columns),
		}

		if isHTMX(r) {
			if err := renderer.RenderPartial(w, "graph", "content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if err := renderer.Render(w, "graph", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func buildMermaidDef(tasks []*task.Task, columns map[string]config.Column) string {
	var b strings.Builder

	b.WriteString("graph LR\n")

	for _, t := range tasks {
		writeNode(&b, t.ID, nodeLabel(t, columns))
	}

	b.WriteByte('\n')

	var refCount int
	for _, t := range tasks {
		refCount += len(t.Refs)
	}

	edgeTypes := make([]string, 0, refCount)

	for _, t := range tasks {
		for _, ref := range t.Refs {
			writeEdge(&b, t.ID, ref)
			edgeTypes = append(edgeTypes, ref.Type)
		}
	}

	b.WriteByte('\n')
	writeLinkStyles(&b, edgeTypes)

	b.WriteByte('\n')
	writeStyles(&b, tasks, columns)

	b.WriteByte('\n')

	for _, t := range tasks {
		writeClickHandler(&b, t.ID)
	}

	return b.String()
}

func nodeLabel(t *task.Task, _ map[string]config.Column) string {
	title := sanitizeMermaid(t.Title)
	if title == "" {
		return t.ID
	}

	if tp := t.Fields["type"]; tp != "" {
		title = tp + ": " + title
	}

	return "<span style='font-size:0.7em;opacity:0.55'>" + t.ID +
		"</span><br/><b>" + title + "</b>"
}

func writeFocusedNode(b *strings.Builder, id, label string) {
	b.WriteString("    subgraph _focus[\"This Task\"]\n")
	b.WriteString("        " + id + "([\"" + label + "\"])\n")
	b.WriteString("    end\n")
}

func writeNode(b *strings.Builder, id, label string) {
	b.WriteString("    " + id + "([\"" + label + "\"])\n")
}

func writeClickHandler(b *strings.Builder, id string) {
	b.WriteString("    click " + id + " \"/task/" + id + "\"\n")
}

func writeEdge(b *strings.Builder, sourceID string, ref task.Ref) {
	switch ref.Type {
	case task.RefParent:
		b.WriteString("    " + sourceID + " --> " + ref.ID + "\n")
	case task.RefBlockedBy:
		b.WriteString("    " + sourceID + " ==> " + ref.ID + "\n")
	case task.RefRelatesTo:
		b.WriteString("    " + sourceID + " -.-> " + ref.ID + "\n")
	}
}

func writeStyles(b *strings.Builder, tasks []*task.Task, columns map[string]config.Column) {
	for _, t := range tasks {
		writeNodeStyle(b, t, columns)
	}
}

func writeNodeStyle(b *strings.Builder, t *task.Task, columns map[string]config.Column) {
	fill := columnColor(columns, "status", t.Fields["status"])
	if fill == "" {
		return
	}

	text := contrastText(fill)
	b.WriteString("    style " + t.ID + " fill:" + fill + ",color:" + text + "\n")
}

func buildStatusLegend(columns map[string]config.Column) []LegendEntry {
	col, ok := columns["status"]
	if !ok {
		return nil
	}

	var legend []LegendEntry

	for _, v := range col.Values {
		if v.Color != "" {
			legend = append(legend, LegendEntry{Name: v.Name, Color: v.Color})
		}
	}

	return legend
}

// findConnectedComponents groups tasks into connected subgraphs using BFS.
// Refs are treated as undirected edges. Results are sorted by first task ID.
func findConnectedComponents(tasks []*task.Task) [][]*task.Task {
	if len(tasks) == 0 {
		return nil
	}

	adj := buildAdjacency(tasks)
	byID := make(map[string]*task.Task, len(tasks))

	for _, t := range tasks {
		byID[t.ID] = t
	}

	visited := make(map[string]bool, len(tasks))

	var components [][]*task.Task

	for _, t := range tasks {
		if visited[t.ID] {
			continue
		}

		comp := bfsComponent(t.ID, adj, visited)

		members := make([]*task.Task, 0, len(comp))
		for _, id := range comp {
			members = append(members, byID[id])
		}

		sort.Slice(members, func(i, j int) bool {
			return members[i].ID < members[j].ID
		})

		components = append(components, members)
	}

	sort.Slice(components, func(i, j int) bool {
		return components[i][0].ID < components[j][0].ID
	})

	return components
}

// buildAdjacency creates a bidirectional adjacency list from task refs.
func buildAdjacency(tasks []*task.Task) map[string][]string {
	ids := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		ids[t.ID] = true
	}

	adj := make(map[string][]string, len(tasks))

	for _, t := range tasks {
		for _, ref := range t.Refs {
			if !ids[ref.ID] {
				continue
			}

			adj[t.ID] = append(adj[t.ID], ref.ID)
			adj[ref.ID] = append(adj[ref.ID], t.ID)
		}
	}

	return adj
}

// bfsComponent collects all node IDs reachable from start via BFS.
func bfsComponent(start string, adj map[string][]string, visited map[string]bool) []string {
	queue := []string{start}
	visited[start] = true

	var result []string

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		result = append(result, cur)

		for _, neighbor := range adj[cur] {
			if visited[neighbor] {
				continue
			}

			visited[neighbor] = true
			queue = append(queue, neighbor)
		}
	}

	return result
}

// isTerminalComponent returns true if every task has status "done" or "cancelled".
func isTerminalComponent(tasks []*task.Task) bool {
	for _, t := range tasks {
		s := t.Fields["status"]
		if s != "done" && s != "cancelled" {
			return false
		}
	}

	return true
}

func writeLinkStyles(b *strings.Builder, edgeTypes []string) {
	for i, et := range edgeTypes {
		switch et {
		case task.RefParent:
			fmt.Fprintf(b, "    linkStyle %d stroke-width:1.5px\n", i)
		case task.RefBlockedBy:
			fmt.Fprintf(b, "    linkStyle %d stroke-width:3.5px\n", i)
		case task.RefRelatesTo:
			fmt.Fprintf(b, "    linkStyle %d stroke-width:1.5px,stroke-dasharray:5 5\n", i)
		}
	}
}

func sanitizeMermaid(s string) string {
	r := strings.NewReplacer(`"`, "#34;", `]`, "#93;")

	return r.Replace(s)
}

func columnColor(columns map[string]config.Column, colName, value string) string {
	col, ok := columns[colName]
	if !ok {
		return ""
	}

	for _, v := range col.Values {
		if v.Name == value {
			return v.Color
		}
	}

	return ""
}

func contrastText(hexColor string) string {
	hex := strings.TrimPrefix(hexColor, "#")

	const hexColorLen = 6

	if len(hex) != hexColorLen {
		return "#fff"
	}

	const (
		hexBase = 16
		hexBits = 8
	)

	r, _ := strconv.ParseUint(hex[0:2], hexBase, hexBits)
	g, _ := strconv.ParseUint(hex[2:4], hexBase, hexBits)
	b, _ := strconv.ParseUint(hex[4:6], hexBase, hexBits)

	lum := lumR*float64(r) + lumG*float64(g) + lumB*float64(b)
	if lum > luminanceThreshold {
		return "#000"
	}

	return "#fff"
}

// buildTaskGraph generates a focused Mermaid graph showing the current task
// and its 1-hop neighbors (forward refs + reverse refs). Returns "" if the
// task has no connections.
func buildTaskGraph(
	taskID string,
	t *task.Task,
	allTasks map[string]*task.Task,
	reverseRefs []ReverseRef,
	columns map[string]config.Column,
) string {
	neighbors := make(map[string]*task.Task)

	for _, ref := range t.Refs {
		if neighbor, ok := allTasks[ref.ID]; ok {
			neighbors[ref.ID] = neighbor
		}
	}

	for _, rr := range reverseRefs {
		neighbors[rr.Source.ID] = rr.Source
	}

	if len(neighbors) == 0 {
		return ""
	}

	sorted := sortedNeighbors(neighbors)

	var b strings.Builder

	b.WriteString("graph TD\n")

	// Write focused node first, wrapped in a labeled subgraph.
	writeFocusedNode(&b, taskID, nodeLabel(t, columns))

	for _, n := range sorted {
		writeNode(&b, n.ID, nodeLabel(n, columns))
	}

	b.WriteByte('\n')

	edgeTypes := writeTaskGraphEdges(&b, taskID, t, reverseRefs, allTasks)

	b.WriteByte('\n')
	writeLinkStyles(&b, edgeTypes)

	b.WriteByte('\n')
	writeNodeStyle(&b, t, columns)

	for _, n := range sorted {
		writeNodeStyle(&b, n, columns)
	}

	b.WriteString("    style _focus fill:none,stroke:none\n")

	b.WriteByte('\n')
	writeClickHandler(&b, taskID)

	for _, n := range sorted {
		writeClickHandler(&b, n.ID)
	}

	return b.String()
}

// writeTaskGraphEdges writes edges for forward refs and reverse refs,
// deduplicating mutual relates-to edges. Returns edge types for link styling.
func writeTaskGraphEdges(
	b *strings.Builder,
	taskID string,
	t *task.Task,
	reverseRefs []ReverseRef,
	allTasks map[string]*task.Task,
) []string {
	type edgeKey struct{ from, to string }

	seen := make(map[edgeKey]bool)

	var edgeTypes []string

	// Forward refs: task → target.
	for _, ref := range t.Refs {
		if _, ok := allTasks[ref.ID]; !ok {
			continue
		}

		key := edgeKey{taskID, ref.ID}
		if seen[key] {
			continue
		}

		seen[key] = true

		// For relates-to, also mark reverse direction to deduplicate.
		if ref.Type == task.RefRelatesTo {
			seen[edgeKey{ref.ID, taskID}] = true
		}

		writeEdge(b, taskID, ref)
		edgeTypes = append(edgeTypes, ref.Type)
	}

	// Reverse refs: source → task (using forward ref type).
	for _, rr := range reverseRefs {
		fwdType := forwardRefType(rr.Label)
		key := edgeKey{rr.Source.ID, taskID}

		if seen[key] {
			continue
		}

		seen[key] = true

		if fwdType == task.RefRelatesTo {
			seen[edgeKey{taskID, rr.Source.ID}] = true
		}

		writeEdge(b, rr.Source.ID, task.Ref{Type: fwdType, ID: taskID})
		edgeTypes = append(edgeTypes, fwdType)
	}

	return edgeTypes
}

// forwardRefType maps a reverse label back to the forward ref type constant.
func forwardRefType(reverseLabel string) string {
	switch reverseLabel {
	case reverseLabelChild:
		return task.RefParent
	case reverseLabelBlocks:
		return task.RefBlockedBy
	case reverseLabelRelated:
		return task.RefRelatesTo
	default:
		return ""
	}
}

// sortedNeighbors returns neighbor tasks in deterministic order (by ID).
func sortedNeighbors(neighbors map[string]*task.Task) []*task.Task {
	result := make([]*task.Task, 0, len(neighbors))
	for _, t := range neighbors {
		result = append(result, t)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}
