package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/services"
	"craftops/internal/view"
)

// newUpdateModsStandaloneCmd provides the update-mods command
func newUpdateModsStandaloneCmd(factory ServiceFactory) *cobra.Command {
	var (
		forceUpdate bool
		noBackup    bool
	)

	cmd := &cobra.Command{
		Use:     "update-mods",
		Short:   "🔄 Update all configured mods to their latest versions",
		Aliases: []string{"update", "upgrade"},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUpdateMods(factory, noBackup, forceUpdate)
		},
	}
	cmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "force update even if versions match")
	cmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup creation before updating")
	return cmd
}

func runUpdateMods(factory ServiceFactory, noBackup, forceUpdate bool) error {
	handler := NewCommandHandler(factory)

	return handler.ExecuteWithContext(func(ctx context.Context) error {
		view.PrintBanner("Mod Update Manager")

		if factory.GetConfig().DryRun {
			view.PrintWarning("🔍 Dry run mode - no changes will be made")
		}

		if err := createBackupIfNeeded(ctx, factory, noBackup); err != nil {
			return err
		}

		result, err := updateMods(ctx, factory.GetModService(), forceUpdate)
		if err != nil {
			return err
		}

		displayUpdateResults(result)
		return nil
	})
}

func createBackupIfNeeded(ctx context.Context, factory ServiceFactory, noBackup bool) error {
	if noBackup || !factory.GetConfig().Backup.Enabled {
		return nil
	}

	view.PrintInfo("Creating backup before updating mods...")
	return runWithSpinner("Creating backup...", func() error {
		backupPath, err := factory.GetBackupService().CreateBackup(ctx)
		if err != nil {
			return handleError(err, "Failed to create backup")
		}
		if backupPath != "" {
			view.PrintSuccess("Backup created successfully")
		}
		return nil
	})
}

func updateMods(ctx context.Context, modService services.ModServiceInterface, forceUpdate bool) (*services.ModUpdateResult, error) {
	view.PrintInfo("Updating mods...")
	var result *services.ModUpdateResult
	err := runWithSpinner("Checking for updates...", func() error {
		var err error
		result, err = modService.UpdateAllMods(ctx, forceUpdate)
		if err != nil {
			return handleError(err, "Mod update failed")
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func displayUpdateResults(result *services.ModUpdateResult) {
	view.PrintSection("Update Results")

	totalMods := len(result.UpdatedMods) + len(result.FailedMods) + len(result.SkippedMods)
	if totalMods == 0 {
		view.PrintInfo("🎯 No mods configured for updates")
		view.PrintInfo("💡 Add Modrinth mod URLs to your configuration file")
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
	view.PrintSuccess(fmt.Sprintf("Successfully updated %d mod(s):", len(updatedMods)))
	for _, mod := range updatedMods {
		fmt.Printf("   ✅ %s\n", view.SuccessColor.Sprint(mod))
	}
	fmt.Println()
}

func displayFailedMods(failedMods map[string]string) {
	if len(failedMods) == 0 {
		return
	}
	view.PrintError(fmt.Sprintf("%d mods failed to update:", len(failedMods)))
	for mod, err := range failedMods {
		fmt.Printf("   ❌ %s: %s\n", view.ErrorColor.Sprint(mod), view.DimColor.Sprint(err))
	}
	fmt.Println()
}

func displaySkippedMods(skippedMods []string) {
	if len(skippedMods) == 0 {
		return
	}
	view.PrintWarning(fmt.Sprintf("%d mods skipped:", len(skippedMods)))
	for _, mod := range skippedMods {
		fmt.Printf("   ⏭️  %s\n", view.WarningColor.Sprint(mod))
	}
	fmt.Println()
}

func displaySummary(result *services.ModUpdateResult) {
	hasUpdates := len(result.UpdatedMods) > 0
	hasFailures := len(result.FailedMods) > 0
	hasSkipped := len(result.SkippedMods) > 0

	if !hasUpdates && !hasFailures && !hasSkipped {
		view.PrintSuccess("All mods are already up to date!")
		view.PrintInfo("💡 Your server is running the latest compatible versions")
		return
	}

	if hasUpdates || hasFailures {
		view.PrintSection("Summary")
		if hasUpdates {
			view.PrintSuccess(fmt.Sprintf("%d successful updates", len(result.UpdatedMods)))
		}
		if hasFailures {
			view.PrintError(fmt.Sprintf("%d failed updates", len(result.FailedMods)))
		}
		if hasSkipped {
			view.PrintWarning(fmt.Sprintf("%d skipped updates", len(result.SkippedMods)))
		}
	}
}
