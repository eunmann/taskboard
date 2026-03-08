package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/index"
	"github.com/eunmann/taskboard/internal/task"
	"github.com/eunmann/taskboard/internal/web/testhelpers"
)

func setupGraphTestIndex(t *testing.T) *index.Index {
	t.Helper()

	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Graph Test
columns:
  status:
    order: 1
    values:
      - name: open
        color: '#22c55e'
      - name: done
        color: '#6b7280'
  type:
    order: 2
    values:
      - name: chore
        color: '#64748b'
      - name: feature
        color: '#8b5cf6'
`)

	writeTestFile(t, dir, "parent1.yaml", `
title: Parent Feature
status: open
type: feature
`)

	writeTestFile(t, dir, "child1.yaml", `
title: Child Task
status: open
type: chore
refs:
  - type: parent
    id: parent1
`)

	writeTestFile(t, dir, "blocked1.yaml", `
title: Blocked Task
status: open
type: chore
refs:
  - type: blocked-by
    id: parent1
`)

	writeTestFile(t, dir, "related1.yaml", `
title: Related Task
status: done
type: chore
refs:
  - type: relates-to
    id: child1
`)

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}

	return idx
}

func TestHandleGraph(t *testing.T) {
	idx := setupGraphTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()

	if !strings.Contains(body, "Parent Feature") {
		t.Error("response missing 'Parent Feature'")
	}

	if !strings.Contains(body, "mermaid") {
		t.Error("response missing mermaid class/script")
	}
}

func TestHandleGraphHTMX(t *testing.T) {
	idx := setupGraphTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()

	if !strings.Contains(body, "Parent Feature") {
		t.Error("HTMX response missing 'Parent Feature'")
	}

	if strings.Contains(body, "<html") {
		t.Error("HTMX response should not contain full HTML layout")
	}
}

func TestGraphPageHTMLStructure(t *testing.T) {
	idx := setupGraphTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertTitle(t, doc, "Graph - Taskboard")
	testhelpers.AssertElementExists(t, doc, "pre.mermaid")
}

func TestGraphPageNavLinks(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertElementExists(t, doc, `a[href="/"]`)
	testhelpers.AssertElementExists(t, doc, `a[href="/graph"]`)
}

func TestBuildMermaidDefNodes(t *testing.T) {
	tasks := []*task.Task{
		{ID: "abc", Title: "First Task", Fields: map[string]string{"type": "fix", "status": "open"}},
		{ID: "def", Title: "Second Task", Fields: map[string]string{"type": "feature", "status": "done"}},
	}

	def := buildMermaidDef(tasks, nil)

	if !strings.Contains(def, "abc") || !strings.Contains(def, "<b>fix: First Task</b>") {
		t.Errorf("missing stadium node abc with type-prefixed bold title in:\n%s", def)
	}

	if !strings.Contains(def, "def") || !strings.Contains(def, "<b>feature: Second Task</b>") {
		t.Errorf("missing stadium node def with type-prefixed bold title in:\n%s", def)
	}
}

func TestBuildMermaidDefEdgesAndLinkStyles(t *testing.T) {
	tasks := []*task.Task{
		{ID: "a", Title: "Parent", Fields: map[string]string{}},
		{
			ID: "b", Title: "Child", Fields: map[string]string{},
			Refs: []task.Ref{{Type: task.RefParent, ID: "a"}},
		},
		{
			ID: "c", Title: "Blocked", Fields: map[string]string{},
			Refs: []task.Ref{{Type: task.RefBlockedBy, ID: "a"}},
		},
		{
			ID: "d", Title: "Related", Fields: map[string]string{},
			Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "b"}},
		},
	}

	def := buildMermaidDef(tasks, nil)

	if !strings.Contains(def, "b --> a") {
		t.Errorf("missing parent edge in:\n%s", def)
	}

	if !strings.Contains(def, "c ==> a") {
		t.Errorf("missing blocked-by edge in:\n%s", def)
	}

	if !strings.Contains(def, "d -.-> b") {
		t.Errorf("missing relates-to edge in:\n%s", def)
	}

	if !strings.Contains(def, "linkStyle 0 stroke-width:1.5px") {
		t.Errorf("missing parent linkStyle in:\n%s", def)
	}

	if !strings.Contains(def, "linkStyle 1 stroke-width:3.5px") {
		t.Errorf("missing blocked-by linkStyle in:\n%s", def)
	}

	if !strings.Contains(def, "linkStyle 2 stroke-width:1.5px,stroke-dasharray:5 5") {
		t.Errorf("missing relates-to linkStyle in:\n%s", def)
	}
}

func TestBuildMermaidDefClickHandlers(t *testing.T) {
	tasks := []*task.Task{
		{ID: "abc", Title: "Task", Fields: map[string]string{}},
		{ID: "def", Title: "Other", Fields: map[string]string{}},
	}

	def := buildMermaidDef(tasks, nil)

	if !strings.Contains(def, `click abc "/task/abc"`) {
		t.Errorf("missing click handler for abc in:\n%s", def)
	}

	if !strings.Contains(def, `click def "/task/def"`) {
		t.Errorf("missing click handler for def in:\n%s", def)
	}
}

func TestBuildMermaidDefEmpty(t *testing.T) {
	def := buildMermaidDef(nil, nil)

	if !strings.Contains(def, "graph LR") {
		t.Errorf("empty graph should still have header, got:\n%s", def)
	}
}

func TestNodeLabelWithTypeAndStatus(t *testing.T) {
	tk := &task.Task{
		ID:     "Abc123",
		Title:  "My Task",
		Fields: map[string]string{"type": "fix", "status": "open"},
	}

	label := nodeLabel(tk, nil)

	if !strings.Contains(label, "<b>fix: My Task</b>") {
		t.Errorf("label missing type-prefixed bold title: %s", label)
	}

	if !strings.Contains(label, "Abc123") {
		t.Errorf("label missing task ID: %s", label)
	}

	if !strings.Contains(label, "font-size:0.7em") {
		t.Errorf("label missing small ID styling: %s", label)
	}

	if strings.Contains(label, "open") {
		t.Errorf("label should not contain status text: %s", label)
	}
}

func TestNodeLabelNoMeta(t *testing.T) {
	tk := &task.Task{
		ID:     "xyz",
		Title:  "Plain Task",
		Fields: map[string]string{},
	}

	label := nodeLabel(tk, nil)

	if !strings.Contains(label, "xyz") {
		t.Errorf("label missing ID: %s", label)
	}

	if !strings.Contains(label, "<b>Plain Task</b>") {
		t.Errorf("label missing bold title: %s", label)
	}
}

func TestNodeLabelEmptyTitle(t *testing.T) {
	tk := &task.Task{
		ID:     "xyz",
		Title:  "",
		Fields: map[string]string{},
	}

	label := nodeLabel(tk, nil)

	if label != "xyz" {
		t.Errorf("nodeLabel() = %q, want ID fallback %q", label, "xyz")
	}
}

func TestNodeLabelPartialMeta(t *testing.T) {
	tk := &task.Task{
		ID:     "p1",
		Title:  "Task",
		Fields: map[string]string{"status": "open"},
	}

	label := nodeLabel(tk, nil)

	if !strings.Contains(label, "<b>Task</b>") {
		t.Errorf("label missing bold title: %s", label)
	}

	if strings.Contains(label, "open") {
		t.Errorf("label should not contain status text: %s", label)
	}
}

func TestSanitizeMermaid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`no special chars`, `no special chars`},
		{`has "quotes"`, `has #34;quotes#34;`},
		{`has [brackets]`, `has [brackets#93;`},
		{`"quoted] combo`, `#34;quoted#93; combo`},
	}

	for _, tc := range tests {
		got := sanitizeMermaid(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeMermaid(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestColumnColor(t *testing.T) {
	columns := map[string]config.Column{
		"status": {Values: []config.Value{
			{Name: "open", Color: "#22c55e"},
			{Name: "done", Color: "#6b7280"},
			{Name: "nocolor"},
		}},
	}

	if c := columnColor(columns, "status", "open"); c != "#22c55e" {
		t.Errorf("columnColor(status, open) = %q, want #22c55e", c)
	}

	if c := columnColor(columns, "status", "done"); c != "#6b7280" {
		t.Errorf("columnColor(status, done) = %q, want #6b7280", c)
	}

	if c := columnColor(columns, "status", "nocolor"); c != "" {
		t.Errorf("columnColor(status, nocolor) = %q, want empty", c)
	}

	if c := columnColor(columns, "status", "unknown"); c != "" {
		t.Errorf("columnColor(status, unknown) = %q, want empty", c)
	}

	if c := columnColor(columns, "missing", "open"); c != "" {
		t.Errorf("columnColor(missing, open) = %q, want empty", c)
	}
}

func TestContrastText(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"#000000", "#fff"},
		{"#ffffff", "#000"},
		{"#22c55e", "#fff"},
		{"#6b7280", "#fff"},
		{"#3b82f6", "#fff"},
		{"#eab308", "#fff"},
		{"invalid", "#fff"},
		{"#abc", "#fff"},
	}

	for _, tc := range tests {
		got := contrastText(tc.hex)
		if got != tc.want {
			t.Errorf("contrastText(%q) = %q, want %q", tc.hex, got, tc.want)
		}
	}
}

func TestBuildMermaidDefNodeStyles(t *testing.T) {
	columns := map[string]config.Column{
		"status": {Values: []config.Value{
			{Name: "open", Color: "#22c55e"},
			{Name: "done", Color: "#6b7280"},
		}},
		"type": {Values: []config.Value{
			{Name: "feature", Color: "#8b5cf6"},
		}},
	}
	tasks := []*task.Task{
		{ID: "a", Title: "Open", Fields: map[string]string{"status": "open", "type": "feature"}},
		{ID: "b", Title: "Done", Fields: map[string]string{"status": "done"}},
		{ID: "c", Title: "None", Fields: map[string]string{}},
	}

	def := buildMermaidDef(tasks, columns)

	if !strings.Contains(def, "style a fill:#22c55e,color:#fff") {
		t.Errorf("missing fill style for node a in:\n%s", def)
	}

	if !strings.Contains(def, "style b fill:#6b7280,color:#fff") {
		t.Errorf("missing fill style for node b in:\n%s", def)
	}

	if strings.Contains(def, "style c fill:") {
		t.Errorf("node c without status/type should have no style in:\n%s", def)
	}
}

func TestStatusLegend(t *testing.T) {
	columns := map[string]config.Column{
		"status": {Values: []config.Value{
			{Name: "open", Color: "#22c55e"},
			{Name: "nocolor"},
			{Name: "done", Color: "#6b7280"},
		}},
	}

	legend := buildStatusLegend(columns)

	if len(legend) != 2 {
		t.Fatalf("buildStatusLegend() returned %d entries, want 2", len(legend))
	}

	if legend[0].Name != "open" || legend[0].Color != "#22c55e" {
		t.Errorf("legend[0] = %+v, want {open, #22c55e}", legend[0])
	}

	if legend[1].Name != "done" || legend[1].Color != "#6b7280" {
		t.Errorf("legend[1] = %+v, want {done, #6b7280}", legend[1])
	}
}

func TestStatusLegendNoStatusColumn(t *testing.T) {
	legend := buildStatusLegend(nil)

	if legend != nil {
		t.Errorf("buildStatusLegend(nil) = %v, want nil", legend)
	}
}

func TestWriteLinkStyles(t *testing.T) {
	edgeTypes := []string{task.RefParent, task.RefBlockedBy, task.RefRelatesTo}

	var b strings.Builder

	writeLinkStyles(&b, edgeTypes)
	got := b.String()

	if !strings.Contains(got, "linkStyle 0 stroke-width:1.5px") {
		t.Errorf("missing parent linkStyle in:\n%s", got)
	}

	if !strings.Contains(got, "linkStyle 1 stroke-width:3.5px") {
		t.Errorf("missing blocked-by linkStyle in:\n%s", got)
	}

	if !strings.Contains(got, "linkStyle 2 stroke-width:1.5px,stroke-dasharray:5 5") {
		t.Errorf("missing relates-to linkStyle in:\n%s", got)
	}
}

func TestWriteLinkStylesEmpty(t *testing.T) {
	var b strings.Builder

	writeLinkStyles(&b, nil)

	if b.Len() != 0 {
		t.Errorf("writeLinkStyles with no edges should produce empty output, got: %q", b.String())
	}
}

func TestWriteNodeStyleBothStatusAndType(t *testing.T) {
	columns := map[string]config.Column{
		"status": {Values: []config.Value{{Name: "open", Color: "#22c55e"}}},
		"type":   {Values: []config.Value{{Name: "fix", Color: "#ef4444"}}},
	}
	tk := &task.Task{ID: "t1", Fields: map[string]string{"status": "open", "type": "fix"}}

	var b strings.Builder

	writeNodeStyle(&b, tk, columns)
	got := b.String()

	want := "style t1 fill:#22c55e,color:#fff"
	if !strings.Contains(got, want) {
		t.Errorf("writeNodeStyle() = %q, want %q", got, want)
	}
}

func TestWriteNodeStyleStatusOnly(t *testing.T) {
	columns := map[string]config.Column{
		"status": {Values: []config.Value{{Name: "open", Color: "#22c55e"}}},
	}
	tk := &task.Task{ID: "t1", Fields: map[string]string{"status": "open"}}

	var b strings.Builder

	writeNodeStyle(&b, tk, columns)
	got := b.String()

	want := "style t1 fill:#22c55e,color:#fff"
	if !strings.Contains(got, want) {
		t.Errorf("writeNodeStyle() = %q, want %q", got, want)
	}
}

func TestWriteNodeStyleTypeOnly(t *testing.T) {
	columns := map[string]config.Column{
		"type": {Values: []config.Value{{Name: "fix", Color: "#ef4444"}}},
	}
	tk := &task.Task{ID: "t1", Fields: map[string]string{"type": "fix"}}

	var b strings.Builder

	writeNodeStyle(&b, tk, columns)

	if b.Len() != 0 {
		t.Errorf("writeNodeStyle() with type-only should produce no output, got: %q", b.String())
	}
}

func TestWriteNodeStyleNeither(t *testing.T) {
	tk := &task.Task{ID: "t1", Fields: map[string]string{}}

	var b strings.Builder

	writeNodeStyle(&b, tk, nil)

	if b.Len() != 0 {
		t.Errorf("writeNodeStyle() with no status/type should produce no output, got: %q", b.String())
	}
}

func TestBuildTaskGraphBasic(t *testing.T) {
	parent := &task.Task{ID: "parent1", Title: "Parent", Fields: map[string]string{}}
	child := &task.Task{
		ID: "child1", Title: "Child", Fields: map[string]string{},
		Refs: []task.Ref{{Type: task.RefParent, ID: "parent1"}},
	}
	allTasks := map[string]*task.Task{"parent1": parent, "child1": child}

	def := buildTaskGraph("child1", child, allTasks, nil, nil)

	if !strings.Contains(def, "child1") {
		t.Errorf("missing focused node child1 in:\n%s", def)
	}

	if !strings.Contains(def, "parent1") {
		t.Errorf("missing neighbor node parent1 in:\n%s", def)
	}

	if !strings.Contains(def, "child1 --> parent1") {
		t.Errorf("missing parent edge in:\n%s", def)
	}
}

func TestBuildTaskGraphReverseRefs(t *testing.T) {
	parent := &task.Task{ID: "parent1", Title: "Parent", Fields: map[string]string{}}
	child := &task.Task{
		ID: "child1", Title: "Child", Fields: map[string]string{},
		Refs: []task.Ref{{Type: task.RefParent, ID: "parent1"}},
	}
	allTasks := map[string]*task.Task{"parent1": parent, "child1": child}
	reverseRefs := []ReverseRef{{Label: reverseLabelChild, Source: child}}

	def := buildTaskGraph("parent1", parent, allTasks, reverseRefs, nil)

	if !strings.Contains(def, "child1 --> parent1") {
		t.Errorf("missing reverse-ref edge in:\n%s", def)
	}
}

func TestBuildTaskGraphEmpty(t *testing.T) {
	isolated := &task.Task{ID: "alone1", Title: "Alone", Fields: map[string]string{}}
	allTasks := map[string]*task.Task{"alone1": isolated}

	def := buildTaskGraph("alone1", isolated, allTasks, nil, nil)

	if def != "" {
		t.Errorf("expected empty string for isolated task, got:\n%s", def)
	}
}

func TestBuildTaskGraphDanglingRef(t *testing.T) {
	tk := &task.Task{
		ID: "t1", Title: "Task", Fields: map[string]string{},
		Refs: []task.Ref{{Type: task.RefParent, ID: "missing"}},
	}
	allTasks := map[string]*task.Task{"t1": tk}

	def := buildTaskGraph("t1", tk, allTasks, nil, nil)

	if def != "" {
		t.Errorf("expected empty string when all refs are dangling, got:\n%s", def)
	}
}

func TestBuildTaskGraphFocusedHighlight(t *testing.T) {
	parent := &task.Task{ID: "parent1", Title: "Parent", Fields: map[string]string{}}
	child := &task.Task{
		ID: "child1", Title: "Child", Fields: map[string]string{},
		Refs: []task.Ref{{Type: task.RefParent, ID: "parent1"}},
	}
	allTasks := map[string]*task.Task{"parent1": parent, "child1": child}

	def := buildTaskGraph("child1", child, allTasks, nil, nil)

	if !strings.Contains(def, "subgraph _focus") {
		t.Errorf("missing subgraph wrapper in:\n%s", def)
	}

	if !strings.Contains(def, "This Task") {
		t.Errorf("missing subgraph title in:\n%s", def)
	}

	if strings.Contains(def, "focusedNodeStyle") {
		t.Errorf("should not contain old focusedNodeStyle in:\n%s", def)
	}
}

func TestBuildTaskGraphDedup(t *testing.T) {
	a := &task.Task{
		ID: "a1", Title: "A", Fields: map[string]string{},
		Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "b1"}},
	}
	b := &task.Task{
		ID: "b1", Title: "B", Fields: map[string]string{},
		Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "a1"}},
	}
	allTasks := map[string]*task.Task{"a1": a, "b1": b}
	reverseRefs := []ReverseRef{{Label: reverseLabelRelated, Source: b}}

	def := buildTaskGraph("a1", a, allTasks, reverseRefs, nil)

	// Count relates-to edges: should be exactly 1, not 2.
	count := strings.Count(def, "-.->")
	if count != 1 {
		t.Errorf("expected 1 relates-to edge (deduplicated), got %d in:\n%s", count, def)
	}
}

