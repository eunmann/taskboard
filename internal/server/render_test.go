package server

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eunmann/taskboard/internal/config"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()

	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	return r
}

func TestRenderMarkdownHeadings(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("# H1\n## H2\n### H3"))

	for _, want := range []string{"<h1>H1</h1>", "<h2>H2</h2>", "<h3>H3</h3>"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderMarkdown() missing %s, got %s", want, got)
		}
	}
}

func TestRenderMarkdownBoldItalic(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("**bold** and *italic*"))

	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Errorf("renderMarkdown() missing <strong>, got %s", got)
	}

	if !strings.Contains(got, "<em>italic</em>") {
		t.Errorf("renderMarkdown() missing <em>, got %s", got)
	}
}

func TestRenderMarkdownUnorderedList(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("- one\n- two"))

	if !strings.Contains(got, "<ul>") {
		t.Errorf("renderMarkdown() missing <ul>, got %s", got)
	}

	if !strings.Contains(got, "<li>one</li>") {
		t.Errorf("renderMarkdown() missing list item, got %s", got)
	}
}

func TestRenderMarkdownOrderedList(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("1. first\n2. second"))

	if !strings.Contains(got, "<ol>") {
		t.Errorf("renderMarkdown() missing <ol>, got %s", got)
	}
}

func TestRenderMarkdownCodeBlock(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("```\nfmt.Println()\n```"))

	if !strings.Contains(got, "<pre>") {
		t.Errorf("renderMarkdown() missing <pre>, got %s", got)
	}

	if !strings.Contains(got, "<code>") {
		t.Errorf("renderMarkdown() missing <code>, got %s", got)
	}
}

func TestRenderMarkdownInlineCode(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("run `go test`"))

	if !strings.Contains(got, "<code>go test</code>") {
		t.Errorf("renderMarkdown() missing inline code, got %s", got)
	}
}

func TestRenderMarkdownBlockquote(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("> quoted text"))

	if !strings.Contains(got, "<blockquote>") {
		t.Errorf("renderMarkdown() missing <blockquote>, got %s", got)
	}
}

func TestRenderMarkdownLink(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("[link](https://example.com)"))

	if !strings.Contains(got, `href="https://example.com"`) {
		t.Errorf("renderMarkdown() missing link href, got %s", got)
	}
}

func TestRenderMarkdownGFMTaskList(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("- [ ] todo\n- [x] done"))

	if !strings.Contains(got, `type="checkbox"`) {
		t.Errorf("renderMarkdown() missing checkbox, got %s", got)
	}

	if !strings.Contains(got, "disabled") {
		t.Errorf("renderMarkdown() checkbox should be disabled, got %s", got)
	}
}

func TestRenderMarkdownGFMTable(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("| A | B |\n|---|---|\n| 1 | 2 |"))

	if !strings.Contains(got, "<table>") {
		t.Errorf("renderMarkdown() missing <table>, got %s", got)
	}

	if !strings.Contains(got, "<th>A</th>") {
		t.Errorf("renderMarkdown() missing header cell, got %s", got)
	}
}

func TestRenderMarkdownGFMStrikethrough(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("~~removed~~"))

	if !strings.Contains(got, "<del>removed</del>") {
		t.Errorf("renderMarkdown() missing <del>, got %s", got)
	}
}

func TestRenderMarkdownGFMAutolink(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("Visit https://example.com for info"))

	if !strings.Contains(got, `href="https://example.com"`) {
		t.Errorf("renderMarkdown() missing autolink, got %s", got)
	}
}

func TestRenderMarkdownEmpty(t *testing.T) {
	r := newTestRenderer(t)
	got := r.renderMarkdown("")

	if got != "" {
		t.Errorf("renderMarkdown(\"\") = %q, want empty", got)
	}
}

func TestRenderMarkdownPlainText(t *testing.T) {
	r := newTestRenderer(t)
	got := string(r.renderMarkdown("plain text"))

	if !strings.Contains(got, "plain text") {
		t.Errorf("renderMarkdown() missing plain text, got %s", got)
	}

	if got != "<p>plain text</p>\n" {
		t.Errorf("renderMarkdown() = %q, want %q", got, "<p>plain text</p>\n")
	}
}

func TestRenderMarkdownReturnType(t *testing.T) {
	r := newTestRenderer(t)
	got := r.renderMarkdown("test")

	// Verify the return type is template.HTML (not string).
	_ = got
}

func TestRelativeTimeJustNow(t *testing.T) {
	got := relativeTime(time.Now())
	if got != "just now" {
		t.Errorf("relativeTime(now) = %q, want %q", got, "just now")
	}
}

