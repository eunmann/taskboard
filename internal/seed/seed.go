// Package seed provides sample task files for demonstration purposes.
package seed

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const fileMode = 0o644

//go:embed samples/*.yaml
var samplesFS embed.FS

// Run copies embedded sample YAML files into dir, skipping files that already
// exist. It returns the number of files written.
func Run(dir string) (int, error) {
	entries, err := fs.ReadDir(samplesFS, "samples")
	if err != nil {
		return 0, fmt.Errorf("read embedded samples: %w", err)
	}

	written := 0

	for _, e := range entries {
		dest := filepath.Join(dir, e.Name())
		if _, err := os.Stat(dest); err == nil {
			continue
		}

		data, err := samplesFS.ReadFile("samples/" + e.Name())
		if err != nil {
			return written, fmt.Errorf("read embedded %s: %w", e.Name(), err)
		}

		if err := os.WriteFile(dest, data, fileMode); err != nil {
			return written, fmt.Errorf("write %s: %w", dest, err)
		}

		written++
	}

	return written, nil
}