func TestForwardRefType(t *testing.T) {
	tests := []struct {
		label string
		want  string
	}{
		{reverseLabelChild, task.RefParent},
		{reverseLabelBlocks, task.RefBlockedBy},
		{reverseLabelRelated, task.RefRelatesTo},
		{"unknown", ""},
	}

	for _, tc := range tests {
		got := forwardRefType(tc.label)
		if got != tc.want {
			t.Errorf("forwardRefType(%q) = %q, want %q", tc.label, got, tc.want)
		}
	}
}

func TestSortedNeighbors(t *testing.T) {
	neighbors := map[string]*task.Task{
		"c1": {ID: "c1"},
		"a1": {ID: "a1"},
		"b1": {ID: "b1"},
	}

	sorted := sortedNeighbors(neighbors)

	if len(sorted) != 3 {
		t.Fatalf("sortedNeighbors() length = %d, want 3", len(sorted))
	}

	if sorted[0].ID != "a1" || sorted[1].ID != "b1" || sorted[2].ID != "c1" {
		ids := make([]string, len(sorted))
		for i, s := range sorted {
			ids[i] = s.ID
		}

		t.Errorf("sortedNeighbors() = %v, want [a1, b1, c1]", ids)
	}
}

func TestFindConnectedComponents(t *testing.T) {
	tasks := []*task.Task{
		{ID: "a", Title: "A", Fields: map[string]string{}, Refs: []task.Ref{{Type: task.RefParent, ID: "b"}}},
		{ID: "b", Title: "B", Fields: map[string]string{}},
		{ID: "c", Title: "C", Fields: map[string]string{}, Refs: []task.Ref{{Type: task.RefBlockedBy, ID: "d"}}},
		{ID: "d", Title: "D", Fields: map[string]string{}},
		{ID: "e", Title: "E", Fields: map[string]string{}},
	}

	components := findConnectedComponents(tasks)

	if len(components) != 3 {
		t.Fatalf("findConnectedComponents() returned %d components, want 3", len(components))
	}

	// Components sorted by first ID: {a,b}, {c,d}, {e}
	if components[0][0].ID != "a" || components[0][1].ID != "b" {
		t.Errorf("component 0 = %v, want [a, b]", compIDs(components[0]))
	}

	if components[1][0].ID != "c" || components[1][1].ID != "d" {
		t.Errorf("component 1 = %v, want [c, d]", compIDs(components[1]))
	}

	if len(components[2]) != 1 || components[2][0].ID != "e" {
		t.Errorf("component 2 = %v, want [e]", compIDs(components[2]))
	}
}

