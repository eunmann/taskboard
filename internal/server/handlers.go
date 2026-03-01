package server

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/index"
	"github.com/eunmann/taskboard/internal/task"
)

// StatusCount pairs a status value with its task count and color.
type StatusCount struct {
	Name  string
	Count int
	Color string
}

// ListData holds template data for the list page.
type ListData struct {
	Title         string
	Project       string
	Tasks         []*task.Task
	Columns       map[string]config.Column
	SortedColumns []ColumnInfo
	Query         Query
	ETag          string
	AllTags       []string
	TotalCount    int
	StatusCounts  []StatusCount
	WarningCount  int
}

// DetailData holds template data for the detail page.
type DetailData struct {
	Title         string
	Project       string
	Task          *task.Task
	Columns       map[string]config.Column
	SortedColumns []ColumnInfo
	AllTasks      map[string]*task.Task
}

func handleList(idx *index.Index, renderer *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := idx.Config()
		allTasks := idx.List()
		q := parseQuery(r, cfg.Columns)
		tasks := applySort(applyFilters(allTasks, q), q)
		etag := computeETag(idx.Version(), q)

		data := ListData{
			Title:         "Tasks",
			Project:       cfg.Project,
			Tasks:         tasks,
			Columns:       cfg.Columns,
			SortedColumns: SortedColumns(cfg.Columns),
			Query:         q,
			ETag:          etag,
			AllTags:       collectTags(allTasks),
			TotalCount:    len(allTasks),
			StatusCounts:  collectStatusCounts(allTasks, cfg.Columns),
			WarningCount:  countWarnings(allTasks),
		}

		if isHTMX(r) {
			w.Header().Set("ETag", etag)

			if err := renderer.RenderPartial(w, "list", "content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if err := renderer.Render(w, "list", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleDetail(idx *index.Index, renderer *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		t := idx.Get(id)
		if t == nil {
			http.NotFound(w, r)

			return
		}

		cfg := idx.Config()
		allTasks := make(map[string]*task.Task)

		for _, tk := range idx.List() {
			allTasks[tk.ID] = tk
		}

		data := DetailData{
			Title:         t.Title,
			Project:       cfg.Project,
			Task:          t,
			Columns:       cfg.Columns,
			SortedColumns: SortedColumns(cfg.Columns),
			AllTasks:      allTasks,
		}

		if isHTMX(r) {
			if err := renderer.RenderPartial(w, "detail", "content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if err := renderer.Render(w, "detail", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleTablePartial(idx *index.Index, renderer *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := idx.Config()
		q := parseQuery(r, cfg.Columns)
		currentETag := computeETag(idx.Version(), q)

		if r.Header.Get("If-None-Match") == currentETag {
			w.WriteHeader(http.StatusNotModified)

			return
		}

		tasks := applySort(applyFilters(idx.List(), q), q)

		data := ListData{
			Tasks:         tasks,
			Columns:       cfg.Columns,
			SortedColumns: SortedColumns(cfg.Columns),
			Query:         q,
			ETag:          currentETag,
		}

		w.Header().Set("ETag", currentETag)

		if r.URL.Query().Get("_poll") != "1" {
			qs := queryParams(q)
			if qs != "" {
				w.Header().Set("HX-Push-Url", "/?"+qs)
			} else {
				w.Header().Set("HX-Push-Url", "/")
			}
		}

		if err := renderer.RenderPartial(w, "list", "table", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func collectStatusCounts(tasks []*task.Task, columns map[string]config.Column) []StatusCount {
	col, ok := columns["status"]
	if !ok {
		return nil
	}

	counts := make(map[string]int)
	for _, t := range tasks {
		counts[t.Fields["status"]]++
	}

	result := make([]StatusCount, 0, len(col.Values))

	for _, v := range col.Values {
		c := counts[v.Name]
		if c > 0 {
			result = append(result, StatusCount{Name: v.Name, Count: c, Color: v.Color})
		}
	}

	return result
}

func countWarnings(tasks []*task.Task) int {
	n := 0

	for _, t := range tasks {
		if len(t.Warnings) > 0 {
			n++
		}
	}

	return n
}

func collectTags(tasks []*task.Task) []string {
	seen := make(map[string]bool)

	for _, t := range tasks {
		for _, tag := range t.Tags {
			seen[tag] = true
		}
	}

	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}

	sort.Strings(tags)

	return tags
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func computeETag(version uint64, q Query) string {
	qs := queryParams(q)
	if qs == "" {
		return fmt.Sprintf(`"%d"`, version)
	}

	return fmt.Sprintf(`"%d:%s"`, version, qs)
}
