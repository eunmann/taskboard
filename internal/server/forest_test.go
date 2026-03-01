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

func setupForestTestIndex(t *testing.T) *index.Index {
	t.Helper()

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
  type:
    order: 2
    values:
      - name: task
      - name: epic
        color: '#06b6d4'
`)

	writeTestFile(t, dir, "parent1.yaml", `
title: Parent Epic
status: open
type: epic
`)

	writeTestFile(t, dir, "child1.yaml", `
title: Child Task
status: open
type: task
refs:
  - type: parent
    id: parent1
`)

	writeTestFile(t, dir, "blocked1.yaml", `
title: Blocked Task
status: open
type: task
refs:
  - type: blocked-by
    id: parent1
`)

	writeTestFile(t, dir, "related1.yaml", `
title: Related Task
status: done
type: task
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

func TestHandleForest(t *testing.T) {
	idx := setupForestTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleForest(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/forest", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()

	if !strings.Contains(body, "Parent Epic") {
		t.Error("response missing 'Parent Epic'")
	}

	if !strings.Contains(body, "mermaid") {
		t.Error("response missing mermaid class/script")
	}
}

func TestHandleForestHTMX(t *testing.T) {
	idx := setupForestTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleForest(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/forest", nil)
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	body := w.Body.String()

	if !strings.Contains(body, "Parent Epic") {
		t.Error("HTMX response missing 'Parent Epic'")
	}

	if strings.Contains(body, "<html") {
		t.Error("HTMX response should not contain full HTML layout")
	}
}

func TestForestPageHTMLStructure(t *testing.T) {
	idx := setupForestTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleForest(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/forest", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertTitle(t, doc, "Forest - Taskboard")
	testhelpers.AssertElementExists(t, doc, "pre.mermaid")
}

func TestForestPageNavLinks(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleForest(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/forest", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertElementExists(t, doc, `a[href="/"]`)
	testhelpers.AssertElementExists(t, doc, `a[href="/forest"]`)
}

func TestBuildMermaidDefNodes(t *testing.T) {
	tasks := []*task.Task{
		{ID: "abc", Title: "First Task", Fields: map[string]string{"type": "bug", "status": "open"}},
		{ID: "def", Title: "Second Task", Fields: map[string]string{"type": "feature", "status": "done"}},
	}

	def := buildMermaidDef(tasks, nil)

	if !strings.Contains(def, `abc(["First Task`) {
		t.Errorf("missing stadium node abc in:\n%s", def)
	}

	if !strings.Contains(def, `def(["Second Task`) {
		t.Errorf("missing stadium node def in:\n%s", def)
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

	if !strings.Contains(def, "graph TD") {
		t.Errorf("empty graph should still have header, got:\n%s", def)
	}
}

func TestNodeLabelWithTypeAndStatus(t *testing.T) {
	tk := &task.Task{
		Title:  "My Task",
		Fields: map[string]string{"type": "bug", "status": "open"},
	}

	label := nodeLabel(tk, nil)

	if !strings.Contains(label, "My Task") {
		t.Errorf("label missing title: %s", label)
	}

	if !strings.Contains(label, "bug") {
		t.Errorf("label missing type: %s", label)
	}

	if strings.Contains(label, "open") {
		t.Errorf("label should not contain status: %s", label)
	}

	if !strings.Contains(label, "<br/>") {
		t.Errorf("label missing line break: %s", label)
	}

	if strings.Contains(label, "border-radius") {
		t.Errorf("label should not contain chip markup: %s", label)
	}
}

func TestNodeLabelNoMeta(t *testing.T) {
	tk := &task.Task{
		Title:  "Plain Task",
		Fields: map[string]string{},
	}

	label := nodeLabel(tk, nil)

	if label != "Plain Task" {
		t.Errorf("nodeLabel() = %q, want %q", label, "Plain Task")
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
		Title:  "Task",
		Fields: map[string]string{"status": "open"},
	}

	label := nodeLabel(tk, nil)

	if label != "Task" {
		t.Errorf("nodeLabel() = %q, want %q (status should not appear in label)", label, "Task")
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
	}
	tasks := []*task.Task{
		{ID: "a", Title: "Open", Fields: map[string]string{"status": "open"}},
		{ID: "b", Title: "Done", Fields: map[string]string{"status": "done"}},
		{ID: "c", Title: "None", Fields: map[string]string{}},
	}

	def := buildMermaidDef(tasks, columns)

	if !strings.Contains(def, "style a fill:#22c55e") {
		t.Errorf("missing style for node a in:\n%s", def)
	}

	if !strings.Contains(def, "style b fill:#6b7280") {
		t.Errorf("missing style for node b in:\n%s", def)
	}

	if strings.Contains(def, "style c fill:") {
		t.Errorf("node c without status should have no style in:\n%s", def)
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