func TestFindConnectedComponentsSingleCluster(t *testing.T) {
	tasks := []*task.Task{
		{ID: "a", Title: "A", Fields: map[string]string{}, Refs: []task.Ref{{Type: task.RefParent, ID: "b"}}},
		{ID: "b", Title: "B", Fields: map[string]string{}, Refs: []task.Ref{{Type: task.RefParent, ID: "c"}}},
		{ID: "c", Title: "C", Fields: map[string]string{}},
	}

	components := findConnectedComponents(tasks)

	if len(components) != 1 {
		t.Fatalf("findConnectedComponents() returned %d components, want 1", len(components))
	}

	if len(components[0]) != 3 {
		t.Errorf("component has %d tasks, want 3", len(components[0]))
	}
}

func TestFindConnectedComponentsEmpty(t *testing.T) {
	components := findConnectedComponents(nil)

	if components != nil {
		t.Errorf("findConnectedComponents(nil) = %v, want nil", components)
	}
}

func TestFindConnectedComponentsBidirectional(t *testing.T) {
	// A refs B but B has no ref to A — they should still be in the same component.
	tasks := []*task.Task{
		{ID: "a", Title: "A", Fields: map[string]string{}, Refs: []task.Ref{{Type: task.RefRelatesTo, ID: "b"}}},
		{ID: "b", Title: "B", Fields: map[string]string{}},
	}

	components := findConnectedComponents(tasks)

	if len(components) != 1 {
		t.Fatalf("findConnectedComponents() returned %d components, want 1", len(components))
	}

	if len(components[0]) != 2 {
		t.Errorf("component has %d tasks, want 2", len(components[0]))
	}
}

