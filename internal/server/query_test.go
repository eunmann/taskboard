package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/task"
)

const (
	sortUpdated = "updated"
	sortTitle   = "title"
)

func testCols() map[string]config.Column {
	return map[string]config.Column{
		"status": {
			Order: 1,
			Values: []config.Value{
				{Name: "open", Color: "#22c55e"},
				{Name: "closed", Color: "#ef4444"},
			},
		},
		"priority": {
			Order: 2,
			Values: []config.Value{
				{Name: "low", Color: "#6b7280"},
				{Name: "high", Color: "#f97316"},
			},
		},
	}
}

func TestParseQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?q=search+term&status=open,closed&priority=high", nil)

	q := parseQuery(req, testCols())

	if q.Search != "search term" {
		t.Errorf("Search = %q, want %q", q.Search, "search term")
	}

	if len(q.Filters["status"]) != 2 {
		t.Errorf("Filters[status] length = %d, want 2", len(q.Filters["status"]))
	}

	if len(q.Filters["priority"]) != 1 {
		t.Errorf("Filters[priority] length = %d, want 1", len(q.Filters["priority"]))
	}
}

func TestParseQueryEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	q := parseQuery(req, testCols())

	if q.Search != "" {
		t.Errorf("Search = %q, want empty", q.Search)
	}

	if len(q.Filters) != 0 {
		t.Errorf("Filters = %v, want empty", q.Filters)
	}

	if q.Sort != sortUpdated {
		t.Errorf("Sort = %q, want %q", q.Sort, sortUpdated)
	}

	if !q.Desc {
		t.Error("Desc = false, want true (default sort is -updated)")
	}
}

func TestParseQuerySortDesc(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?sort=-created", nil)

	q := parseQuery(req, testCols())

	if q.Sort != "created" {
		t.Errorf("Sort = %q, want %q", q.Sort, "created")
	}

	if !q.Desc {
		t.Error("Desc = false, want true")
	}
}

func TestParseQuerySortAsc(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?sort=title", nil)

	q := parseQuery(req, testCols())

	if q.Sort != sortTitle {
		t.Errorf("Sort = %q, want %q", q.Sort, sortTitle)
	}

	if q.Desc {
		t.Error("Desc = true, want false")
	}
}

func TestApplySearch(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Title: "Fix login bug", Tags: []string{"backend"}, Fields: map[string]string{}},
		{ID: "2", Title: "Add dark mode", Tags: []string{"frontend"}, Fields: map[string]string{}},
		{ID: "3", Title: "Update API docs", Tags: []string{"docs"}, Fields: map[string]string{}},
	}

	result := applySearch(tasks, "login")
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("search 'login' returned %d results, want 1", len(result))
	}

	result = applySearch(tasks, "backend")
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("search 'backend' (tag) returned %d results, want 1", len(result))
	}
}

func TestFilterByColumn(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Fields: map[string]string{"status": "open"}},
		{ID: "2", Fields: map[string]string{"status": "closed"}},
		{ID: "3", Fields: map[string]string{"status": "open"}},
	}

	result := filterByColumn(tasks, "status", []string{"open"})
	if len(result) != 2 {
		t.Errorf("filter status=open returned %d, want 2", len(result))
	}

	result = filterByColumn(tasks, "status", []string{"closed"})
	if len(result) != 1 {
		t.Errorf("filter status=closed returned %d, want 1", len(result))
	}
}

func TestApplyFilters(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Title: "Task A", Fields: map[string]string{"status": "open", "priority": "high"}},
		{ID: "2", Title: "Task B", Fields: map[string]string{"status": "closed", "priority": "low"}},
		{ID: "3", Title: "Task C", Fields: map[string]string{"status": "open", "priority": "low"}},
	}

	q := Query{
		Search:  "task",
		Filters: map[string][]string{"status": {"open"}},
	}

	result := applyFilters(tasks, q)
	if len(result) != 2 {
		t.Errorf("combined filter returned %d, want 2", len(result))
	}
}

func TestSortParam(t *testing.T) {
	if got := sortParam(sortTitle, sortUpdated, true); got != sortTitle {
		t.Errorf("sortParam(title, updated, true) = %q, want %q", got, sortTitle)
	}

	if got := sortParam(sortUpdated, sortUpdated, true); got != sortUpdated {
		t.Errorf("sortParam(updated, updated, true) = %q, want %q", got, sortUpdated)
	}

	if got := sortParam(sortUpdated, sortUpdated, false); got != "-updated" {
		t.Errorf("sortParam(updated, updated, false) = %q, want %q", got, "-updated")
	}
}

func TestQueryParams(t *testing.T) {
	q := Query{
		Search:  "bug",
		Filters: map[string][]string{"status": {"open"}},
		Sort:    "title",
		Desc:    false,
	}

	got := queryParams(q)
	if got != "q=bug&status=open&sort=title" {
		t.Errorf("queryParams() = %q, want %q", got, "q=bug&status=open&sort=title")
	}
}

func TestQueryParamsDefaultSort(t *testing.T) {
	q := Query{
		Filters: map[string][]string{},
		Sort:    "updated",
		Desc:    true,
	}

	got := queryParams(q)
	if got != "" {
		t.Errorf("queryParams() with default sort = %q, want empty", got)
	}
}

