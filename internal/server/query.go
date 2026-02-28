package server

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/task"
)

// Query holds parsed request parameters for task filtering and sorting.
type Query struct {
	Search  string
	Filters map[string][]string
	Sort    string
	Desc    bool
}

func parseQuery(r *http.Request, columns map[string]config.Column) Query {
	q := Query{
		Search:  strings.TrimSpace(r.URL.Query().Get("q")),
		Filters: make(map[string][]string),
	}

	sortParam := r.URL.Query().Get("sort")
	if sortParam != "" {
		if after, ok := strings.CutPrefix(sortParam, "-"); ok {
			q.Sort = after
			q.Desc = true
		} else {
			q.Sort = sortParam
		}
	} else {
		q.Sort = "updated"
		q.Desc = true
	}

	for colName := range columns {
		val := r.URL.Query().Get(colName)
		if val == "" {
			continue
		}

		parts := strings.Split(val, ",")

		var filterVals []string

		for _, v := range parts {
			v = strings.TrimSpace(v)
			if v != "" {
				filterVals = append(filterVals, v)
			}
		}

		if len(filterVals) > 0 {
			q.Filters[colName] = filterVals
		}
	}

	return q
}

// sortParam returns the URL sort parameter for toggling a column.
func sortParam(column, currentSort string, currentDesc bool) string {
	if column == currentSort {
		if currentDesc {
			return column
		}

		return "-" + column
	}

	return column
}

// sortIndicator returns the sort direction indicator for a column header.
func sortIndicator(column, currentSort string, currentDesc bool) string {
	if column != currentSort {
		return ""
	}

	if currentDesc {
		return " ↓"
	}

	return " ↑"
}

func applyFilters(tasks []*task.Task, q Query) []*task.Task {
	result := tasks

	if q.Search != "" {
		result = applySearch(result, q.Search)
	}

	for col, values := range q.Filters {
		result = filterByColumn(result, col, values)
	}

	return result
}

func applySort(tasks []*task.Task, q Query) []*task.Task {
	sorted := make([]*task.Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool

		switch q.Sort {
		case "title":
			less = sorted[i].Title < sorted[j].Title
		case "updated":
			less = sorted[i].Updated.Before(sorted[j].Updated)
		case "created":
			less = sorted[i].Created.Before(sorted[j].Created)
		default:
			vi := sorted[i].Fields[q.Sort]
			vj := sorted[j].Fields[q.Sort]
			less = vi < vj
		}

		if q.Desc {
			return !less
		}

		return less
	})

	return sorted
}

func applySearch(tasks []*task.Task, search string) []*task.Task {
	lower := strings.ToLower(search)

	var result []*task.Task

	for _, t := range tasks {
		if matchesSearch(t, lower) {
			result = append(result, t)
		}
	}

	return result
}

func matchesSearch(t *task.Task, lower string) bool {
	if strings.Contains(strings.ToLower(t.Title), lower) {
		return true
	}

	if strings.Contains(strings.ToLower(t.ID), lower) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Description), lower) {
		return true
	}

	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), lower) {
			return true
		}
	}

	for _, val := range t.Fields {
		if strings.Contains(strings.ToLower(val), lower) {
			return true
		}
	}

	return false
}

// queryParams serializes a Query to URL query parameters (e.g. "q=foo&status=open&sort=-updated").
func queryParams(q Query) string {
	return buildQueryString(q, "")
}

// queryParamsWithSort serializes a Query with sort toggled for the given column.
func queryParamsWithSort(q Query, column string) string {
	toggled := Query{
		Search:  q.Search,
		Filters: q.Filters,
		Sort:    column,
		Desc:    false,
	}

	sp := sortParam(column, q.Sort, q.Desc)
	if after, ok := strings.CutPrefix(sp, "-"); ok {
		toggled.Sort = after
		toggled.Desc = true
	} else {
		toggled.Sort = sp
	}

	return buildQueryString(toggled, "")
}

// queryParamsWithout serializes a Query excluding a specific filter, with trailing "&" if non-empty.
func queryParamsWithout(q Query, exclude string) string {
	filtered := Query{
		Search:  q.Search,
		Filters: make(map[string][]string),
		Sort:    q.Sort,
		Desc:    q.Desc,
	}

	for k, v := range q.Filters {
		if k != exclude {
			filtered.Filters[k] = v
		}
	}

	s := buildQueryString(filtered, "")
	if s != "" {
		s += "&"
	}

	return s
}

func buildQueryString(q Query, exclude string) string {
	var parts []string

	if q.Search != "" {
		parts = append(parts, "q="+url.QueryEscape(q.Search))
	}

	// Sort filter keys for deterministic output.
	filterKeys := make([]string, 0, len(q.Filters))
	for k := range q.Filters {
		if k != exclude {
			filterKeys = append(filterKeys, k)
		}
	}

	sort.Strings(filterKeys)

	for _, k := range filterKeys {
		escaped := make([]string, len(q.Filters[k]))
		for i, v := range q.Filters[k] {
			escaped[i] = url.QueryEscape(v)
		}

		parts = append(parts, url.QueryEscape(k)+"="+strings.Join(escaped, ","))
	}

	sortVal := q.Sort
	if q.Desc {
		sortVal = "-" + sortVal
	}

	if sortVal != "-updated" {
		parts = append(parts, "sort="+sortVal)
	}

	return strings.Join(parts, "&")
}

func filterByColumn(tasks []*task.Task, col string, values []string) []*task.Task {
	allowed := make(map[string]bool, len(values))
	for _, v := range values {
		allowed[v] = true
	}

	var result []*task.Task

	for _, t := range tasks {
		if allowed[t.Fields[col]] {
			result = append(result, t)
		}
	}

	return result
}
