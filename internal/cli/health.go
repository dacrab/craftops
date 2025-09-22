package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"craftops/internal/config"
	"craftops/internal/services"
	"craftops/internal/view"
)

func newHealthCheckCmd(factory ServiceFactory) *cobra.Command {
	handler := NewCommandHandler(factory)

	return createSimpleCommand("health-check", "🏥 Run comprehensive system health checks", []string{"health", "check"}, func() error {
		return handler.ExecuteWithContext(func(ctx context.Context) error {
			view.PrintBanner("System Health Check")

			// Collect all health checks with progress
			var allChecks []services.HealthCheck

			view.PrintStep(1, 5, "Checking system paths and permissions...")
			pathChecks := checkPaths(factory.GetConfig())
			allChecks = append(allChecks, pathChecks...)

			view.PrintStep(2, 5, "Validating mod management service...")
			modChecks := factory.GetModService().HealthCheck(ctx)
			allChecks = append(allChecks, modChecks...)

			view.PrintStep(3, 5, "Testing server management capabilities...")
			serverChecks := factory.GetServerService().HealthCheck(ctx)
			allChecks = append(allChecks, serverChecks...)

			view.PrintStep(4, 5, "Verifying backup system...")
			backupChecks := factory.GetBackupService().HealthCheck(ctx)
			allChecks = append(allChecks, backupChecks...)

			view.PrintStep(5, 5, "Checking notification configuration...")
			notificationChecks := factory.GetNotificationService().HealthCheck(ctx)
			allChecks = append(allChecks, notificationChecks...)

			// Display results and summary
			displayHealthResults(allChecks)
			return displayHealthSummary(allChecks)
		})
	})
}

// checkPaths performs path-related health checks
func checkPaths(cfg *config.Config) []services.HealthCheck {
	checks := []services.HealthCheck{}

	paths := map[string]string{
		"Server directory":  cfg.Paths.Server,
		"Mods directory":    cfg.Paths.Mods,
		"Backups directory": cfg.Paths.Backups,
		"Logs directory":    cfg.Paths.Logs,
	}

	for name, path := range paths {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				// Check write permissions
				testFile := filepath.Join(path, ".health_check_test")
				if file, err := os.Create(testFile); err == nil {
					file.Close()
					os.Remove(testFile)
					checks = append(checks, services.HealthCheck{
						Name:    name,
						Status:  "✅",
						Message: "OK",
					})
				} else {
					checks = append(checks, services.HealthCheck{
						Name:    name,
						Status:  "❌",
						Message: "No write permission",
					})
				}
			} else {
				checks = append(checks, services.HealthCheck{
					Name:    name,
					Status:  "❌",
					Message: "Path exists but is not a directory",
				})
			}
		} else {
			checks = append(checks, services.HealthCheck{
				Name:    name,
				Status:  "⚠️",
				Message: "Directory does not exist",
			})
		}
	}

	return checks
}

// displayHealthResults displays health check results in a beautiful table
func displayHealthResults(checks []services.HealthCheck) {
	view.PrintSection("Detailed Results")

	// Prepare table data
	headers := []string{"Component", "Status", "Details"}
	rows := make([][]string, len(checks))

	for i, check := range checks {
		rows[i] = []string{check.Name, check.Status, check.Message}
	}

	view.PrintTable(headers, rows)
}

// displayHealthSummary displays the health check summary and handles exit codes
func displayHealthSummary(checks []services.HealthCheck) error {
	passed, warnings, failed := countHealthChecks(checks)

	view.PrintSection("Health Check Summary")
	if failed > 0 {
		view.PrintError(fmt.Sprintf("%d checks failed", failed))
		if warnings > 0 {
			view.PrintWarning(fmt.Sprintf("%d warnings", warnings))
		}
		if passed > 0 {
			view.PrintSuccess(fmt.Sprintf("%d checks passed", passed))
		}
		fmt.Println()
		view.PrintError("🚨 System is not ready for production use!")
		view.PrintInfo("💡 Please fix the failed checks above and run again.")
		os.Exit(1)
	} else if warnings > 0 {
		view.PrintWarning(fmt.Sprintf("%d warnings found", warnings))
		view.PrintSuccess(fmt.Sprintf("%d checks passed", passed))
		fmt.Println()
		view.PrintWarning("System is functional but has warnings.")
		view.PrintInfo("💡 Consider addressing the warnings for optimal performance.")
	} else {
		view.PrintSuccess(fmt.Sprintf("All %d checks passed perfectly!", passed))
		fmt.Println()
		view.PrintSuccess("🚀 System is ready for production use!")
	}
	return nil
}

// countHealthChecks tallies the health check results
func countHealthChecks(checks []services.HealthCheck) (passed, warnings, failed int) {
	for _, check := range checks {
		switch check.Status {
		case "✅":
			passed++
		case "⚠️":
			warnings++
		case "❌":
			failed++
		}
	}
	return
}
