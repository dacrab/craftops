package cli

import (
	"context"

	"github.com/spf13/cobra"

	"craftops/internal/view"
)

func newBackupCmd(factory ServiceFactory) *cobra.Command {
	handler := NewCommandHandler(factory)

	backupCmd := &cobra.Command{
		Use:     "backup",
		Aliases: []string{"bk"},
		Short:   "💾 Backup management commands",
		Long:    `Commands for creating and managing server backups.`,
	}

	// Create backup command
	createCmd := createSimpleCommand("create", "💾 Create a backup of the server", []string{"now", "run"}, func() error {
		view.PrintInfo("Creating backup...")
		var backupPath string
		err := handler.ExecuteWithSpinner("Creating backup...", func(ctx context.Context) error {
			var e error
			backupPath, e = factory.GetBackupService().CreateBackup(ctx)
			return e
		})
		if err != nil {
			return handleError(err, "Backup failed")
		}

		if backupPath != "" {
			view.PrintSuccess("Backup created: " + backupPath)
		} else {
			view.PrintInfo("Backup creation skipped (disabled or dry run)")
		}
		return nil
	})

	// List backups command
	listCmd := createSimpleCommand("list", "📋 List available backups", []string{"ls"}, func() error {
		backups, err := factory.GetBackupService().ListBackups()
		if err != nil {
			return handleError(err, "Failed to list backups")
		}
		printBackupTable(backups)
		return nil
	})

	// Restore backup command
	var force bool
	restoreCmd := createArgsCommand("restore <path-to-backup>", "♻️ Restore a backup archive", cobra.ExactArgs(1), func(args []string) error {
		view.PrintInfo("Restoring backup...")
		return handler.ExecuteWithSpinner("Restoring...", func(ctx context.Context) error {
			err := factory.GetBackupService().RestoreBackup(ctx, args[0], force)
			if err != nil {
				return handleError(err, "Restore failed")
			}
			view.PrintSuccess("Backup restored successfully")
			return nil
		})
	})
	restoreCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files in server directory")

	backupCmd.AddCommand(createCmd, listCmd, restoreCmd)
	return backupCmd
}
