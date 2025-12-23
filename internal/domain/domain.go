// Package domain contains shared data structures and core validation logic.
package domain

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// HealthStatus represents the possible states of a system component.
type HealthStatus string

const (
	// StatusOK indicates a component is healthy
	StatusOK HealthStatus = "OK"
	// StatusWarn indicates a component has warnings
	StatusWarn HealthStatus = "WARN"
	// StatusError indicates a component has errors
	StatusError HealthStatus = "ERROR"
)

// HealthCheck captures the result of a diagnostic operation.
type HealthCheck struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Message string       `json:"message"`
}

// ServerStatus describes the current operational state of the Minecraft server.
type ServerStatus struct {
	IsRunning   bool      `json:"is_running"`
	SessionName string    `json:"session_name,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// ModInfo contains metadata for a specific mod version.
type ModInfo struct {
	VersionID   string `json:"version_id"`
	Version     string `json:"version_number"`
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ProjectName string `json:"project_name"`
}

// ModUpdateResult aggregates the outcomes of a bulk mod update.
type ModUpdateResult struct {
	UpdatedMods []string          `json:"updated_mods"`
	FailedMods  map[string]string `json:"failed_mods"`
	SkippedMods []string          `json:"skipped_mods"`
}

// InstalledMod represents a jar file found in the mods directory.
type InstalledMod struct {
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// BackupInfo contains metadata for a specific backup archive.
type BackupInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size_bytes"`
}

// SizeFormatted returns a human-readable representation of the file size.
func (b BackupInfo) SizeFormatted() string {
	mb := float64(b.Size) / (1024 * 1024)
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", mb/1024)
	}
	return fmt.Sprintf("%.1f MB", mb)
}

// CheckPath verifies if a path exists and is a directory, returning a HealthCheck.
func CheckPath(name, path string) HealthCheck {
	info, err := os.Stat(path)
	if err != nil {
		return HealthCheck{Name: name, Status: StatusWarn, Message: "Does not exist"}
	}
	if !info.IsDir() {
		return HealthCheck{Name: name, Status: StatusError, Message: "Not a directory"}
	}
	return HealthCheck{Name: name, Status: StatusOK, Message: "OK"}
}

// Sentinel errors for standard failure modes.
var (
	ErrServerJarNotFound = errors.New("server JAR file not found")
	ErrBackupsDisabled   = errors.New("backups are disabled")
)

// ServiceError wraps an error with service-specific context.
type ServiceError struct {
	Service string
	Op      string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s.%s: %v", e.Service, e.Op, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Service, e.Err)
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError creates a new contextual service error.
func NewServiceError(service, op string, err error) error {
	return &ServiceError{Service: service, Op: op, Err: err}
}

// APIError captures details from failed external service calls.
type APIError struct {
	URL        string
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("API error [%d]: %s (url: %s)", e.StatusCode, e.Message, e.URL)
	}
	return fmt.Sprintf("API error: %s (url: %s)", e.Message, e.URL)
}

// IsRetryable returns true if the error code suggests a transient failure.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == 429
}
