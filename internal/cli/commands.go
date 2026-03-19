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
	rootCmd.AddCommand(serverCmd, modsCmd, backupCmd, healthCmd, initCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)
	modsCmd.AddCommand(modsUpdateCmd, modsListCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd, backupDeleteCmd)

	modsUpdateCmd.Flags().BoolVar(&forceUpdate, "force", false, "force update even if mod is current")
	modsUpdateCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip pre-update backup")
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
			a.Terminal.Errorf("Failed to start server: %v", err)
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
			a.Terminal.Errorf("Failed to stop server: %v", err)
			return err
		}
		a.Terminal.Success("Server stopped")
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
				a.Terminal.Warningf("Warning notifications failed: %v", err)
			}
		}
		a.Terminal.Info("Restarting server...")
		if err := a.Server.Restart(ctx); err != nil {
			a.Terminal.Errorf("Failed to restart: %v", err)
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
	Short: "Show server status",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
		status, err := a.Server.Status(cmd.Context())
		if err != nil {
			a.Terminal.Errorf("Failed to get status: %v", err)
			return err
		}
		if status.IsRunning {
			a.Terminal.Success("Server is running")
			a.Terminal.Printf("  Session : %s\n", status.SessionName)
		} else {
			a.Terminal.Warning("Server is not running")
			a.Terminal.Printf("  Session : %s\n", status.SessionName)
		}
		a.Terminal.Printf("  Checked : %s\n", status.CheckedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

// ── Mods ─────────────────────────────────────────────────────────────────────

var modsCmd = &cobra.Command{
	Use:   "mods",
	Short: "Mod management",
}

var modsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all configured mods",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), app(cmd)
		a.Terminal.Banner("Mod Update Manager")
		if !noBackup && a.Config.Backup.Enabled {
			a.Terminal.Info("Creating pre-update backup...")
			if path, err := a.Backup.Create(ctx); err != nil && !errors.Is(err, domain.ErrBackupsDisabled) {
				return err
			} else if path != "" {
				a.Terminal.Successf("Backup created: %s", path)
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

var modsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed mods",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := app(cmd)
		mods, err := a.Mods.ListInstalled()
		if err != nil {
			a.Terminal.Errorf("Failed to list mods: %v", err)
			return err
		}
		if len(mods) == 0 {
			a.Terminal.Warning("No mods installed in " + a.Config.Paths.Mods)
			return nil
		}
		a.Terminal.Section(fmt.Sprintf("Installed Mods (%d)", len(mods)))
		headers := []string{"Name", "Size", "Modified"}
		rows := make([][]string, len(mods))
		for i, m := range mods {
			rows[i] = []string{m.Name, domain.FormatSize(m.Size), m.Modified.Format("2006-01-02 15:04:05")}
		}
		a.Terminal.Table(headers, rows)
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

	printList(fmt.Sprintf("Updated (%d):", len(result.UpdatedMods)), result.UpdatedMods, a.Terminal.SuccessSprint)
	if len(result.FailedMods) > 0 {
		a.Terminal.Errorf("Failed (%d):", len(result.FailedMods))
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
	printList(fmt.Sprintf("Skipped (%d):", len(result.SkippedMods)), result.SkippedMods, a.Terminal.WarningSprint)
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
				a.Terminal.Warning("Backups are disabled in config")
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
			a.Terminal.Errorf("Failed to list backups: %v", err)
			return err
		}
		if len(backups) == 0 {
			a.Terminal.Warning("No backups found in " + a.Config.Paths.Backups)
			return nil
		}
		a.Terminal.Section(fmt.Sprintf("Backups (%d)", len(backups)))
		headers := []string{"Name", "Date", "Size"}
		rows := make([][]string, len(backups))
		for i, b := range backups {
			rows[i] = []string{b.Name, b.CreatedAt.Format("2006-01-02 15:04:05"), b.SizeFormatted()}
		}
		a.Terminal.Table(headers, rows)
		return nil
	},
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a backup by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a := app(cmd)
		name := args[0]
		backups, err := a.Backup.List()
		if err != nil {
			return err
		}
		for _, b := range backups {
			if b.Name == name {
				if err := os.Remove(b.Path); err != nil {
					return fmt.Errorf("failed to delete backup: %w", err)
				}
				a.Terminal.Successf("Deleted backup: %s", name)
				return nil
			}
		}
		return fmt.Errorf("backup not found: %s", name)
	},
}

// ── Health ────────────────────────────────────────────────────────────────────

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), app(cmd)
		a.Terminal.Banner("System Health Check")

		var checks []domain.HealthCheck
		a.Terminal.Step(1, 4, "Checking paths...")
		checks = append(checks, domain.CheckPath("Server directory", a.Config.Paths.Server))
		checks = append(checks, domain.CheckPath("Mods directory", a.Config.Paths.Mods))
		checks = append(checks, domain.CheckPath("Backups directory", a.Config.Paths.Backups))
		checks = append(checks, domain.CheckPath("Logs directory", a.Config.Paths.Logs))
		a.Terminal.Step(2, 4, "Checking server...")
		checks = append(checks, a.Server.HealthCheck(ctx)...)
		checks = append(checks, a.Mods.HealthCheck(ctx)...)
		a.Terminal.Step(3, 4, "Checking backup & notifications...")
		checks = append(checks, a.Backup.HealthCheck(ctx)...)
		checks = append(checks, a.Notification.HealthCheck(ctx)...)
		a.Terminal.Step(4, 4, "Done")

		a.Terminal.Section("Results")
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
		a.Terminal.Errorf("%d failed, %d warnings, %d passed", failed, warned, passed)
		return fmt.Errorf("%d health checks failed", failed)
	}
	if warned > 0 {
		a.Terminal.Warningf("%d warnings, %d passed", warned, passed)
	} else {
		a.Terminal.Successf("All %d checks passed", passed)
	}
	return nil
}

// ── Init ──────────────────────────────────────────────────────────────────────

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file",
	// Skip the normal app initialization — config may not exist yet.
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
	RunE: func(_ *cobra.Command, _ []string) error {
		t := ui.NewTerminal()

		if outputPath == "" {
			outputPath = "config.toml"
		}

		t.Step(1, 3, "Checking output path: "+outputPath)
		if info, err := os.Stat(outputPath); err == nil && !force {
			if info.IsDir() {
				return errors.New("output path is a directory")
			}
			t.Warning("Config already exists: " + outputPath)
			t.Info("Use --force to overwrite")
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		t.Step(2, 3, "Generating default configuration...")
		cfg := config.DefaultConfig()

		t.Step(3, 3, "Saving...")
		if err := cfg.SaveConfig(outputPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		t.Success("Configuration created: " + outputPath)
		t.Println()
		t.Info("Next steps:")
		t.Printf("  1. Edit the config:   %s\n", outputPath)
		t.Println("  2. Add mod sources:   [mods.modrinth_sources] in config")
		t.Println("  3. Check setup:       craftops health")
		t.Println("  4. Update mods:       craftops mods update")
		t.Println("  5. Start server:      craftops server start")
		return nil
	},
}
