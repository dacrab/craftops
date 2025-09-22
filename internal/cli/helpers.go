package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"craftops/internal/services"
	"craftops/internal/view"
)

// CommandHandler is a standardized command execution wrapper
type CommandHandler struct {
	factory ServiceFactory
}

// NewCommandHandler creates a new command handler with the service factory
func NewCommandHandler(factory ServiceFactory) *CommandHandler {
	return &CommandHandler{factory: factory}
}

// ExecuteWithContext runs a command with context and error handling
func (h *CommandHandler) ExecuteWithContext(operation func(context.Context) error) error {
	ctx, cancel := h.getContext()
	defer cancel()
	return operation(ctx)
}

// ExecuteWithSpinner runs an operation with a spinner and standardized error handling
func (h *CommandHandler) ExecuteWithSpinner(description string, operation func(context.Context) error) error {
	return h.ExecuteWithContext(func(ctx context.Context) error {
		return runWithSpinner(description, func() error {
			return operation(ctx)
		})
	})
}

// getContext returns a context with timeout
func (h *CommandHandler) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Minute)
}

// runWithSpinner runs an operation while showing a spinner
func runWithSpinner(description string, operation func() error) error {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSpinnerType(14),
	)
	defer func() { _ = bar.Finish() }()
	return operation()
}

// handleError provides consistent error formatting and logging
func handleError(err error, message string) error {
	if err != nil {
		view.PrintError(fmt.Sprintf("%s: %v", message, err))
		return err
	}
	return nil
}

// createSimpleCommand creates a command with standard error handling
func createSimpleCommand(use, short string, aliases []string, handler func() error) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Aliases: aliases,
		RunE: func(_ *cobra.Command, _ []string) error {
			return handler()
		},
	}
	return cmd
}

// createArgsCommand creates a command that requires arguments
func createArgsCommand(use, short string, args cobra.PositionalArgs, handler func([]string) error) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  args,
		RunE: func(_ *cobra.Command, args []string) error {
			return handler(args)
		},
	}
}

// printServerStatus prints formatted server status
func printServerStatus(status *services.ServerStatus) {
	view.PrintSection("🎮 Server Status")
	if status.IsRunning {
		view.PrintSuccess("Server is running")
		if status.PID != nil {
			fmt.Printf("PID: %d\n", *status.PID)
		}
		if status.Uptime != "" {
			fmt.Printf("Uptime: %s\n", status.Uptime)
		}
		if status.MemoryUsage != "" {
			fmt.Printf("Memory: %s\n", status.MemoryUsage)
		}
	} else {
		view.PrintError("Server is not running")
	}
}

// printBackupTable prints a formatted table of backups
func printBackupTable(backups []services.BackupInfo) {
	if len(backups) == 0 {
		view.PrintWarning("No backups found")
		return
	}

	view.PrintSection("💾 Available Backups")
	headers := []string{"Name", "Date", "Size"}
	rows := make([][]string, len(backups))
	for i, backup := range backups {
		rows[i] = []string{backup.Name, backup.CreatedAt, backup.Size}
	}
	view.PrintTable(headers, rows)
	view.PrintInfo(fmt.Sprintf("Total: %d backups", len(backups)))
}

// printModsTable prints a formatted table of installed mods
func printModsTable(mods []map[string]interface{}) {
	if len(mods) == 0 {
		view.PrintWarning("No mods found in mods directory")
		return
	}

	view.PrintSection("Installed Mods")
	headers := []string{"Name", "Filename", "Size", "Modified"}
	rows := make([][]string, 0, len(mods))
	for _, m := range mods {
		rows = append(rows, []string{
			fmt.Sprintf("%v", m["name"]),
			fmt.Sprintf("%v", m["filename"]),
			fmt.Sprintf("%v", m["size"]),
			fmt.Sprintf("%v", m["modified"]),
		})
	}
	view.PrintTable(headers, rows)
}
