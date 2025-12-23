package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"craftops/internal/config"
	"craftops/internal/domain"
)

var (
	forceUpdate bool
	noBackup    bool
	outputPath  string
	force       bool
)

// ============================================================================
// SERVER COMMANDS
// ============================================================================

// serverCmd groups all lifecycle-related commands
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Minecraft server management",
}

// serverStartCmd launches the server process
var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Minecraft server",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := App(cmd)
		a.Terminal.Info("Starting server...")
		if err := a.Server.Start(context.Background()); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to start server: %v", err))
			return err
		}
		a.Terminal.Success("Server is now running")
		return nil
	},
}

// serverStopCmd shuts down the server process
var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Minecraft server",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := App(cmd)
		a.Terminal.Info("Stopping server...")
		if err := a.Server.Stop(context.Background()); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to stop server: %v", err))
			return err
		}
		a.Terminal.Success("Server has been stopped")
		return nil
	},
}

// serverRestartCmd triggers a full restart with warnings if configured
var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Minecraft server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, a := context.Background(), App(cmd)
		if len(a.Config.Notifications.WarningIntervals) > 0 {
			a.Terminal.Info("Sending restart warnings...")
			if err := a.Notification.SendRestartWarnings(ctx); err != nil {
				a.Terminal.Warning(fmt.Sprintf("Warning notifications failed: %v", err))
			}
		}
		a.Terminal.Info("Restarting server...")
		if err := a.Server.Restart(ctx); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to restart: %v", err))
			_ = a.Notification.SendError(ctx, fmt.Sprintf("Server restart failed: %v", err))
			return err
		}
		a.Terminal.Success("Server restarted")
		_ = a.Notification.SendSuccess(ctx, "Server restarted successfully")
		return nil
	},
}

// serverStatusCmd reports whether the server process is alive
var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := App(cmd)
		status, err := a.Server.Status(context.Background())
		if err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to get status: %v", err))
			return err
		}
		if status.IsRunning {
			a.Terminal.Success("Server is running")
		} else {
			a.Terminal.Error("Server is not running")
		}
		return nil
	},
}

// MOD COMMANDS

// updateModsCmd checks for and installs mod updates
var updateModsCmd = &cobra.Command{
	Use:   "update-mods",
	Short: "Update all configured mods",
	RunE: func(cmd *cobra.Command, args []string) error {
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

// ============================================================================
// BACKUP COMMANDS
// ============================================================================

// backupCmd groups all backup-related commands
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup management",
}

// backupCreateCmd forces an immediate backup
var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := App(cmd)
		a.Terminal.Info("Creating backup...")
		path, err := a.Backup.Create(context.Background())
		if err != nil {
			if err == domain.ErrBackupsDisabled {
				a.Terminal.Warning("Backups disabled")
				return nil
			}
			return err
		}
		if path != "" {
			a.Terminal.Success("Backup created: " + path)
		}
		return nil
	},
}

// backupListCmd displays all available backup archives
var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := App(cmd)
		backups, err := a.Backup.List()
		if err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to list backups: %v", err))
			return err
		}
		if len(backups) == 0 {
			a.Terminal.Warning("No backups found")
			return nil
		}
		a.Terminal.Section("Available Backups")
		headers := []string{"Name", "Date", "Size"}
		rows := make([][]string, len(backups))
		for i, b := range backups {
			rows[i] = []string{b.Name, b.CreatedAt.Format("2006-01-02 15:04:05"), b.SizeFormatted()}
		}
		a.Terminal.Table(headers, rows)
		a.Terminal.Printf("\nTotal: %d backups\n", len(backups))
		return nil
	},
}

// ============================================================================
// HEALTH CHECK COMMAND
// ============================================================================

