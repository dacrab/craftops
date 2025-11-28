package domain

import (
	"errors"
	"fmt"
	"time"
)

// HealthStatus represents the status of a health check
type HealthStatus string

const (
	StatusOK    HealthStatus = "OK"
	StatusWarn  HealthStatus = "WARN"
	StatusError HealthStatus = "ERROR"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Message string       `json:"message"`
}

// ServerStatus represents the current state of a Minecraft server
type ServerStatus struct {
	IsRunning   bool      `json:"is_running"`
	SessionName string    `json:"session_name,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// ModInfo represents information about a mod
type ModInfo struct {
	VersionID   string `json:"version_id"`
	Version     string `json:"version_number"`
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ProjectName string `json:"project_name"`
}

// ModUpdateResult represents the result of a mod update operation
type ModUpdateResult struct {
	UpdatedMods []string          `json:"updated_mods"`
	FailedMods  map[string]string `json:"failed_mods"`
	SkippedMods []string          `json:"skipped_mods"`
}

// InstalledMod represents an installed mod file
type InstalledMod struct {
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// BackupInfo represents information about a backup file
type BackupInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size_bytes"`
}

// SizeFormatted returns human-readable size
func (b BackupInfo) SizeFormatted() string {
	mb := float64(b.Size) / (1024 * 1024)
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", mb/1024)
	}
	return fmt.Sprintf("%.1f MB", mb)
}

// Sentinel errors
var (
	ErrServerNotRunning  = errors.New("server is not running")
	ErrServerRunning     = errors.New("server is already running")
	ErrServerJarNotFound = errors.New("server JAR file not found")
	ErrJavaNotFound      = errors.New("java runtime not found")
	ErrScreenNotFound    = errors.New("screen not found")
	ErrBackupsDisabled   = errors.New("backups are disabled")
	ErrNoModSources      = errors.New("no mod sources configured")
	ErrInvalidConfig     = errors.New("invalid configuration")
)

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [%s]: %s", e.Field, e.Message)
}

// ServiceError represents a service-level error with context
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

// NewServiceError creates a new service error
func NewServiceError(service, op string, err error) error {
	return &ServiceError{Service: service, Op: op, Err: err}
}

// APIError represents an API call error
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

// IsRetryable returns true if the error is retryable
func (e *APIError) IsRetryable() bool {
	return e.StatusCode >= 500 || e.StatusCode == 429
}
