// Package domain defines shared data types used across service, CLI, and UI layers.
package domain

import (
	"errors"
	"fmt"
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
	CheckedAt   time.Time `json:"checked_at"`
	SessionName string    `json:"session_name,omitempty"`
	IsRunning   bool      `json:"is_running"`
}

// ModInfo holds metadata for a mod version from Modrinth.
type ModInfo struct {
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
	Modified time.Time `json:"modified"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
}

// BackupInfo holds metadata for a backup archive.
type BackupInfo struct {
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size_bytes"`
}

// DefaultSessionName is the GNU screen session used to run the server when
// no session_name is configured.
const DefaultSessionName = "minecraft"

// DirPerm is the permission mode applied to directories created by craftops.
const DirPerm = 0o750

// Sentinel errors.
var (
	ErrServerJarNotFound = errors.New("server JAR file not found")
	ErrBackupsDisabled   = errors.New("backups are disabled")
)

// APIError captures details from a failed HTTP API call.
type APIError struct {
	URL        string
	Message    string
	StatusCode int
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
