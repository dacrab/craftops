package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/config"
)

var (
	cfgFile string
	debug   bool
	dryRun  bool

	// Version is set by ldflags during build
	Version = "2.1.0"

	// Global app instance (set during PersistentPreRunE)
	application *AppContainer
)

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
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if application != nil {
			application.Close()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done")
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("CraftOps v{{.Version}}\n")
	rootCmd.Run = func(cmd *cobra.Command, args []string) { _ = cmd.Help() }
}

// initApp handles configuration loading and dependency injection for all commands
func initApp(cmd *cobra.Command, args []string) error {
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

	application = NewApp(cfg)
	return nil
}

// App returns the global application instance, initialized via PersistentPreRunE
func App() *AppContainer {
	return application
}
