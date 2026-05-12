package domain

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type HealthStatus string

const (
	StatusOK    HealthStatus = "OK"
	StatusWarn  HealthStatus = "WARN"
	StatusError HealthStatus = "ERROR"
)

type HealthCheck struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Message string       `json:"message"`
}

type ServerStatus struct {
	IsRunning   bool      `json:"is_running"`
	SessionName string    `json:"session_name,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

type ModInfo struct {
	VersionID   string `json:"version_id"`
	Version     string `json:"version_number"`
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ProjectName string `json:"project_name"`
}

type ModUpdateResult struct {
	UpdatedMods []string          `json:"updated_mods"`
	FailedMods  map[string]string `json:"failed_mods"`
	SkippedMods []string          `json:"skipped_mods"`
}

type InstalledMod struct {
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

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

var (
	ErrServerJarNotFound = errors.New("server JAR file not found")
	ErrBackupsDisabled   = errors.New("backups are disabled")
)

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

func (e *APIError) IsRetryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == 429
}
