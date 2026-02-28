// Package main is the entry point for the taskboard CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/eunmann/taskboard/internal/config"
	"github.com/eunmann/taskboard/internal/index"
	"github.com/eunmann/taskboard/internal/server"
)

const defaultPort = 9746

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	dir := flag.String("dir", ".tasks", "path to tasks directory")
	port := flag.Int("port", defaultPort, "HTTP server port")
	initDir := flag.Bool("init", false, "initialize tasks directory and exit")
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("taskboard %s (%s)\n", version, commit)

		return
	}

	if *initDir {
		runInit(*dir)

		return
	}

	os.Exit(run(*dir, *port))
}

func runInit(dir string) {
	configPath := filepath.Join(dir, config.ConfigFile)

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Directory already initialized: %s exists.\n", configPath)

		return
	}

	if err := config.WriteDefault(dir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized task directory at %s/ with default config.\n", dir)
}

func run(dir string, port int) int {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: tasks directory not found at %s/\nRun 'taskboard --init' to create one, or specify a directory with --dir.\n", dir)

		return 1
	}

	configPath := filepath.Join(dir, config.ConfigFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: config.yaml not found in %s/\nRun 'taskboard --init' to generate a default configuration.\n", dir)

		return 1
	}

	idx, err := index.New(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		return 1
	}

	watcher, err := index.NewWatcher(idx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating watcher: %v\n", err)

		return 1
	}

	defer func() {
		if closeErr := watcher.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "error closing watcher: %v\n", closeErr)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go watcher.Start(ctx)

	addr := fmt.Sprintf(":%d", port)

	srv, err := server.New(idx, addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server: %v\n", err)

		return 1
	}

	srv.OnReady(func(a net.Addr) {
		fmt.Printf("Taskboard running at http://localhost%s\n", a.String())
	})

	if err := srv.ListenAndServe(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)

		return 1
	}

	return 0
}
