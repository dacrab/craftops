// Package main is the entry point for the craftops CLI application.
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
	// Wire a root context that is cancelled on SIGINT or SIGTERM.
	// This propagates to every cmd.Context() call inside command handlers,
	// so long-running operations (backups, downloads) honour Ctrl-C cleanly.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := cli.Execute(ctx)
	cancel() // release signal resources

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
