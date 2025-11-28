package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"craftops/internal/app"
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

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Minecraft server management",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Minecraft server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
		a.Terminal.Info("Starting server...")
		if err := a.Server.Start(ctx); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to start server: %v", err))
			return err
		}
		a.Terminal.Success("Server is now running")
		return nil
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Minecraft server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
		a.Terminal.Info("Stopping server...")
		if err := a.Server.Stop(ctx); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to stop server: %v", err))
			return err
		}
		a.Terminal.Success("Server has been stopped")
		return nil
	},
}

var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Minecraft server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
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

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
		status, err := a.Server.Status(ctx)
		if err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to get status: %v", err))
			return err
		}
		a.Terminal.Println()
		a.Terminal.Println("Server Status")
		a.Terminal.Println("─────────────────")
		if status.IsRunning {
			a.Terminal.Success("Server is running")
		} else {
			a.Terminal.Error("Server is not running")
		}
		return nil
	},
}

// ============================================================================
// MOD COMMANDS
// ============================================================================

var updateModsCmd = &cobra.Command{
	Use:   "update-mods",
	Short: "Update all configured mods",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
		a.Terminal.Banner("Mod Update Manager")
		if a.Config.DryRun {
			a.Terminal.Warning("Dry run mode - no changes will be made")
		}
		if !noBackup && a.Config.Backup.Enabled {
			a.Terminal.Info("Creating backup before updating...")
			if backupPath, err := a.Backup.Create(ctx); err != nil {
				if err != domain.ErrBackupsDisabled {
					a.Terminal.Error(fmt.Sprintf("Backup failed: %v", err))
					return err
				}
			} else if backupPath != "" {
				a.Terminal.Success("Backup created")
			}
		}
		a.Terminal.Info("Updating mods...")
		result, err := a.Mods.UpdateAll(ctx, forceUpdate)
		if err != nil {
			a.Terminal.Error(fmt.Sprintf("Mod update failed: %v", err))
			return err
		}
		displayModResults(a, result)
		return nil
	},
}

func displayModResults(a *app.App, result *domain.ModUpdateResult) {
	a.Terminal.Section("Update Results")
	total := len(result.UpdatedMods) + len(result.FailedMods) + len(result.SkippedMods)
	if total == 0 {
		a.Terminal.Info("No mods configured for updates")
		return
	}
	if len(result.UpdatedMods) > 0 {
		a.Terminal.Success(fmt.Sprintf("Updated %d mods:", len(result.UpdatedMods)))
		for _, mod := range result.UpdatedMods {
			a.Terminal.Printf("   %s\n", a.Terminal.SuccessSprint(mod))
		}
		a.Terminal.Println()
	}
	if len(result.FailedMods) > 0 {
		a.Terminal.Error(fmt.Sprintf("Failed %d mods:", len(result.FailedMods)))
		for mod, errMsg := range result.FailedMods {
			a.Terminal.Printf("   %s: %s\n", a.Terminal.ErrorSprint(mod), a.Terminal.DimSprint(errMsg))
		}
		a.Terminal.Println()
	}
	if len(result.SkippedMods) > 0 {
		a.Terminal.Warning(fmt.Sprintf("Skipped %d mods:", len(result.SkippedMods)))
		for _, mod := range result.SkippedMods {
			a.Terminal.Printf("   %s\n", a.Terminal.WarningSprint(mod))
		}
		a.Terminal.Println()
	}
	if len(result.UpdatedMods) == 0 && len(result.FailedMods) == 0 {
		a.Terminal.Success("All mods are up to date!")
	}
}

// ============================================================================
// BACKUP COMMANDS
// ============================================================================

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup management",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
		a.Terminal.Info("Creating backup...")
		backupPath, err := a.Backup.Create(ctx)
		if err != nil {
			if err == domain.ErrBackupsDisabled {
				a.Terminal.Warning("Backups are disabled in configuration")
				return nil
			}
			a.Terminal.Error(fmt.Sprintf("Backup failed: %v", err))
			return err
		}
		if backupPath != "" {
			a.Terminal.Success(fmt.Sprintf("Backup created: %s", backupPath))
		}
		return nil
	},
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := App()
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

var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()
		a := App()
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

func checkPaths(a *app.App) []domain.HealthCheck {
	paths := map[string]string{
		"Server directory":  a.Config.Paths.Server,
		"Mods directory":    a.Config.Paths.Mods,
		"Backups directory": a.Config.Paths.Backups,
		"Logs directory":    a.Config.Paths.Logs,
	}
	checks := make([]domain.HealthCheck, 0, len(paths))
	for name, path := range paths {
		checks = append(checks, checkPath(name, path))
	}
	return checks
}

func checkPath(name, path string) domain.HealthCheck {
	info, err := os.Stat(path)
	if err != nil {
		return domain.HealthCheck{Name: name, Status: domain.StatusWarn, Message: "Does not exist"}
	}
	if !info.IsDir() {
		return domain.HealthCheck{Name: name, Status: domain.StatusError, Message: "Not a directory"}
	}
	testFile := filepath.Join(path, ".health_test")
	f, err := os.Create(testFile)
	if err != nil {
		return domain.HealthCheck{Name: name, Status: domain.StatusError, Message: "No write permission"}
	}
	f.Close()
	os.Remove(testFile)
	return domain.HealthCheck{Name: name, Status: domain.StatusOK, Message: "OK"}
}

func displayHealthSummary(a *app.App, checks []domain.HealthCheck) error {
	passed, warnings, failed := 0, 0, 0
	for _, check := range checks {
		switch check.Status {
		case domain.StatusOK:
			passed++
		case domain.StatusWarn:
			warnings++
		case domain.StatusError:
			failed++
		}
	}
	a.Terminal.Section("Summary")
	if failed > 0 {
		a.Terminal.Error(fmt.Sprintf("%d checks failed", failed))
		if warnings > 0 {
			a.Terminal.Warning(fmt.Sprintf("%d warnings", warnings))
		}
		if passed > 0 {
			a.Terminal.Success(fmt.Sprintf("%d checks passed", passed))
		}
		a.Terminal.Println()
		a.Terminal.Error("System is not ready for production!")
		a.Terminal.Info("Please fix the failed checks and run again.")
		return fmt.Errorf("%d health checks failed", failed)
	}
	if warnings > 0 {
		a.Terminal.Warning(fmt.Sprintf("%d warnings", warnings))
		a.Terminal.Success(fmt.Sprintf("%d checks passed", passed))
		a.Terminal.Println()
		a.Terminal.Warning("System functional but has warnings.")
	} else {
		a.Terminal.Success(fmt.Sprintf("All %d checks passed!", passed))
		a.Terminal.Println()
		a.Terminal.Success("System ready for production!")
	}
	return nil
}

// ============================================================================
// INIT CONFIG COMMAND
// ============================================================================

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
	// Server commands
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)

	// Mod commands
	rootCmd.AddCommand(updateModsCmd)
	updateModsCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update even if versions match")
	updateModsCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip backup before updating")

	// Backup commands
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd)

	// Health check
	rootCmd.AddCommand(healthCheckCmd)

	// Init config
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path for config file")
	initCmd.Flags().BoolVar(&force, "force", false, "overwrite existing config")
}
