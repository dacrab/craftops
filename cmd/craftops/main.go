package main

import (
	"fmt"
	"os"

	"craftops/internal/cli"
)

var (
	Version = "2.0.1" // Can be overridden by ldflags
	Author  = "dacrab"
	License = "MIT"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
