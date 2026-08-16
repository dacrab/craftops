package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/config"
)

// Version is set by ldflags during build.
var Version = "dev"

// activeApp holds the initialized app for the duration of a command run.
// It is set in PersistentPreRunE and cleared in PersistentPostRun.
var activeApp *app

var rootCmd = &cobra.Command{
	Use:               "craftops",
	Short:             "Modern Minecraft server operations and mod management",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: initApp,
	PersistentPostRun: func(cmd *cobra.Command, _ []string) {
		if activeApp != nil {
			activeApp.Close()
			activeApp = nil
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file path")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	rootCmd.PersistentFlags().Bool("dry-run", false, "show what would be done")
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("CraftOps v{{.Version}}\n")
	rootCmd.Run = func(cmd *cobra.Command, _ []string) { _ = cmd.Help() }

	rootCmd.AddCommand(serverCmd, modsCmd, backupCmd, healthCmd, initCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)
	modsCmd.AddCommand(modsUpdateCmd, modsListCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd, backupDeleteCmd)

	modsUpdateCmd.Flags().Bool("force", false, "force update even if mod is current")
	modsUpdateCmd.Flags().Bool("no-backup", false, "skip pre-update backup")
	initCmd.Flags().StringP("output", "o", "", "config file output path")
	initCmd.Flags().Bool("force", false, "overwrite existing config file")
}

// Execute runs the root command.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func initApp(cmd *cobra.Command, _ []string) error {
	cfgFile, _ := cmd.Flags().GetString("config")
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		cfg.Logging.Level = "DEBUG"
	}
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		cfg.DryRun = true
	}

	activeApp = newApp(cfg)
	return nil
}

// appFrom returns the initialized app, or nil if initApp was skipped.
func appFrom(_ *cobra.Command) *app {
	return activeApp
}