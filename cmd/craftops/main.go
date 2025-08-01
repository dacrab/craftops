package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"craftops/internal/cli"
)

const (
	Version = "2.0.0"
	Author  = "dacrab"
	License = "MIT"
)

func main() {
	// Detect if we're being called by an alias and adjust behavior if needed
	execName := filepath.Base(os.Args[0])

	// Support common aliases
	switch strings.ToLower(execName) {
	case "mmu", "minecraft-mod-updater":
		// These are just aliases, no special behavior needed
		// The CLI will work the same way
	}

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
