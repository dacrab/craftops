package main

import (
	"fmt"
	"os"

	"craftops/internal/cli"
)

const (
	Version = "2.0.0"
	Author  = "dacrab"
	License = "MIT"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
