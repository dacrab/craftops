// Package domain defines shared data types used across service, CLI, and UI layers.
package domain

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// HealthStatus represents a diagnostic check outcome.
type HealthStatus string

// Health status values.
const (
	StatusOK    HealthStatus = "OK"
	StatusWarn  HealthStatus = "WARN"
	StatusError HealthStatus = "ERROR"
)

// HealthCheck is the result of a single diagnostic check.
type HealthCheck struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Message string       `json:"message"`
}

// ServerStatus describes whether the Minecraft server process is active.
type ServerStatus struct {
	IsRunning   bool      `json:"is_running"`
	SessionName string    `json:"session_name,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// ModInfo holds metadata for a mod version from Modrinth.
type ModInfo struct {
	VersionID   string `json:"version_id"`
	Version     string `json:"version_number"`
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ProjectName string `json:"project_name"`
}

// ModUpdateResult aggregates outcomes of a bulk mod update.
type ModUpdateResult struct {
	UpdatedMods []string          `json:"updated_mods"`
	FailedMods  map[string]string `json:"failed_mods"`
	SkippedMods []string          `json:"skipped_mods"`
}

// InstalledMod represents a .jar file in the mods directory.
type InstalledMod struct {
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// BackupInfo holds metadata for a backup archive.
type BackupInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size_bytes"`
}

// FormatSize returns a human-readable file size (e.g. "4.2 MB").
func FormatSize(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

// CheckPath verifies if a path exists and is a directory.
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

// Sentinel errors.
var (
	ErrServerJarNotFound = errors.New("server JAR file not found")
	ErrBackupsDisabled   = errors.New("backups are disabled")
)

// APIError captures details from a failed HTTP API call.
type APIError struct {
	URL        string
	StatusCode int
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("API error [%d]: %s (url: %s)", e.StatusCode, e.Message, e.URL)
	}
	return fmt.Sprintf("API error: %s (url: %s)", e.Message, e.URL)
}

// IsRetryable returns true for transient HTTP failures (5xx, 429).
func (e *APIError) IsRetryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == 429
}
