package cli

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"craftops/internal/services"
)

var (
	forceUpdate bool
	noBackup    bool
)

// updateModsCmd represents the update-mods command
var updateModsCmd = &cobra.Command{
	Use:   "update-mods",
	Short: "🔄 Update all configured mods to their latest versions",
	Long: `Update all configured mods to their latest compatible versions.

This command will:
• Check for updates from all configured mod sources
• Download newer versions if available
• Create a backup before updating (unless --no-backup is specified)
• Provide detailed progress and results`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()

		printBanner("Mod Update Manager")

		if cfg.DryRun {
			printWarning("🔍 Dry run mode - no changes will be made")
		}

		// Initialize services
		modService := services.NewModService(cfg, logger)
		backupService := services.NewBackupService(cfg, logger)
		notificationService := services.NewNotificationService(cfg, logger)

		// Create backup if requested and enabled
		if !noBackup && cfg.Backup.Enabled {
			printInfo("Creating backup before updating mods...")
			bar := progressbar.NewOptions(-1,
				progressbar.OptionSetDescription("Creating backup..."),
				progressbar.OptionSpinnerType(14),
			)

			backupPath, err := backupService.CreateBackup(ctx)
			bar.Finish()

			if err != nil {
				printError(fmt.Sprintf("Failed to create backup: %v", err))
				return err
			}

			if backupPath != "" {
				printSuccess("Backup created successfully")
			}
		}

		// Update mods
		printInfo("Updating mods...")
		bar := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("Checking for updates..."),
			progressbar.OptionSpinnerType(14),
		)

		result, err := modService.UpdateAllMods(ctx, forceUpdate)
		bar.Finish()

		if err != nil {
			printError(fmt.Sprintf("Mod update failed: %v", err))
			notificationService.SendErrorNotification(ctx, fmt.Sprintf("Mod update failed: %v", err))
			return err
		}

		// Display results
		displayUpdateResults(result)

		// Send notification
		if len(result.UpdatedMods) > 0 {
			message := fmt.Sprintf("Updated %d mods successfully", len(result.UpdatedMods))
			notificationService.SendSuccessNotification(ctx, message)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateModsCmd)

	updateModsCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update even if versions match")
	updateModsCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup creation before updating")
}

// displayUpdateResults displays the results of mod updates
func displayUpdateResults(result *services.ModUpdateResult) {
	printSection("Update Results")

	totalMods := len(result.UpdatedMods) + len(result.FailedMods) + len(result.SkippedMods)

	if totalMods == 0 {
		printInfo("🎯 No mods configured for updates")
		printInfo("💡 Add Modrinth mod URLs to your configuration file")
		return
	}

	if len(result.UpdatedMods) > 0 {
		printSuccess(fmt.Sprintf("🎉 Successfully updated %d mods:", len(result.UpdatedMods)))
		for _, mod := range result.UpdatedMods {
			fmt.Printf("   ✅ %s\n", successColor.Sprint(mod))
		}
		fmt.Println()
	}

	if len(result.FailedMods) > 0 {
		printError(fmt.Sprintf("❌ Failed to update %d mods:", len(result.FailedMods)))
		for mod, err := range result.FailedMods {
			fmt.Printf("   ❌ %s: %s\n", errorColor.Sprint(mod), dimColor.Sprint(err))
		}
		fmt.Println()
	}

	if len(result.SkippedMods) > 0 {
		printWarning(fmt.Sprintf("⏭️  Skipped %d mods:", len(result.SkippedMods)))
		for _, mod := range result.SkippedMods {
			fmt.Printf("   ⏭️  %s\n", warningColor.Sprint(mod))
		}
		fmt.Println()
	}

	if len(result.UpdatedMods) == 0 && len(result.FailedMods) == 0 && len(result.SkippedMods) == 0 {
		printSuccess("🎯 All mods are already up to date!")
		printInfo("💡 Your server is running the latest compatible versions")
	}

	// Summary
	if len(result.UpdatedMods) > 0 || len(result.FailedMods) > 0 {
		printSection("Summary")
		if len(result.UpdatedMods) > 0 {
			printSuccess(fmt.Sprintf("✅ %d successful updates", len(result.UpdatedMods)))
		}
		if len(result.FailedMods) > 0 {
			printError(fmt.Sprintf("❌ %d failed updates", len(result.FailedMods)))
		}
		if len(result.SkippedMods) > 0 {
			printWarning(fmt.Sprintf("⏭️  %d skipped updates", len(result.SkippedMods)))
		}
	}
}
