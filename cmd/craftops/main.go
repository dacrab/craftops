// Package main is the entry point for the craftops CLI application.
package main

import (
	"fmt"
	"os"

	"craftops/internal/cli"
)

func main() {
	// Execute the root command and exit with a non-zero status on error.
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
