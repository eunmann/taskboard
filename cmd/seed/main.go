// Package main implements the seed command that creates sample task files.
package main

import (
	"fmt"
	"os"

	"github.com/eunmann/taskboard/internal/seed"
)

const tasksDir = ".tasks"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	n, err := seed.Run(tasksDir)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	fmt.Printf("seed: wrote %d task files to %s/\n", n, tasksDir)

	return nil
}
