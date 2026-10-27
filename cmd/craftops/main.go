// Package main is the craftops CLI entry point.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"craftops/internal/cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return cli.Execute(ctx)
}