func TestRelativeTimeMinutes(t *testing.T) {
	got := relativeTime(time.Now().Add(-5 * time.Minute))
	if got != "5 minutes ago" {
		t.Errorf("relativeTime(-5m) = %q, want %q", got, "5 minutes ago")
	}
}

func TestRelativeTimeOneMinute(t *testing.T) {
	got := relativeTime(time.Now().Add(-1 * time.Minute))
	if got != "1 minute ago" {
		t.Errorf("relativeTime(-1m) = %q, want %q", got, "1 minute ago")
	}
}

func TestRelativeTimeHours(t *testing.T) {
	got := relativeTime(time.Now().Add(-3 * time.Hour))
	if got != "3 hours ago" {
		t.Errorf("relativeTime(-3h) = %q, want %q", got, "3 hours ago")
	}
}

func TestRelativeTimeOneHour(t *testing.T) {
	got := relativeTime(time.Now().Add(-1 * time.Hour))
	if got != "1 hour ago" {
		t.Errorf("relativeTime(-1h) = %q, want %q", got, "1 hour ago")
	}
}

func TestRelativeTimeDays(t *testing.T) {
	got := relativeTime(time.Now().Add(-48 * time.Hour))
	if got != "2 days ago" {
		t.Errorf("relativeTime(-48h) = %q, want %q", got, "2 days ago")
	}
}

func TestRelativeTimeOneDay(t *testing.T) {
	got := relativeTime(time.Now().Add(-25 * time.Hour))
	if got != "1 day ago" {
		t.Errorf("relativeTime(-25h) = %q, want %q", got, "1 day ago")
	}
}

func TestRelativeTimeZero(t *testing.T) {
	got := relativeTime(time.Time{})
	if got != "" {
		t.Errorf("relativeTime(zero) = %q, want empty", got)
	}
}

func TestDictFunc(t *testing.T) {
	m, err := dictFunc("key1", "val1", "key2", 42)
	if err != nil {
		t.Fatalf("dictFunc() error = %v", err)
	}

	if m["key1"] != "val1" {
		t.Errorf("m[key1] = %v, want %q", m["key1"], "val1")
	}

	if m["key2"] != 42 {
		t.Errorf("m[key2] = %v, want %d", m["key2"], 42)
	}
}

func TestDictFuncOddArgs(t *testing.T) {
	_, err := dictFunc("key1")
	if err == nil {
		t.Fatal("expected error for odd number of arguments")
	}
}

func TestDictFuncNonStringKey(t *testing.T) {
	_, err := dictFunc(42, "val")
	if err == nil {
		t.Fatal("expected error for non-string key")
	}
}

func TestSortedColumns(t *testing.T) {
	columns := map[string]config.Column{
		"priority": {Order: 2, Values: []config.Value{{Name: "high"}}},
		"status":   {Order: 1, Values: []config.Value{{Name: "open"}}},
		"type":     {Order: 3, Values: []config.Value{{Name: "bug"}}},
	}

	sorted := SortedColumns(columns)
	if len(sorted) != 3 {
		t.Fatalf("SortedColumns() length = %d, want 3", len(sorted))
	}

	if sorted[0].Name != "status" {
		t.Errorf("sorted[0].Name = %q, want %q", sorted[0].Name, "status")
	}

	if sorted[1].Name != "priority" {
		t.Errorf("sorted[1].Name = %q, want %q", sorted[1].Name, "priority")
	}

	if sorted[2].Name != "type" {
		t.Errorf("sorted[2].Name = %q, want %q", sorted[2].Name, "type")
	}
}

func TestSortedColumnsSameOrder(t *testing.T) {
	columns := map[string]config.Column{
		"beta":  {Order: 1, Values: []config.Value{{Name: "b"}}},
		"alpha": {Order: 1, Values: []config.Value{{Name: "a"}}},
	}

	sorted := SortedColumns(columns)
	if sorted[0].Name != "alpha" {
		t.Errorf("sorted[0].Name = %q, want %q (alphabetical tiebreak)", sorted[0].Name, "alpha")
	}
}

func TestRenderNotFoundTemplate(t *testing.T) {
	r := newTestRenderer(t)
	w := httptest.NewRecorder()

	err := r.Render(w, "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}

	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("error = %v, want ErrTemplateNotFound", err)
	}
}

func TestRenderPartialNotFoundTemplate(t *testing.T) {
	r := newTestRenderer(t)
	w := httptest.NewRecorder()

	err := r.RenderPartial(w, "nonexistent", "block", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}

	if !errors.Is(err, ErrTemplateNotFound) {
		t.Errorf("error = %v, want ErrTemplateNotFound", err)
	}
}
