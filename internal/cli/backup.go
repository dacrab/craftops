package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/services"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "💾 Backup management commands",
	Long:  `Commands for creating and managing server backups.`,
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "💾 Create a backup of the server",
	Long: `Create a compressed backup of the server directory.

The backup will include all server files except those matching the configured
exclude patterns. Backups are automatically cleaned up based on the retention policy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		backupService := services.NewBackupService(cfg, logger)

		printInfo("Creating backup...")
		backupPath, err := backupService.CreateBackup(ctx)

		if err != nil {
			printError(fmt.Sprintf("Backup failed: %v", err))
			return err
		}

		if backupPath != "" {
			printSuccess(fmt.Sprintf("Backup created: %s", backupPath))
		} else {
			printInfo("Backup creation skipped (disabled or dry run)")
		}

		return nil
	},
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "📋 List available backups",
	Long:  `List all available backups with their creation dates and sizes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		backupService := services.NewBackupService(cfg, logger)

		backups, err := backupService.ListBackups()
		if err != nil {
			printError(fmt.Sprintf("Failed to list backups: %v", err))
			return err
		}

		if len(backups) == 0 {
			printWarning("No backups found")
			return nil
		}

		fmt.Println("\n💾 Available Backups")
		fmt.Println("────────────────────")
		fmt.Printf("%-40s %-20s %s\n", "Name", "Date", "Size")
		fmt.Printf("%-40s %-20s %s\n", "────", "────", "────")

		for _, backup := range backups {
			fmt.Printf("%-40s %-20s %s\n", backup.Name, backup.CreatedAt, backup.Size)
		}

		fmt.Printf("\nTotal: %d backups\n", len(backups))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupListCmd)
}
