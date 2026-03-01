package server

import (
	"fmt"
	"net/http"
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

// ForestData holds template data for the forest graph page.
type ForestData struct {
	Title        string
	Project      string
	TaskCount    int
	MermaidDef   string
	StatusLegend []LegendEntry
}

func handleForest(idx *index.Index, renderer *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := idx.Config()
		allTasks := idx.List()

		data := ForestData{
			Title:        "Forest",
			Project:      cfg.Project,
			TaskCount:    len(allTasks),
			MermaidDef:   buildMermaidDef(allTasks, cfg.Columns),
			StatusLegend: buildStatusLegend(cfg.Columns),
		}

		if isHTMX(r) {
			if err := renderer.RenderPartial(w, "forest", "content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if err := renderer.Render(w, "forest", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func buildMermaidDef(tasks []*task.Task, columns map[string]config.Column) string {
	var b strings.Builder

	b.WriteString("graph TD\n")

	for _, t := range tasks {
		label := nodeLabel(t, columns)
		fmt.Fprintf(&b, "    %s([\"%s\"])\n", t.ID, label)
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
		fmt.Fprintf(&b, "    click %s \"/task/%s\"\n", t.ID, t.ID)
	}

	return b.String()
}

func nodeLabel(t *task.Task, _ map[string]config.Column) string {
	title := sanitizeMermaid(t.Title)
	if title == "" {
		title = t.ID
	}

	if v := t.Fields["type"]; v != "" {
		return title + "<br/>" + sanitizeMermaid(v)
	}

	return title
}

func writeEdge(b *strings.Builder, sourceID string, ref task.Ref) {
	switch ref.Type {
	case task.RefParent:
		fmt.Fprintf(b, "    %s --> %s\n", sourceID, ref.ID)
	case task.RefBlockedBy:
		fmt.Fprintf(b, "    %s ==> %s\n", sourceID, ref.ID)
	case task.RefRelatesTo:
		fmt.Fprintf(b, "    %s -.-> %s\n", sourceID, ref.ID)
	}
}

func writeStyles(b *strings.Builder, tasks []*task.Task, columns map[string]config.Column) {
	for _, t := range tasks {
		status := t.Fields["status"]
		if status == "" {
			continue
		}

		bg := columnColor(columns, "status", status)
		if bg == "" {
			continue
		}

		text := contrastText(bg)
		fmt.Fprintf(b, "    style %s fill:%s,color:%s,stroke:%s,stroke-width:2px\n",
			t.ID, bg, text, bg)
	}
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
