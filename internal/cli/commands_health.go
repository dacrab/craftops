package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/domain"
)

// healthCheckCmd runs an end-to-end diagnostic suite
var healthCheckCmd = &cobra.Command{
	Use:   "health-check",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, _ []string) error {
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