func TestIsTerminalComponent(t *testing.T) {
	tests := []struct {
		name  string
		tasks []*task.Task
		want  bool
	}{
		{
			name: "all done",
			tasks: []*task.Task{
				{ID: "a", Fields: map[string]string{"status": "done"}},
				{ID: "b", Fields: map[string]string{"status": "done"}},
			},
			want: true,
		},
		{
			name: "all cancelled",
			tasks: []*task.Task{
				{ID: "a", Fields: map[string]string{"status": "cancelled"}},
			},
			want: true,
		},
		{
			name: "mix done and cancelled",
			tasks: []*task.Task{
				{ID: "a", Fields: map[string]string{"status": "done"}},
				{ID: "b", Fields: map[string]string{"status": "cancelled"}},
			},
			want: true,
		},
		{
			name: "one open among done",
			tasks: []*task.Task{
				{ID: "a", Fields: map[string]string{"status": "done"}},
				{ID: "b", Fields: map[string]string{"status": "open"}},
			},
			want: false,
		},
		{
			name: "in-progress",
			tasks: []*task.Task{
				{ID: "a", Fields: map[string]string{"status": "in-progress"}},
			},
			want: false,
		},
		{
			name:  "empty component",
			tasks: nil,
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isTerminalComponent(tc.tasks)
			if got != tc.want {
				t.Errorf("isTerminalComponent() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHandleGraphFiltersDoneComponents(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Filter Test
columns:
  status:
    order: 1
    values:
      - name: open
        color: '#22c55e'
      - name: done
        color: '#6b7280'
`)
	writeTestFile(t, dir, "active1.yaml", "title: Active Task\nstatus: open\n")
	writeTestFile(t, dir, "active2.yaml", "title: Active Dep\nstatus: open\nrefs:\n  - type: parent\n    id: active1\n")
	writeTestFile(t, dir, "done1.yaml", "title: Done Task\nstatus: done\n")
	writeTestFile(t, dir, "done2.yaml", "title: Done Dep\nstatus: done\nrefs:\n  - type: parent\n    id: done1\n")

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()

	if !strings.Contains(body, "Active Task") {
		t.Error("response missing active cluster task")
	}

	if strings.Contains(body, "Done Task") {
		t.Error("response should not contain done cluster task")
	}

	if !strings.Contains(body, "1 completed cluster hidden") {
		t.Error("response missing filtered count message")
	}
}

func TestHandleGraphAllDone(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: All Done Test
columns:
  status:
    order: 1
    values:
      - name: done
        color: '#6b7280'
`)
	writeTestFile(t, dir, "d1.yaml", "title: Done One\nstatus: done\n")
	writeTestFile(t, dir, "d2.yaml", "title: Done Two\nstatus: done\n")

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()

	if !strings.Contains(body, "All task clusters are completed") {
		t.Error("response missing all-done empty state message")
	}
}

func TestFindConnectedComponentsDanglingRef(t *testing.T) {
	// Task A refs nonexistent "missing" — dangling ref is skipped in adjacency,
	// so A becomes an isolated single-node component.
	tasks := []*task.Task{
		{ID: "a", Title: "A", Fields: map[string]string{}, Refs: []task.Ref{{Type: task.RefParent, ID: "missing"}}},
		{ID: "b", Title: "B", Fields: map[string]string{}},
	}

	components := findConnectedComponents(tasks)

	if len(components) != 2 {
		t.Fatalf("findConnectedComponents() returned %d components, want 2", len(components))
	}

	if len(components[0]) != 1 || components[0][0].ID != "a" {
		t.Errorf("component 0 = %v, want [a]", compIDs(components[0]))
	}

	if len(components[1]) != 1 || components[1][0].ID != "b" {
		t.Errorf("component 1 = %v, want [b]", compIDs(components[1]))
	}
}

func TestHandleGraphFiltersSingleDoneNode(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Single Done Test
columns:
  status:
    order: 1
    values:
      - name: open
        color: '#22c55e'
      - name: done
        color: '#6b7280'
`)
	// Isolated done node — should be filtered.
	writeTestFile(t, dir, "solo_done.yaml", "title: Solo Done\nstatus: done\n")
	// Isolated open node — should be kept.
	writeTestFile(t, dir, "solo_open.yaml", "title: Solo Open\nstatus: open\n")
	// Connected active pair — should be kept.
	writeTestFile(t, dir, "pair_a.yaml", "title: Pair A\nstatus: open\n")
	writeTestFile(t, dir, "pair_b.yaml", "title: Pair B\nstatus: open\nrefs:\n  - type: parent\n    id: pair_a\n")

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()
	doc := testhelpers.ParseHTML(t, w)

	// Two active components rendered (solo_open + pair).
	testhelpers.AssertElementCount(t, doc, "pre.mermaid", 2)

	if !strings.Contains(body, "Solo Open") {
		t.Error("response missing isolated open node")
	}

	if !strings.Contains(body, "Pair A") {
		t.Error("response missing connected active pair")
	}

	if strings.Contains(body, "Solo Done") {
		t.Error("response should not contain isolated done node")
	}

	if !strings.Contains(body, "1 completed cluster hidden") {
		t.Error("response missing filtered count message")
	}
}

func TestHandleGraphKeepsMixedStatusComponent(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Mixed Test
columns:
  status:
    order: 1
    values:
      - name: open
        color: '#22c55e'
      - name: done
        color: '#6b7280'
`)
	// Connected component: one open, one done — must NOT be filtered.
	writeTestFile(t, dir, "mix_open.yaml", "title: Mix Open\nstatus: open\n")
	writeTestFile(t, dir, "mix_done.yaml", "title: Mix Done\nstatus: done\nrefs:\n  - type: parent\n    id: mix_open\n")

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()
	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertElementCount(t, doc, "pre.mermaid", 1)

	if !strings.Contains(body, "Mix Open") {
		t.Error("response missing open task in mixed component")
	}

	if !strings.Contains(body, "Mix Done") {
		t.Error("response missing done task in mixed component")
	}

	if strings.Contains(body, "completed cluster") {
		t.Error("no clusters should be filtered in mixed component test")
	}
}

func TestHandleGraphForest(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Forest Test
columns:
  status:
    order: 1
    values:
      - name: open
        color: '#22c55e'
      - name: done
        color: '#6b7280'
      - name: cancelled
        color: '#ef4444'
`)
	// Isolated done node — filtered.
	writeTestFile(t, dir, "iso_done.yaml", "title: Iso Done\nstatus: done\n")
	// Isolated cancelled node — filtered.
	writeTestFile(t, dir, "iso_cancel.yaml", "title: Iso Cancel\nstatus: cancelled\n")
	// Isolated open node — kept.
	writeTestFile(t, dir, "iso_open.yaml", "title: Iso Open\nstatus: open\n")
	// Connected all-done cluster — filtered.
	writeTestFile(t, dir, "clust_d1.yaml", "title: Cluster Done A\nstatus: done\n")
	writeTestFile(t, dir, "clust_d2.yaml", "title: Cluster Done B\nstatus: done\nrefs:\n  - type: parent\n    id: clust_d1\n")
	// Connected mixed cluster — kept.
	writeTestFile(t, dir, "clust_m1.yaml", "title: Cluster Mix Open\nstatus: open\n")
	writeTestFile(t, dir, "clust_m2.yaml", "title: Cluster Mix Done\nstatus: done\nrefs:\n  - type: blocked-by\n    id: clust_m1\n")

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleGraph(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()
	doc := testhelpers.ParseHTML(t, w)

	// 2 active components: iso_open + mixed cluster.
	testhelpers.AssertElementCount(t, doc, "pre.mermaid", 2)

	// Active components present.
	if !strings.Contains(body, "Iso Open") {
		t.Error("response missing isolated open node")
	}

	if !strings.Contains(body, "Cluster Mix Open") {
		t.Error("response missing open task in mixed cluster")
	}

	if !strings.Contains(body, "Cluster Mix Done") {
		t.Error("response missing done task in mixed cluster (kept because partner is open)")
	}

	// Filtered components absent.
	if strings.Contains(body, "Iso Done") {
		t.Error("response should not contain isolated done node")
	}

	if strings.Contains(body, "Iso Cancel") {
		t.Error("response should not contain isolated cancelled node")
	}

	if strings.Contains(body, "Cluster Done A") {
		t.Error("response should not contain all-done cluster task A")
	}

	if strings.Contains(body, "Cluster Done B") {
		t.Error("response should not contain all-done cluster task B")
	}

	// 3 filtered: iso_done, iso_cancel, clust_d1+d2.
	if !strings.Contains(body, "3 completed clusters hidden") {
		t.Error("response missing correct filtered count (expected 3)")
	}
}

func compIDs(tasks []*task.Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}

	return ids
}
