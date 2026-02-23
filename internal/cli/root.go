package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/config"
)

var (
	cfgFile string
	debug   bool
	dryRun  bool

	// Version is set by ldflags during build
	Version = "dev"
)

// appKey is the context key for the appContainer
type appKey struct{}

// rootCmd defines the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "craftops",
	Short: "Modern Minecraft server operations and mod management",
	Long: `CraftOps is a CLI tool for Minecraft server operations and mod management.

Features:
  - Server lifecycle management (start, stop, restart)
  - Automated mod updates from Modrinth
  - Backups with retention policies
  - Discord notifications
  - Health checks`,
	PersistentPreRunE: initApp,
	PersistentPostRun: func(cmd *cobra.Command, _ []string) {
		if a, ok := cmd.Context().Value(appKey{}).(*appContainer); ok {
			a.Close()
		}
	},
}

// Execute runs the root command with the provided context.
// The context should carry OS signal cancellation so long-running operations
// respect SIGINT/SIGTERM.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done")
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("CraftOps v{{.Version}}\n")
	rootCmd.Run = func(cmd *cobra.Command, _ []string) { _ = cmd.Help() }
}

// initApp handles configuration loading and dependency injection for all commands
func initApp(cmd *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if debug {
		cfg.Debug = true
		cfg.Logging.Level = "DEBUG"
	}
	if dryRun {
		cfg.DryRun = true
	}

	application := newApp(cfg)
	// Inject the application container into the command context to avoid global state "lock-in"
	ctx := context.WithValue(cmd.Context(), appKey{}, application)
	cmd.SetContext(ctx)
	return nil
}

// App extracts the appContainer from the command context
func app(cmd *cobra.Command) *appContainer {
	if a, ok := cmd.Context().Value(appKey{}).(*appContainer); ok {
		return a
	}
	return nil
}
