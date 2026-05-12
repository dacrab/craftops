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
	// Context cancelled on SIGINT/SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := cli.Execute(ctx)
	cancel()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
