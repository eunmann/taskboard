package task

import (
	"fmt"
	"slices"

	"github.com/eunmann/taskboard/internal/config"
)

// Warning represents a non-fatal validation issue with a task.
type Warning struct {
	Field   string
	Message string
}

func (w Warning) String() string {
	return fmt.Sprintf("%s: %s", w.Field, w.Message)
}

// Validate checks a task against the config and returns warnings.
func Validate(t *Task, columns map[string]config.Column) []Warning {
	reqWarnings := validateRequiredFields(t)
	idWarnings := validateIDMatch(t)
	fieldWarnings := validateFields(t, columns)
	refWarnings := validateRefs(t)

	total := len(reqWarnings) + len(idWarnings) + len(fieldWarnings) + len(refWarnings)
	if total == 0 {
		return nil
	}

	warnings := make([]Warning, 0, total)
	warnings = append(warnings, reqWarnings...)
	warnings = append(warnings, idWarnings...)
	warnings = append(warnings, fieldWarnings...)
	warnings = append(warnings, refWarnings...)

	return warnings
}

// ValidateDanglingRefs checks for refs that point to non-existent task IDs.
// This requires the full task index and is called separately from Validate.
func ValidateDanglingRefs(t *Task, allIDs map[string]bool) []Warning {
	var warnings []Warning

	for _, ref := range t.Refs {
		if !allIDs[ref.ID] {
			warnings = append(warnings, Warning{
				Field:   "refs",
				Message: fmt.Sprintf("reference to unknown task '%s'", ref.ID),
			})
		}
	}

	return warnings
}

func validateRequiredFields(t *Task) []Warning {
	var warnings []Warning

	if t.Title == "" {
		warnings = append(warnings, Warning{
			Field:   "title",
			Message: "missing required field 'title'",
		})
	}

	if t.Created.IsZero() {
		warnings = append(warnings, Warning{
			Field:   "created",
			Message: "missing required field 'created'",
		})
	}

	if t.Updated.IsZero() {
		warnings = append(warnings, Warning{
			Field:   "updated",
			Message: "missing required field 'updated'",
		})
	}

	return warnings
}

func validateIDMatch(t *Task) []Warning {
	if t.FileName == "" {
		return nil
	}

	filenameID := idFromFilename(t.FileName)
	if t.ID != filenameID && t.ID != "" {
		return []Warning{{
			Field:   "id",
			Message: fmt.Sprintf("filename '%s' does not match id field '%s'", t.FileName, t.ID),
		}}
	}

	return nil
}

func validateFields(t *Task, columns map[string]config.Column) []Warning {
	var warnings []Warning

	for field, value := range t.Fields {
		col, ok := columns[field]
		if !ok {
			warnings = append(warnings, Warning{
				Field:   field,
				Message: fmt.Sprintf("unknown field '%s' (not defined in config)", field),
			})

			continue
		}

		if !slices.Contains(col.ValueNames(), value) {
			warnings = append(warnings, Warning{
				Field:   field,
				Message: fmt.Sprintf("'%s' is not a valid value for column '%s'", value, field),
			})
		}
	}

	return warnings
}

func validateRefs(t *Task) []Warning {
	var warnings []Warning

	if t.SkippedRefs > 0 {
		warnings = append(warnings, Warning{
			Field:   "refs",
			Message: fmt.Sprintf("%d ref(s) skipped (missing type or id)", t.SkippedRefs),
		})
	}

	parentCount := 0

	for _, ref := range t.Refs {
		if ref.Type == RefParent {
			parentCount++
		}
	}

	if parentCount > 1 {
		warnings = append(warnings, Warning{
			Field:   "refs",
			Message: "task has multiple parent references (only one allowed)",
		})
	}

	return warnings
}
