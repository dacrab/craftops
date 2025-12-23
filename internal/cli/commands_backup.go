package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/domain"
)

// backupCmd groups all backup-related commands
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup management",
}

// backupCreateCmd forces an immediate backup
var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := App(cmd)
		a.Terminal.Info("Creating backup...")
		path, err := a.Backup.Create(context.Background())
		if err != nil {
			if err == domain.ErrBackupsDisabled {
				a.Terminal.Warning("Backups disabled")
				return nil
			}
			return err
		}
		if path != "" {
			a.Terminal.Success("Backup created: " + path)
		}
		return nil
	},
}

// backupListCmd displays all available backup archives
var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := App(cmd)
		backups, err := a.Backup.List()
		if err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to list backups: %v", err))
			return err
		}
		if len(backups) == 0 {
			a.Terminal.Warning("No backups found")
			return nil
		}
		a.Terminal.Section("Available Backups")
		headers := []string{"Name", "Date", "Size"}
		rows := make([][]string, len(backups))
		for i, b := range backups {
			rows[i] = []string{b.Name, b.CreatedAt.Format("2006-01-02 15:04:05"), b.SizeFormatted()}
		}
		a.Terminal.Table(headers, rows)
		a.Terminal.Printf("\nTotal: %d backups\n", len(backups))
		return nil
	},
}
