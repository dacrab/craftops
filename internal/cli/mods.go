package cli

import (
	"github.com/spf13/cobra"
)

func newModsCmd(factory ServiceFactory) *cobra.Command {
	modsCmd := &cobra.Command{
		Use:     "mods",
		Aliases: []string{"m"},
		Short:   "🧩 Mods management commands",
	}

	// List installed mods command
	listCmd := createSimpleCommand("list", "📋 List installed mods", nil, func() error {
		mods, err := factory.GetModService().ListInstalledMods()
		if err != nil {
			return handleError(err, "Failed to list mods")
		}
		printModsTable(mods)
		return nil
	})

	modsCmd.AddCommand(listCmd)
	return modsCmd
}
