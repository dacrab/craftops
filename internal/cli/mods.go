package cli

import (
	"context"
	"fmt"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"craftops/internal/services"
)

var (
	forceUpdate bool
	noBackup    bool
)

var updateModsCmd = &cobra.Command{
	Use:   "update-mods",
	Short: "🔄 Update all configured mods to their latest versions",
	Long: `Update all configured mods to their latest compatible versions.

This command will:
• Check for updates from all configured mod sources
• Download newer versions if available
• Create a backup before updating (unless --no-backup is specified)
• Provide detailed progress and results`,
	RunE: runUpdateMods,
}

func runUpdateMods(cmd *cobra.Command, args []string) error {
	ctx := getContext()

	printBanner("Mod Update Manager")

	if cfg.DryRun {
		printWarning("🔍 Dry run mode - no changes will be made")
	}

	modService := services.NewModService(cfg, logger)

	if err := createBackupIfNeeded(ctx); err != nil {
		return err
	}

	result, err := updateMods(ctx, modService)
	if err != nil {
		return err
	}

	displayUpdateResults(result)
	return nil
}

func createBackupIfNeeded(ctx context.Context) error {
	if noBackup || !cfg.Backup.Enabled {
		return nil
	}

	printInfo("Creating backup before updating mods...")
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Creating backup..."),
		progressbar.OptionSpinnerType(14),
	)
	defer bar.Finish()

	backupService := services.NewBackupService(cfg, logger)
	backupPath, err := backupService.CreateBackup(ctx)
	if err != nil {
		printError(fmt.Sprintf("Failed to create backup: %v", err))
		return err
	}

	if backupPath != "" {
		printSuccess("Backup created successfully")
	}
	return nil
}

func updateMods(ctx context.Context, modService *services.ModService) (*services.ModUpdateResult, error) {
	printInfo("Updating mods...")
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Checking for updates..."),
		progressbar.OptionSpinnerType(14),
	)
	defer bar.Finish()

	result, err := modService.UpdateAllMods(ctx, forceUpdate)
	if err != nil {
		printError(fmt.Sprintf("Mod update failed: %v", err))
		return nil, err
	}

	return result, nil
}
func displayUpdateResults(result *services.ModUpdateResult) {
	printSection("Update Results")

	totalMods := len(result.UpdatedMods) + len(result.FailedMods) + len(result.SkippedMods)
	if totalMods == 0 {
		printInfo("🎯 No mods configured for updates")
		printInfo("💡 Add Modrinth mod URLs to your configuration file")
		return
	}

	displayUpdatedMods(result.UpdatedMods)
	displayFailedMods(result.FailedMods)
	displaySkippedMods(result.SkippedMods)
	displaySummary(result)
}

func displayUpdatedMods(updatedMods []string) {
	if len(updatedMods) == 0 {
		return
	}

	printSuccess(fmt.Sprintf("🎉 Successfully updated %d mods:", len(updatedMods)))
	for _, mod := range updatedMods {
		fmt.Printf("   ✅ %s\n", successColor.Sprint(mod))
	}
	fmt.Println()
}

func displayFailedMods(failedMods map[string]string) {
	if len(failedMods) == 0 {
		return
	}

	printError(fmt.Sprintf("❌ Failed to update %d mods:", len(failedMods)))
	for mod, err := range failedMods {
		fmt.Printf("   ❌ %s: %s\n", errorColor.Sprint(mod), dimColor.Sprint(err))
	}
	fmt.Println()
}

func displaySkippedMods(skippedMods []string) {
	if len(skippedMods) == 0 {
		return
	}

	printWarning(fmt.Sprintf("⏭️  Skipped %d mods:", len(skippedMods)))
	for _, mod := range skippedMods {
		fmt.Printf("   ⏭️  %s\n", warningColor.Sprint(mod))
	}
	fmt.Println()
}

func displaySummary(result *services.ModUpdateResult) {
	hasUpdates := len(result.UpdatedMods) > 0
	hasFailures := len(result.FailedMods) > 0
	hasSkipped := len(result.SkippedMods) > 0

	if !hasUpdates && !hasFailures && !hasSkipped {
		printSuccess("🎯 All mods are already up to date!")
		printInfo("💡 Your server is running the latest compatible versions")
		return
	}

	if hasUpdates || hasFailures {
		printSection("Summary")
		if hasUpdates {
			printSuccess(fmt.Sprintf("✅ %d successful updates", len(result.UpdatedMods)))
		}
		if hasFailures {
			printError(fmt.Sprintf("❌ %d failed updates", len(result.FailedMods)))
		}
		if hasSkipped {
			printWarning(fmt.Sprintf("⏭️  %d skipped updates", len(result.SkippedMods)))
		}
	}
}
func init() {
	rootCmd.AddCommand(updateModsCmd)
	updateModsCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update even if versions match")
	updateModsCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup creation before updating")
}
