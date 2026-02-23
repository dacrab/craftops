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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := cli.Execute(ctx)
	stop() // release signal resources before any exit
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
