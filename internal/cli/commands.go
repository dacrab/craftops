// Package cli provides the command-line interface for craftops.
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"craftops/internal/config"
	"craftops/internal/domain"
	"craftops/internal/ui"
)

// Flag variables for commands.
var (
	forceUpdate bool
	noBackup    bool
	outputPath  string
	force       bool
)

func init() {
	rootCmd.AddCommand(serverCmd, updateModsCmd, backupCmd, healthCheckCmd, initCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd)

	updateModsCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update even if mod is current")
	updateModsCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip pre-update backup")
	initCmd.Flags().StringVarP(&outputPath, "output", "o", "", "config file output path")
	initCmd.Flags().BoolVar(&force, "force", false, "overwrite existing config file")
}

// ── Server ────────────────────────────────────────────────────────────────────

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Minecraft server management",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Minecraft server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
		a.Terminal.Info("Starting server...")
		if err := a.Server.Start(cmd.Context()); err != nil {
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
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
		a.Terminal.Info("Stopping server...")
		if err := a.Server.Stop(cmd.Context()); err != nil {
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
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), app(cmd)
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
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
		status, err := a.Server.Status(cmd.Context())
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

// ── Mods ─────────────────────────────────────────────────────────────────────

var updateModsCmd = &cobra.Command{
	Use:   "update-mods",
	Short: "Update all configured mods",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), app(cmd)
		a.Terminal.Banner("Mod Update Manager")
		if !noBackup && a.Config.Backup.Enabled {
			a.Terminal.Info("Creating backup...")
			if _, err := a.Backup.Create(ctx); err != nil && !errors.Is(err, domain.ErrBackupsDisabled) {
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

func displayModResults(a *appContainer, result *domain.ModUpdateResult) {
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
		// Sort keys for deterministic output — FailedMods is a map.
		keys := slices.Sorted(func(yield func(string) bool) {
			for k := range result.FailedMods {
				if !yield(k) {
					return
				}
			}
		})
		for _, m := range keys {
			a.Terminal.Printf("   %s: %s\n", a.Terminal.ErrorSprint(m), a.Terminal.DimSprint(result.FailedMods[m]))
		}
		a.Terminal.Println()
	}
	printList(fmt.Sprintf("Skipped %d mods:", len(result.SkippedMods)), result.SkippedMods, a.Terminal.WarningSprint)
}

// ── Backup ────────────────────────────────────────────────────────────────────

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup management",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
		a.Terminal.Info("Creating backup...")
		path, err := a.Backup.Create(cmd.Context())
		if err != nil {
			if errors.Is(err, domain.ErrBackupsDisabled) {
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

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
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

// ── Health ────────────────────────────────────────────────────────────────────

var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), app(cmd)
		a.Terminal.Banner("System Health Check")

		var checks []domain.HealthCheck
		a.Terminal.Step(1, 5, "Checking paths and permissions...")
		checks = append(checks, domain.CheckPath("Server directory", a.Config.Paths.Server))
		checks = append(checks, domain.CheckPath("Mods directory", a.Config.Paths.Mods))
		checks = append(checks, domain.CheckPath("Backups directory", a.Config.Paths.Backups))
		checks = append(checks, domain.CheckPath("Logs directory", a.Config.Paths.Logs))
		a.Terminal.Step(2, 5, "Validating mod service...")
		checks = append(checks, a.Mods.HealthCheck(ctx)...)
		a.Terminal.Step(3, 5, "Testing server management...")
		checks = append(checks, a.Server.HealthCheck(ctx)...)
		a.Terminal.Step(4, 5, "Verifying backup system...")
		checks = append(checks, a.Backup.HealthCheck(ctx)...)
		a.Terminal.Step(5, 5, "Checking notifications...")
		checks = append(checks, a.Notification.HealthCheck(ctx)...)

		a.Terminal.Section("Detailed Results")
		a.Terminal.HealthCheckTable(checks)
		return healthSummary(a, checks)
	},
}

func healthSummary(a *appContainer, checks []domain.HealthCheck) error {
	var passed, warned, failed int
	for _, c := range checks {
		switch c.Status {
		case domain.StatusOK:
			passed++
		case domain.StatusWarn:
			warned++
		case domain.StatusError:
			failed++
		}
	}
	a.Terminal.Section("Summary")
	if failed > 0 {
		a.Terminal.Error(fmt.Sprintf("%d failed, %d warnings, %d passed", failed, warned, passed))
		return fmt.Errorf("%d health checks failed", failed)
	}
	if warned > 0 {
		a.Terminal.Warning(fmt.Sprintf("%d warnings, %d passed", warned, passed))
	} else {
		a.Terminal.Success(fmt.Sprintf("All %d checks passed!", passed))
	}
	return nil
}

// ── Init ──────────────────────────────────────────────────────────────────────

var initCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Initialize a new configuration file",
	// Skip the normal app initialization — config may not exist yet.
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
	RunE: func(_ *cobra.Command, _ []string) error {
		t := ui.NewTerminal()

		if outputPath == "" {
			outputPath = "config.toml"
		}

		t.Step(1, 4, "Checking output path: "+outputPath)
		if info, err := os.Stat(outputPath); err == nil && !force {
			if info.IsDir() {
				return errors.New("output path is a directory")
			}
			t.Warning("Config already exists: " + outputPath)
			t.Info("Use --force to overwrite")
			return nil
		}

		t.Step(2, 4, "Creating directory structure...")
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		t.Step(3, 4, "Generating default configuration...")
		cfg := config.DefaultConfig()

		t.Step(4, 4, "Saving configuration file...")
		if err := cfg.SaveConfig(outputPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		t.Success("Configuration created: " + outputPath)
		t.Println()
		t.Info("Next steps:")
		t.Printf("  1. Edit the config: nano %s\n", outputPath)
		t.Println("  2. Add Modrinth mod URLs to [mods.modrinth_sources]")
		t.Println("  3. Run health check: craftops health-check")
		t.Println("  4. Start managing: craftops update-mods")
		return nil
	},
}
