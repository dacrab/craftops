package cli

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"craftops/internal/config"
	"craftops/internal/domain"
	"craftops/internal/ui"
)

// ── Server ──

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Minecraft server management",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Minecraft server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := appFrom()
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
		a := appFrom()
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
		ctx, a := cmd.Context(), appFrom()
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
		a := appFrom()
		status, err := a.Server.Status(cmd.Context())
		if err != nil {
			a.Terminal.Errorf("Failed to get status: %v", err)
			return err
		}
		if status.IsRunning {
			a.Terminal.Success("Server is running")
		} else {
			a.Terminal.Warning("Server is not running")
		}
		a.Terminal.Printf("  Session : %s\n", status.SessionName)
		a.Terminal.Printf("  Checked : %s\n", status.CheckedAt.Format(ui.TimeFormat))
		return nil
	},
}

// ── Mods ──

var modsCmd = &cobra.Command{
	Use:   "mods",
	Short: "Mod management",
}

var modsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all configured mods",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), appFrom()
		noBackup, _ := cmd.Flags().GetBool("no-backup")
		a.Terminal.Banner("Mod Update Manager")
		if !noBackup && a.Config.Backup.Enabled {
			a.Terminal.Info("Creating pre-update backup...")
			if path, err := a.Backup.Create(ctx); err != nil {
				return err
			} else if path != "" {
				a.Terminal.Successf("Backup created: %s", path)
			}
		}
		a.Terminal.Info("Updating mods...")
		forceUpdate, _ := cmd.Flags().GetBool("force")
		displayModResults(a, a.Mods.UpdateAll(ctx, forceUpdate))
		return nil
	},
}

var modsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed mods",
	RunE: func(_ *cobra.Command, _ []string) error {
		a := appFrom()
		mods, err := a.Mods.ListInstalled()
		if err != nil {
			a.Terminal.Errorf("Failed to list mods: %v", err)
			return err
		}
		rows := make([][]string, len(mods))
		for i, m := range mods {
			rows[i] = []string{m.Name, ui.FormatSize(m.Size), m.Modified.Format(ui.TimeFormat)}
		}
		renderList(a, "No mods installed in "+a.Config.Paths.Mods, "Installed Mods", []string{"Name", "Size", "Modified"}, rows)
		return nil
	},
}

func displayModResults(a *app, result *domain.ModUpdateResult) {
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
		for _, m := range slices.Sorted(maps.Keys(result.FailedMods)) {
			a.Terminal.Printf("   %s: %s\n", a.Terminal.ErrorSprint(m), a.Terminal.DimSprint(result.FailedMods[m]))
		}
		a.Terminal.Println()
	}
	printList(fmt.Sprintf("Skipped (%d):", len(result.SkippedMods)), result.SkippedMods, a.Terminal.WarningSprint)
}

// ── Backup ──

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup management",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := appFrom()
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
	RunE: func(_ *cobra.Command, _ []string) error {
		a := appFrom()
		backups, err := a.Backup.List()
		if err != nil {
			a.Terminal.Errorf("Failed to list backups: %v", err)
			return err
		}
		rows := make([][]string, len(backups))
		for i, b := range backups {
			rows[i] = []string{b.Name, b.CreatedAt.Format(ui.TimeFormat), ui.FormatSize(b.Size)}
		}
		renderList(a, "No backups found in "+a.Config.Paths.Backups, "Backups", []string{"Name", "Date", "Size"}, rows)
		return nil
	},
}

// renderList warns on empty listings, otherwise renders a titled table.
func renderList(a *app, emptyMsg, sectionTitle string, headers []string, rows [][]string) {
	if len(rows) == 0 {
		a.Terminal.Warning(emptyMsg)
		return
	}
	a.Terminal.Section(fmt.Sprintf("%s (%d)", sectionTitle, len(rows)))
	a.Terminal.Table(headers, rows)
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a backup by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		a := appFrom()
		if err := a.Backup.Delete(args[0]); err != nil {
			return err
		}
		a.Terminal.Successf("Deleted backup: %s", args[0])
		return nil
	},
}

// ── Health ──

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := cmd.Context(), appFrom()
		a.Terminal.Banner("System Health Check")

		var checks []domain.HealthCheck
		checks = append(checks, ui.CheckPath("Server directory", a.Config.Paths.Server))
		checks = append(checks, ui.CheckPath("Mods directory", a.Config.Paths.Mods))
		checks = append(checks, ui.CheckPath("Backups directory", a.Config.Paths.Backups))
		checks = append(checks, ui.CheckPath("Logs directory", a.Config.Paths.Logs))
		checks = append(checks, a.Server.HealthCheck(ctx)...)
		checks = append(checks, a.Mods.HealthCheck(ctx)...)
		checks = append(checks, a.Backup.HealthCheck(ctx)...)
		checks = append(checks, a.Notification.HealthCheck(ctx)...)

		a.Terminal.Section("Results")
		a.Terminal.HealthCheckTable(checks)
		return healthSummary(a, checks)
	},
}

func healthSummary(a *app, checks []domain.HealthCheck) error {
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

// ── Init ──

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file",
	// Skip normal app initialization — config may not exist yet.
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
	RunE: func(cmd *cobra.Command, _ []string) error {
		t := ui.NewTerminal()

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath == "" {
			outputPath = "config.toml"
		}
		force, _ := cmd.Flags().GetBool("force")

		if info, err := os.Stat(outputPath); err == nil && !force {
			if info.IsDir() {
				return errors.New("output path is a directory")
			}
			t.Warning("Config already exists: " + outputPath)
			t.Info("Use --force to overwrite")
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), domain.DirPerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		cfg := config.DefaultConfig()

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