func TestQueryParamsWithSort(t *testing.T) {
	q := Query{
		Search:  "fix",
		Filters: map[string][]string{"priority": {"high"}},
		Sort:    "updated",
		Desc:    true,
	}

	got := queryParamsWithSort(q, "title")
	if got != "q=fix&priority=high&sort=title" {
		t.Errorf("queryParamsWithSort() = %q, want %q", got, "q=fix&priority=high&sort=title")
	}
}

func TestQueryParamsWithout(t *testing.T) {
	q := Query{
		Search: "test",
		Filters: map[string][]string{
			"status":   {"open"},
			"priority": {"high"},
		},
		Sort: "title",
		Desc: false,
	}

	got := queryParamsWithout(q, "status")
	if got != "q=test&priority=high&sort=title&" {
		t.Errorf("queryParamsWithout() = %q, want %q", got, "q=test&priority=high&sort=title&")
	}
}

func TestQueryParamsWithoutEmpty(t *testing.T) {
	q := Query{
		Filters: map[string][]string{},
		Sort:    "updated",
		Desc:    true,
	}

	got := queryParamsWithout(q, "status")
	if got != "" {
		t.Errorf("queryParamsWithout() empty = %q, want empty", got)
	}
}

func TestQueryParamsURLEncoding(t *testing.T) {
	q := Query{
		Search:  "foo bar",
		Filters: map[string][]string{},
		Sort:    "updated",
		Desc:    true,
	}

	got := queryParams(q)
	if got != "q=foo+bar" {
		t.Errorf("queryParams() = %q, want %q", got, "q=foo+bar")
	}
}

func TestQueryParamsFilterURLEncoding(t *testing.T) {
	q := Query{
		Filters: map[string][]string{"status": {"in-progress", "code review"}},
		Sort:    "updated",
		Desc:    true,
	}

	got := queryParams(q)
	want := "status=in-progress,code+review"

	if got != want {
		t.Errorf("queryParams() = %q, want %q", got, want)
	}
}

func TestApplySortByCreated(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	tasks := []*task.Task{
		{ID: "1", Title: "B", Created: t2, Fields: map[string]string{}},
		{ID: "2", Title: "A", Created: t1, Fields: map[string]string{}},
	}

	q := Query{Sort: "created", Desc: false, Filters: map[string][]string{}}
	result := applySort(tasks, q)

	if result[0].ID != "2" {
		t.Errorf("sort by created asc: first = %q, want %q", result[0].ID, "2")
	}
}

func TestApplySortByFieldValue(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Fields: map[string]string{"priority": "high"}},
		{ID: "2", Fields: map[string]string{"priority": "critical"}},
	}

	q := Query{Sort: "priority", Desc: false, Filters: map[string][]string{}}
	result := applySort(tasks, q)

	if result[0].ID != "2" {
		t.Errorf("sort by field asc: first = %q, want %q", result[0].ID, "2")
	}
}

func TestApplySortDesc(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Title: "A", Fields: map[string]string{}},
		{ID: "2", Title: "B", Fields: map[string]string{}},
	}

	q := Query{Sort: sortTitle, Desc: true, Filters: map[string][]string{}}
	result := applySort(tasks, q)

	if result[0].ID != "2" {
		t.Errorf("sort by title desc: first = %q, want %q", result[0].ID, "2")
	}
}

func TestMatchesSearchByID(t *testing.T) {
	tasks := []*task.Task{
		{ID: "abc123", Title: "Something", Fields: map[string]string{}},
		{ID: "def456", Title: "Other", Fields: map[string]string{}},
	}

	result := applySearch(tasks, "abc")
	if len(result) != 1 || result[0].ID != "abc123" {
		t.Errorf("search by ID returned %d results, want 1", len(result))
	}
}

func TestMatchesSearchByDescription(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Title: "Task", Description: "Fix the login bug", Fields: map[string]string{}},
		{ID: "2", Title: "Other", Description: "Add feature", Fields: map[string]string{}},
	}

	result := applySearch(tasks, "login")
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("search by description returned %d results, want 1", len(result))
	}
}

func TestMatchesSearchByFieldValue(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Title: "Task", Fields: map[string]string{"status": "open"}},
		{ID: "2", Title: "Other", Fields: map[string]string{"status": "closed"}},
	}

	result := applySearch(tasks, "open")
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("search by field value returned %d results, want 1", len(result))
	}
}

func TestMatchesSearchCaseInsensitive(t *testing.T) {
	tasks := []*task.Task{
		{ID: "1", Title: "FIX LOGIN", Fields: map[string]string{}},
	}

	result := applySearch(tasks, "fix login")
	if len(result) != 1 {
		t.Error("case-insensitive search failed")
	}
}

func TestSortIndicator(t *testing.T) {
	if got := sortIndicator(sortTitle, sortUpdated, true); got != "" {
		t.Errorf("sortIndicator(title, updated, true) = %q, want empty", got)
	}

	if got := sortIndicator(sortUpdated, sortUpdated, true); got != " ↓" {
		t.Errorf("sortIndicator(updated, updated, true) = %q, want %q", got, " ↓")
	}

	if got := sortIndicator(sortUpdated, sortUpdated, false); got != " ↑" {
		t.Errorf("sortIndicator(updated, updated, false) = %q, want %q", got, " ↑")
	}
}
