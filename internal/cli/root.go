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

	// Version is set by ldflags during build.
	Version = "dev"
)

type appKey struct{}

var rootCmd = &cobra.Command{
	Use:           "craftops",
	Short:         "Modern Minecraft server operations and mod management",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: initApp,
	PersistentPostRun: func(cmd *cobra.Command, _ []string) {
		if a, ok := cmd.Context().Value(appKey{}).(*app); ok {
			a.Close()
		}
	},
}

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
	ctx := context.WithValue(cmd.Context(), appKey{}, application)
	cmd.SetContext(ctx)
	return nil
}

// Panics if called before initApp — programming error, not user error.
func appFrom(cmd *cobra.Command) *app {
	a, ok := cmd.Context().Value(appKey{}).(*app)
	if !ok || a == nil {
		panic("appFrom: app not found in context — was initApp skipped?")
	}
	return a
}
