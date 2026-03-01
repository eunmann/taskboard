package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eunmann/taskboard/internal/index"
	"github.com/eunmann/taskboard/internal/web/testhelpers"
)

func setupTestIndex(t *testing.T) *index.Index {
	t.Helper()

	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Test Project
columns:
  status:
    order: 1
    values:
      - name: open
        color: green
      - name: closed
        color: red
  priority:
    order: 2
    values:
      - name: low
        color: gray
      - name: high
        color: red
`)

	writeTestFile(t, dir, "abc12345.yaml", `
title: First task
status: open
priority: high
tags: [backend]
description: |
  This is a **bold** description.

  ## Subheading

  - [ ] todo item
  - [x] done item
`)

	writeTestFile(t, dir, "def67890.yaml", `
title: Second task
status: closed
priority: low
tags: [frontend]
`)

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return idx
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestHandleList(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "First task") {
		t.Error("response missing 'First task'")
	}

	if !containsString(body, "Second task") {
		t.Error("response missing 'Second task'")
	}
}

func TestHandleListHTMX(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("ETag") == "" {
		t.Error("missing ETag header on HTMX response")
	}
}

func TestHandleDetail(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "First task") {
		t.Error("response missing task title")
	}

	if !containsString(body, "<strong>bold</strong>") {
		t.Error("markdown not rendered")
	}

	if !containsString(body, "<h2>Subheading</h2>") {
		t.Error("GFM heading not rendered")
	}

	if !containsString(body, `type="checkbox"`) {
		t.Error("GFM task list checkbox not rendered")
	}
}

func TestHandleDetailNotFound(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleTablePartial(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("ETag") == "" {
		t.Error("missing ETag header")
	}
}

func TestHandleTablePartial304(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table", nil)
	req.Header.Set("If-None-Match", `"1"`)

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotModified {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotModified)
	}
}

func TestTablePartialSearch(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "First task") {
		t.Error("response missing 'First task'")
	}

	if containsString(body, "Second task") {
		t.Error("response should not contain 'Second task'")
	}
}

func TestTablePartialFilter(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?status=open", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "First task") {
		t.Error("response missing open task 'First task'")
	}

	if containsString(body, "Second task") {
		t.Error("response should not contain closed task 'Second task'")
	}
}

func TestTablePartialSort(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?sort=title", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	firstIdx := strings.Index(body, "First task")
	secondIdx := strings.Index(body, "Second task")

	if firstIdx < 0 || secondIdx < 0 {
		t.Fatal("response missing expected tasks")
	}

	if firstIdx >= secondIdx {
		t.Error("'First task' should appear before 'Second task' with sort=title (ascending)")
	}
}

func TestTablePartialCombined(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?q=task&status=open&sort=title", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "First task") {
		t.Error("response missing 'First task' (matches q=task + status=open)")
	}

	if containsString(body, "Second task") {
		t.Error("response should not contain 'Second task' (status=closed)")
	}
}

func TestTablePartialETagVariesWithQuery(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req1 := httptest.NewRequest(http.MethodGet, "/partials/table", nil)
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	w2 := httptest.NewRecorder()
	handler(w2, req2)

	etag1 := w1.Header().Get("ETag")
	etag2 := w2.Header().Get("ETag")

	if etag1 == "" || etag2 == "" {
		t.Fatal("missing ETag header(s)")
	}

	if etag1 == etag2 {
		t.Errorf("ETags should differ for different queries: both = %s", etag1)
	}
}

func TestTablePartial304SameQuery(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	// First request to get the ETag.
	req1 := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag header")
	}

	// Second request with same query and matching ETag.
	req2 := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	req2.Header.Set("If-None-Match", etag)

	w2 := httptest.NewRecorder()
	handler(w2, req2)

	if w2.Code != http.StatusNotModified {
		t.Errorf("status = %d, want %d", w2.Code, http.StatusNotModified)
	}
}

func TestTablePartial200DifferentQuery(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	// Get ETag for empty query.
	req1 := httptest.NewRequest(http.MethodGet, "/partials/table", nil)
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag header")
	}

	// Request with different query but old ETag should return 200.
	req2 := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	req2.Header.Set("If-None-Match", etag)

	w2 := httptest.NewRecorder()
	handler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (different query should not 304)", w2.Code, http.StatusOK)
	}
}

func TestTablePartialPushUrl(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	pushURL := w.Header().Get("HX-Push-Url")
	if pushURL == "" {
		t.Error("missing HX-Push-Url header on non-poll request")
	}

	if !containsString(pushURL, "q=First") {
		t.Errorf("HX-Push-Url = %q, want it to contain q=First", pushURL)
	}
}

func TestTablePartialPollNoPushUrl(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?_poll=1&q=First", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Header().Get("HX-Push-Url") != "" {
		t.Error("poll request should not have HX-Push-Url header")
	}
}

func TestTablePartialSearchPreservesFilters(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?q=task&status=open", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, `name="status"`) {
		t.Error("response should contain hidden input for status filter")
	}

	if !containsString(body, `value="open"`) {
		t.Error("response should preserve status=open filter value")
	}
}

func TestLayoutHistoryElt(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	body := w.Body.String()

	if !containsString(body, "hx-history-elt") {
		t.Error("layout <main> missing hx-history-elt attribute")
	}
}

func TestLayoutHTMXConfig(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	body := w.Body.String()

	if !containsString(body, `htmx-config`) {
		t.Error("layout missing htmx-config meta tag")
	}

	if !containsString(body, `refreshOnHistoryMiss`) {
		t.Error("htmx-config missing refreshOnHistoryMiss setting")
	}
}

func TestDetailBackLink(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()

	handler(w, req)

	body := w.Body.String()

	if !containsString(body, "history.back()") {
		t.Error("back link missing history.back() onclick")
	}

	if containsString(body, `hx-get="/"`) {
		t.Error("back link should not have hx-get attribute")
	}

	if containsString(body, "hx-push-url") {
		t.Error("back link should not have hx-push-url attribute")
	}
}

func TestFilterCheckboxAppliesImmediately(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	body := w.Body.String()

	if !containsString(body, "applyFilters(filters)") {
		t.Error("response missing applyFilters call on checkbox change")
	}

	if containsString(body, "tb-apply") {
		t.Error("response should not contain separate Filter button (class tb-apply)")
	}
}

func TestFilterDropdownNoIndividualApply(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	body := w.Body.String()

	if containsString(body, "data-base") {
		t.Error("response should not contain per-dropdown data-base attribute")
	}
}

func TestHandleDetailHTMX(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "First task") {
		t.Error("HTMX detail response missing task title")
	}
}

func TestCollectTags(t *testing.T) {
	idx := setupTestIndex(t)
	tags := collectTags(idx.List())

	if len(tags) != 2 {
		t.Fatalf("collectTags() length = %d, want 2", len(tags))
	}

	// Should be sorted.
	if tags[0] != "backend" || tags[1] != "frontend" {
		t.Errorf("collectTags() = %v, want [backend frontend]", tags)
	}
}

func TestCollectTagsEmpty(t *testing.T) {
	tags := collectTags(nil)
	if len(tags) != 0 {
		t.Errorf("collectTags(nil) = %v, want empty", tags)
	}
}

func TestIsHTMX(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if isHTMX(req) {
		t.Error("isHTMX() = true for normal request, want false")
	}

	req.Header.Set("HX-Request", "true")

	if !isHTMX(req) {
		t.Error("isHTMX() = false for HTMX request, want true")
	}
}

func TestComputeETag(t *testing.T) {
	q := Query{Filters: map[string][]string{}, Sort: "updated", Desc: true}
	etag := computeETag(1, q)

	if etag != `"1"` {
		t.Errorf("computeETag(1, default) = %q, want %q", etag, `"1"`)
	}

	q2 := Query{Search: "test", Filters: map[string][]string{}, Sort: "updated", Desc: true}
	etag2 := computeETag(1, q2)

	if etag2 == etag {
		t.Error("ETags should differ with different queries")
	}
}

func TestListPageHTMLStructure(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertTitle(t, doc, "Tasks - Taskboard")
	testhelpers.AssertElementExists(t, doc, "main[hx-history-elt]")
	testhelpers.AssertElementExists(t, doc, `meta[name="htmx-config"]`)
	testhelpers.AssertElementExists(t, doc, `input[name="q"]`)
	testhelpers.AssertElementExists(t, doc, "#task-table")
	testhelpers.AssertTableRowCount(t, doc, "#task-table table", 2)
}

func TestListPageSearchInput(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleList(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertHTMXAttr(t, doc, `input[name="q"]`, "hx-get", "/partials/table")
	testhelpers.AssertHTMXAttr(t, doc, `input[name="q"]`, "hx-target", "#task-table")
}

func TestDetailPageHTMLStructure(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")

	w := httptest.NewRecorder()
	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertTitle(t, doc, "First task - Taskboard")
	testhelpers.AssertElementExists(t, doc, ".prose")
	testhelpers.AssertElementExists(t, doc, "h2")
	testhelpers.AssertElementExists(t, doc, `input[type="checkbox"]`)
}

func TestTablePartialHTMLStructure(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertElementExists(t, doc, "#task-table")
	testhelpers.AssertTableRowCount(t, doc, "#task-table table", 2)
	testhelpers.AssertElementExists(t, doc, "thead th")

	// Verify rows are clickable with HTMX.
	testhelpers.AssertHTMXAttr(t, doc, "tbody tr:first-child", "hx-push-url", "true")
}

func TestTablePartialSearchRowCount(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?q=First", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertTableRowCount(t, doc, "#task-table table", 1)
}

func TestTablePartialEmptyState(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleTablePartial(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/partials/table?q=nonexistent", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	// When no tasks match, there should be no table, just the empty state.
	testhelpers.AssertElementCount(t, doc, "table", 0)
	testhelpers.AssertElementExists(t, doc, "#task-table")
}

func TestDetailPageColumnBadges(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")

	w := httptest.NewRecorder()
	handler(w, req)

	doc := testhelpers.ParseHTML(t, w)

	// Task has status=open, priority=high — column badges are spans.
	testhelpers.AssertElementCount(t, doc, "dl span.inline-flex", 2)
	// Tag=backend is a clickable link.
	testhelpers.AssertElementCount(t, doc, "dl a.inline-flex", 1)
}

func setupTestIndexWithRefs(t *testing.T) *index.Index {
	t.Helper()

	dir := t.TempDir()

	writeTestFile(t, dir, "config.yaml", `
project: Test Project
columns:
  status:
    order: 1
    values:
      - name: open
        color: green
`)

	writeTestFile(t, dir, "parent01.yaml", `
title: Parent task
status: open
`)

	writeTestFile(t, dir, "child001.yaml", `
title: Child task
status: open
refs:
  - type: parent
    id: parent01
`)

	writeTestFile(t, dir, "blocked1.yaml", `
title: Blocked task
status: open
refs:
  - type: blocked-by
    id: parent01
`)

	writeTestFile(t, dir, "related1.yaml", `
title: Related task
status: open
refs:
  - type: relates-to
    id: parent01
`)

	idx, err := index.New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return idx
}

func TestDetailReverseRefs(t *testing.T) {
	idx := setupTestIndexWithRefs(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/parent01", nil)
	req.SetPathValue("id", "parent01")

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	if !containsString(body, "Referenced by") {
		t.Error("response missing 'Referenced by' section")
	}

	if !containsString(body, "child:") {
		t.Error("response missing 'child:' reverse label")
	}

	if !containsString(body, "blocks:") {
		t.Error("response missing 'blocks:' reverse label")
	}

	if !containsString(body, "related:") {
		t.Error("response missing 'related:' reverse label")
	}

	if !containsString(body, "Child task") {
		t.Error("response missing child task title")
	}

	if !containsString(body, "Blocked task") {
		t.Error("response missing blocked task title")
	}

	if !containsString(body, "Related task") {
		t.Error("response missing related task title")
	}
}

func TestDetailNoReverseRefs(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")

	w := httptest.NewRecorder()

	handler(w, req)

	body := w.Body.String()

	if containsString(body, "Referenced by") {
		t.Error("response should not contain 'Referenced by' when no reverse refs exist")
	}
}

func TestDetailWithRefsMermaidGraph(t *testing.T) {
	idx := setupTestIndexWithRefs(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/parent01", nil)
	req.SetPathValue("id", "parent01")

	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertElementExists(t, doc, "pre.mermaid")
}

func TestDetailWithoutRefsMermaidGraph(t *testing.T) {
	idx := setupTestIndex(t)

	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	handler := handleDetail(idx, renderer)

	req := httptest.NewRequest(http.MethodGet, "/task/abc12345", nil)
	req.SetPathValue("id", "abc12345")

	w := httptest.NewRecorder()

	handler(w, req)

	testhelpers.AssertStatus(t, w, http.StatusOK)

	doc := testhelpers.ParseHTML(t, w)

	testhelpers.AssertElementCount(t, doc, "pre.mermaid", 0)
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
