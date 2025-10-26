package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"craftops/internal/services"
)

// healthCheckCmd represents the health-check command
var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "Run comprehensive system health checks",
	Long: `Run comprehensive health checks to validate your configuration and system setup.

This command checks:
• Configuration file validity
• Directory permissions and accessibility
• Server JAR file existence
• Java runtime availability
• API connectivity (Modrinth)
• Discord webhook configuration
• Backup system functionality`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()

		printBanner("System Health Check")

		// Initialize services
		modService := services.NewModService(cfg, logger)
		serverService := services.NewServerService(cfg, logger)
		backupService := services.NewBackupService(cfg, logger)
		notificationService := services.NewNotificationService(cfg, logger)

		// Collect all health checks with progress
		var allChecks []services.HealthCheck

		printStep(1, 5, "Checking system paths and permissions...")
		pathChecks := checkPaths()
		allChecks = append(allChecks, pathChecks...)

		printStep(2, 5, "Validating mod management service...")
		modChecks := modService.HealthCheck(ctx)
		allChecks = append(allChecks, modChecks...)

		printStep(3, 5, "Testing server management capabilities...")
		serverChecks := serverService.HealthCheck(ctx)
		allChecks = append(allChecks, serverChecks...)

		printStep(4, 5, "Verifying backup system...")
		backupChecks := backupService.HealthCheck(ctx)
		allChecks = append(allChecks, backupChecks...)

		printStep(5, 5, "Checking notification configuration...")
		notificationChecks := notificationService.HealthCheck(ctx)
		allChecks = append(allChecks, notificationChecks...)

		// Display results
		displayHealthResults(allChecks)

		// Summary and exit code
		passed := 0
		warnings := 0
		failed := 0

		for _, check := range allChecks {
			switch check.Status {
			case "OK":
				passed++
			case "WARN":
				warnings++
			case "ERROR":
				failed++
			}
		}

		printSection("Health Check Summary")
		if failed > 0 {
			printError(fmt.Sprintf("%d checks failed", failed))
			if warnings > 0 {
				printWarning(fmt.Sprintf("%d warnings", warnings))
			}
			if passed > 0 {
				printSuccess(fmt.Sprintf("%d checks passed", passed))
			}
			fmt.Println()
			printError("System is not ready for production use!")
			printInfo("Please fix the failed checks above and run again.")
			return fmt.Errorf("health check failed: %d checks failed", failed)
		} else if warnings > 0 {
			printWarning(fmt.Sprintf("%d warnings found", warnings))
			printSuccess(fmt.Sprintf("%d checks passed", passed))
			fmt.Println()
			printWarning("System is functional but has warnings.")
			printInfo("Consider addressing the warnings for optimal performance.")
		} else {
			printSuccess(fmt.Sprintf("All %d checks passed.", passed))
			fmt.Println()
			printSuccess("System is ready for production use!")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCheckCmd)
}

// checkPaths performs path-related health checks
func checkPaths() []services.HealthCheck {
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
						Status:  "OK",
						Message: "OK",
					})
				} else {
					checks = append(checks, services.HealthCheck{
						Name:    name,
						Status:  "ERROR",
						Message: "No write permission",
					})
				}
			} else {
				checks = append(checks, services.HealthCheck{
					Name:    name,
					Status:  "ERROR",
					Message: "Path exists but is not a directory",
				})
			}
		} else {
			checks = append(checks, services.HealthCheck{
				Name:    name,
				Status:  "WARN",
				Message: "Directory does not exist",
			})
		}
	}

	return checks
}

// displayHealthResults displays health check results in a beautiful table
func displayHealthResults(checks []services.HealthCheck) {
	printSection("Detailed Results")

	// Prepare table data
	headers := []string{"Component", "Status", "Details"}
	rows := make([][]string, len(checks))

	for i, check := range checks {
		status := check.Status
		switch status {
		case "OK":
			status = successColor.Sprint(status)
		case "WARN":
			status = warningColor.Sprint(status)
		case "ERROR":
			status = errorColor.Sprint(status)
		}
		rows[i] = []string{check.Name, status, check.Message}
	}

	printTable(headers, rows)
}
