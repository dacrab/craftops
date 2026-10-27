// Package domain defines shared data types used across service, CLI, and UI layers.
package domain

import (
	"errors"
	"fmt"
	"time"
)

// HealthStatus represents a diagnostic check outcome.
type HealthStatus string

// Health check outcomes.
const (
	StatusOK    HealthStatus = "OK"
	StatusWarn  HealthStatus = "WARN"
	StatusError HealthStatus = "ERROR"
)

// HealthCheck is the result of a single diagnostic check.
type HealthCheck struct {
	Name    string
	Status  HealthStatus
	Message string
}

// ServerStatus describes whether the Minecraft server process is active.
type ServerStatus struct {
	CheckedAt   time.Time
	SessionName string
	IsRunning   bool
}

// ModInfo holds metadata for a mod version from Modrinth.
type ModInfo struct {
	DownloadURL string
	Filename    string
	ProjectName string
}

// ModUpdateResult aggregates outcomes of a bulk mod update.
type ModUpdateResult struct {
	UpdatedMods []string
	FailedMods  map[string]string
	SkippedMods []string
}

// InstalledMod represents a .jar file in the mods directory.
type InstalledMod struct {
	Modified time.Time
	Name     string
	Size     int64
}

// BackupInfo holds metadata for a backup archive.
type BackupInfo struct {
	CreatedAt time.Time
	Name      string
	Path      string
	Size      int64
}

// DefaultSessionName is the GNU screen session used to run the server when
// no session_name is configured.
const DefaultSessionName = "minecraft"

// DirPerm is the permission mode applied to directories created by craftops.
const DirPerm = 0o750

// Sentinel errors returned by the service layer.
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
