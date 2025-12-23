package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/domain"
)

var (
	forceUpdate bool
	noBackup    bool
)

// updateModsCmd checks for and installs mod updates
var updateModsCmd = &cobra.Command{
	Use:   "update-mods",
	Short: "Update all configured mods",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := context.Background(), App(cmd)
		a.Terminal.Banner("Mod Update Manager")
		if !noBackup && a.Config.Backup.Enabled {
			a.Terminal.Info("Creating backup...")
			if _, err := a.Backup.Create(ctx); err != nil && err != domain.ErrBackupsDisabled {
				return err
			}
		}
		a.Terminal.Info("Updating mods...")
		result, err := a.Mods.UpdateAll(ctx, forceUpdate)
		if err != nil {
			return err
		}
		displayModResults(a, result)
		return nil
	},
}

// displayModResults prints a formatted summary of the mod update operation
func displayModResults(a *AppContainer, result *domain.ModUpdateResult) {
	a.Terminal.Section("Update Results")
	if len(result.UpdatedMods) == 0 && len(result.FailedMods) == 0 && len(result.SkippedMods) == 0 {
		a.Terminal.Info("No mods configured for updates")
		return
	}

	printList := func(title string, mods []string, sprint func(string) string) {
		if len(mods) == 0 {
			return
		}
		a.Terminal.Println(title)
		for _, m := range mods {
			a.Terminal.Printf("   %s\n", sprint(m))
		}
		a.Terminal.Println()
	}

	printList(fmt.Sprintf("Updated %d mods:", len(result.UpdatedMods)), result.UpdatedMods, a.Terminal.SuccessSprint)

	if len(result.FailedMods) > 0 {
		a.Terminal.Error(fmt.Sprintf("Failed %d mods:", len(result.FailedMods)))
		for m, e := range result.FailedMods {
			a.Terminal.Printf("   %s: %s\n", a.Terminal.ErrorSprint(m), a.Terminal.DimSprint(e))
		}
		a.Terminal.Println()
	}

	printList(fmt.Sprintf("Skipped %d mods:", len(result.SkippedMods)), result.SkippedMods, a.Terminal.WarningSprint)
}

