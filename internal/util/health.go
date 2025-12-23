//nolint:revive // util is a common package name for shared utilities
package util

import (
	"context"
	"os/exec"
	"time"

	"craftops/internal/domain"
)

// CheckBinary verifies if a binary is available in PATH
func CheckBinary(ctx context.Context, binary, name string) domain.HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, binary, "-version").Run(); err == nil {
		return domain.HealthCheck{Name: name, Status: domain.StatusOK, Message: "Available"}
	}
	return domain.HealthCheck{Name: name, Status: domain.StatusError, Message: binary + " not found"}
}

