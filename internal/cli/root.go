package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"craftops/internal/config"
	"craftops/internal/logger"
	"craftops/internal/view"
)

var (
	cfgFile string
	debug   bool
	dryRun  bool
	// Version can be set by ldflags during build
	Version = "2.0.1"
)

// NewRootCmd creates the root command and sets up its flags and subcommands.
func NewRootCmd(serviceFactory ServiceFactory) *cobra.Command {
	var actualFactory ServiceFactory = serviceFactory

	rootCmd := &cobra.Command{
		Use:   "craftops",
		Short: "🎮 Modern Minecraft server operations and mod management",
		Long: `CraftOps is a comprehensive CLI tool for Minecraft server operations and mod management.
	
Features:
• Server lifecycle management (start, stop, restart) - Linux/macOS only
• Automated backups with retention policies
• Discord notifications and player warnings
• Health checks and configuration validation`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// ensure view output goes to cobra's output writer
			var w io.Writer = cmd.OutOrStdout()
			view.SetWriter(w)
			// Initialize factory only if not provided (for real application, not tests)
			if actualFactory == nil {
				cfg, err := config.LoadConfig(cfgFile)
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}

				// Override config with CLI flags
				if debug {
					cfg.Debug = true
					cfg.Logging.Level = "DEBUG"
				}
				if dryRun {
					cfg.DryRun = true
				}

				appLogger := logger.New(cfg)
				actualFactory = NewServiceFactory(cfg, appLogger)
			}
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "show what would be done without making changes")

	// Add subcommands
	addBackupCommands(rootCmd, actualFactory)
	addHealthCheckCommand(rootCmd, actualFactory)
	addInitCommand(rootCmd)
	addModsCommands(rootCmd, actualFactory)
	addServerCommands(rootCmd, actualFactory)
	addRootShortcuts(rootCmd, actualFactory)

	// Add version flag and alias top-level help/version
	rootCmd.Flags().BoolP("version", "v", false, "show version information")
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if version, _ := cmd.Flags().GetBool("version"); version {
			fmt.Printf("CraftOps v%s\n", Version)
			return
		}
		_ = cmd.Help()
	}

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	cmd := NewRootCmd(nil)
	return cmd.Execute()
}

// Helper functions for adding subcommands
func addBackupCommands(rootCmd *cobra.Command, factory ServiceFactory) {
	rootCmd.AddCommand(newBackupCmd(factory))
}

func addHealthCheckCommand(rootCmd *cobra.Command, factory ServiceFactory) {
	rootCmd.AddCommand(newHealthCheckCmd(factory))
}

func addInitCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newInitCmd())
}

func addModsCommands(rootCmd *cobra.Command, factory ServiceFactory) {
	rootCmd.AddCommand(newModsCmd(factory))
	rootCmd.AddCommand(newUpdateModsStandaloneCmd(factory))
}

func addServerCommands(rootCmd *cobra.Command, factory ServiceFactory) {
	rootCmd.AddCommand(newServerCmd(factory))
}

// addRootShortcuts registers simplified top-level commands for better UX
func addRootShortcuts(rootCmd *cobra.Command, factory ServiceFactory) {
	handler := NewCommandHandler(factory)

	// Server shortcuts
	startCmd := createSimpleCommand("start", "Start the Minecraft server", []string{"up"}, func() error {
		view.PrintInfo("Starting server...")
		return handler.ExecuteWithSpinner("Starting server...", func(ctx context.Context) error {
			if err := factory.GetServerService().Start(ctx); err != nil {
				return handleError(err, "Failed to start server")
			}
			view.PrintSuccess("Server is now running")
			return nil
		})
	})

	stopCmd := createSimpleCommand("stop", "Stop the Minecraft server", nil, func() error {
		view.PrintInfo("Stopping server...")
		return handler.ExecuteWithSpinner("Stopping server...", func(ctx context.Context) error {
			if err := factory.GetServerService().Stop(ctx); err != nil {
				return handleError(err, "Failed to stop server")
			}
			view.PrintSuccess("Server has been stopped")
			return nil
		})
	})

	restartCmd := createSimpleCommand("restart", "Restart the Minecraft server", []string{"reload"}, func() error {
		return handler.ExecuteWithContext(func(ctx context.Context) error {
			srv := factory.GetServerService()
			notify := factory.GetNotificationService()

			// Send warnings if configured
			if len(factory.GetConfig().Notifications.WarningIntervals) > 0 {
				view.PrintInfo("Sending restart warnings to players...")
				if err := notify.SendRestartWarnings(ctx); err != nil {
					view.PrintWarning("Failed to send warnings: " + err.Error())
				}
			}

			view.PrintInfo("Restarting server...")
			return runWithSpinner("Restarting server...", func() error {
				if err := srv.Restart(ctx); err != nil {
					_ = notify.SendErrorNotification(ctx, "Server restart failed: "+err.Error())
					return handleError(err, "Failed to restart server")
				}
				view.PrintSuccess("Server has been restarted")
				_ = notify.SendSuccessNotification(ctx, "Server restarted successfully")
				return nil
			})
		})
	})

	statusCmd := createSimpleCommand("status", "Show server status", nil, func() error {
		return handler.ExecuteWithContext(func(ctx context.Context) error {
			status, err := factory.GetServerService().GetStatus(ctx)
			if err != nil {
				return handleError(err, "Failed to get server status")
			}
			printServerStatus(status)
			return nil
		})
	})

	rootCmd.AddCommand(startCmd, stopCmd, restartCmd, statusCmd)
}