// healthCheckCmd runs an end-to-end diagnostic suite
var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		a := App(cmd)
		a.Terminal.Banner("System Health Check")
		var allChecks []domain.HealthCheck
		a.Terminal.Step(1, 5, "Checking paths and permissions...")
		allChecks = append(allChecks, checkPaths(a)...)
		a.Terminal.Step(2, 5, "Validating mod service...")
		allChecks = append(allChecks, a.Mods.HealthCheck(ctx)...)
		a.Terminal.Step(3, 5, "Testing server management...")
		allChecks = append(allChecks, a.Server.HealthCheck(ctx)...)
		a.Terminal.Step(4, 5, "Verifying backup system...")
		allChecks = append(allChecks, a.Backup.HealthCheck(ctx)...)
		a.Terminal.Step(5, 5, "Checking notifications...")
		allChecks = append(allChecks, a.Notification.HealthCheck(ctx)...)
		a.Terminal.Section("Detailed Results")
		a.Terminal.HealthCheckTable(allChecks)
		return displayHealthSummary(a, allChecks)
	},
}

func checkPaths(a *AppContainer) []domain.HealthCheck {
	return []domain.HealthCheck{
		domain.CheckPath("Server directory", a.Config.Paths.Server),
		domain.CheckPath("Mods directory", a.Config.Paths.Mods),
		domain.CheckPath("Backups directory", a.Config.Paths.Backups),
		domain.CheckPath("Logs directory", a.Config.Paths.Logs),
	}
}

// displayHealthSummary aggregates check results into a final terminal report
func displayHealthSummary(a *AppContainer, checks []domain.HealthCheck) error {
	var p, w, f int
	for _, c := range checks {
		switch c.Status {
		case domain.StatusOK:
			p++
		case domain.StatusWarn:
			w++
		case domain.StatusError:
			f++
		}
	}

	a.Terminal.Section("Summary")
	if f > 0 {
		a.Terminal.Error(fmt.Sprintf("%d failed, %d warnings, %d passed", f, w, p))
		return fmt.Errorf("%d health checks failed", f)
	}

	if w > 0 {
		a.Terminal.Warning(fmt.Sprintf("%d warnings, %d passed", w, p))
	} else {
		a.Terminal.Success(fmt.Sprintf("All %d checks passed!", p))
	}
	return nil
}

// ============================================================================
// INIT CONFIG COMMAND
// ============================================================================

// initCmd scaffolds a new configuration file with default settings
var initCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Initialize a new configuration file",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil // Skip app initialization - config may not exist yet
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if outputPath == "" {
			outputPath = "config.toml"
		}
		fmt.Printf("[1/4] Checking output path: %s\n", outputPath)
		if info, err := os.Stat(outputPath); err == nil && !force {
			if info.IsDir() {
				return fmt.Errorf("output path is a directory")
			}
			fmt.Printf("WARNING: Config already exists: %s\n", outputPath)
			fmt.Println("Use --force to overwrite")
			return nil
		}
		fmt.Println("[2/4] Creating directory structure...")
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		fmt.Println("[3/4] Generating default configuration...")
		cfg := config.DefaultConfig()
		fmt.Println("[4/4] Saving configuration file...")
		if err := cfg.SaveConfig(outputPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("\nSUCCESS: Configuration created: %s\n\n", outputPath)
		fmt.Println("Next steps:")
		fmt.Printf("  1. Edit the config: nano %s\n", outputPath)
		fmt.Println("  2. Add Modrinth mod URLs to [mods.modrinth_sources]")
		fmt.Println("  3. Run health check: craftops health-check")
		fmt.Println("  4. Start managing: craftops update-mods")
		return nil
	},
}

// ============================================================================
// INIT - REGISTER ALL COMMANDS
// ============================================================================

func init() {
	rootCmd.AddCommand(serverCmd, updateModsCmd, backupCmd, healthCheckCmd, initCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd)

	updateModsCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update")
	updateModsCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup")
	initCmd.Flags().StringVarP(&outputPath, "output", "o", "", "config path")
	initCmd.Flags().BoolVar(&force, "force", false, "overwrite")
}
