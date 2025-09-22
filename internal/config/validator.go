package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s (%v): %s", e.Field, e.Value, e.Message)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator validates configuration values.
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new configuration validator.
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// ValidateConfig performs comprehensive validation of the configuration.
func (v *Validator) ValidateConfig(cfg *Config) error {
	validationSteps := []func(*Config){
		v.validateMinecraft,
		v.validatePaths,
		v.validateServer,
		v.validateMods,
		v.validateBackup,
		v.validateNotifications,
		v.validateLogging,
	}

	for _, validate := range validationSteps {
		validate(cfg)
	}

	if len(v.errors) > 0 {
		return v.errors
	}
	return nil
}

func (v *Validator) validateMinecraft(cfg *Config) {
	if cfg.Minecraft.Version == "" {
		v.addError("minecraft.version", cfg.Minecraft.Version, "version cannot be empty")
	}

	validModloaders := []string{"fabric", "forge", "quilt", "neoforge"}
	if !v.containsString(validModloaders, strings.ToLower(cfg.Minecraft.Modloader)) {
		v.addError("minecraft.modloader", cfg.Minecraft.Modloader,
			fmt.Sprintf("must be one of: %s", strings.Join(validModloaders, ", ")))
	}
}

func (v *Validator) validatePaths(cfg *Config) {
	paths := map[string]string{
		"paths.server":  cfg.Paths.Server,
		"paths.mods":    cfg.Paths.Mods,
		"paths.backups": cfg.Paths.Backups,
		"paths.logs":    cfg.Paths.Logs,
	}

	for field, path := range paths {
		if path == "" {
			v.addError(field, path, "path cannot be empty")
			continue
		}

		// Paths are normalized to absolute earlier; no need to require absolute here strictly.

		// Check if parent directory exists for logs and backups
		if field == "paths.logs" || field == "paths.backups" {
			parent := filepath.Dir(path)
			if _, err := os.Stat(parent); os.IsNotExist(err) {
				if err := os.MkdirAll(parent, 0755); err != nil {
					v.addError(field, path, fmt.Sprintf("cannot create parent directory: %v", err))
				}
			}
		}
	}
}

func (v *Validator) validateServer(cfg *Config) {
	if cfg.Server.JarName == "" {
		v.addError("server.jar_name", cfg.Server.JarName, "jar name cannot be empty")
	}

	if cfg.Server.MaxStopWait <= 0 {
		v.addError("server.max_stop_wait", cfg.Server.MaxStopWait, "must be positive")
	}

	if cfg.Server.StartupTimeout <= 0 {
		v.addError("server.startup_timeout", cfg.Server.StartupTimeout, "must be positive")
	}

	if len(cfg.Server.JavaFlags) == 0 {
		v.addError("server.java_flags", cfg.Server.JavaFlags, "at least one Java flag should be specified")
	}
}

func (v *Validator) validateMods(cfg *Config) {
	if cfg.Mods.ConcurrentDownloads <= 0 {
		v.addError("mods.concurrent_downloads", cfg.Mods.ConcurrentDownloads, "must be positive")
	}

	if cfg.Mods.ConcurrentDownloads > 20 {
		v.addError("mods.concurrent_downloads", cfg.Mods.ConcurrentDownloads, "should not exceed 20 to avoid overwhelming servers")
	}

	if cfg.Mods.MaxRetries < 0 {
		v.addError("mods.max_retries", cfg.Mods.MaxRetries, "cannot be negative")
	}

	if cfg.Mods.RetryDelay < 0 {
		v.addError("mods.retry_delay", cfg.Mods.RetryDelay, "cannot be negative")
	}

	if cfg.Mods.Timeout <= 0 {
		v.addError("mods.timeout", cfg.Mods.Timeout, "must be positive")
	}

	// Validate Modrinth URLs
	for i, url := range cfg.Mods.ModrinthSources {
		if !strings.HasPrefix(url, "https://modrinth.com/mod/") {
			v.addError(fmt.Sprintf("mods.modrinth_sources[%d]", i), url, "must be a valid Modrinth mod URL")
		}
	}
}

func (v *Validator) validateBackup(cfg *Config) {
	if cfg.Backup.MaxBackups < 0 {
		v.addError("backup.max_backups", cfg.Backup.MaxBackups, "cannot be negative")
	}

	if cfg.Backup.CompressionLevel < 0 || cfg.Backup.CompressionLevel > 9 {
		v.addError("backup.compression_level", cfg.Backup.CompressionLevel, "must be between 0 and 9")
	}
}

func (v *Validator) validateNotifications(cfg *Config) {
	if cfg.Notifications.DiscordWebhook != "" {
		if !strings.HasPrefix(cfg.Notifications.DiscordWebhook, "https://discord.com/api/webhooks/") {
			v.addError("notifications.discord_webhook", cfg.Notifications.DiscordWebhook, "must be a valid Discord webhook URL")
		}
	}

	for i, interval := range cfg.Notifications.WarningIntervals {
		if interval <= 0 {
			v.addError(fmt.Sprintf("notifications.warning_intervals[%d]", i), interval, "must be positive")
		}
	}
}

func (v *Validator) validateLogging(cfg *Config) {
	validLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	if !v.containsString(validLevels, strings.ToUpper(cfg.Logging.Level)) {
		v.addError("logging.level", cfg.Logging.Level,
			fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")))
	}

	validFormats := []string{"json", "text"}
	if !v.containsString(validFormats, strings.ToLower(cfg.Logging.Format)) {
		v.addError("logging.format", cfg.Logging.Format,
			fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")))
	}

	if !cfg.Logging.ConsoleEnabled && !cfg.Logging.FileEnabled {
		v.addError("logging", "both disabled", "at least one logging output (console or file) must be enabled")
	}
}

func (v *Validator) addError(field string, value interface{}, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

func (v *Validator) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
