// Package config handles TOML configuration loading, defaults, and validation.
package config

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"

	"craftops/internal/domain"
)

// Config is the top-level application configuration.
type Config struct {
	Paths         PathsConfig        `toml:"paths"`
	Minecraft     MinecraftConfig    `toml:"minecraft"`
	Logging       LoggingConfig      `toml:"logging"`
	Server        ServerConfig       `toml:"server"`
	Notifications NotificationConfig `toml:"notifications"`
	Backup        BackupConfig       `toml:"backup"`
	Mods          ModsConfig         `toml:"mods"`
	DryRun        bool               `toml:"dry_run"`
}

// MinecraftConfig specifies game version and mod loader.
type MinecraftConfig struct {
	Version   string `toml:"version"`
	Modloader string `toml:"modloader"`
}

// PathsConfig defines filesystem locations.
type PathsConfig struct {
	Server  string `toml:"server"`
	Mods    string `toml:"mods"`
	Backups string `toml:"backups"`
	Logs    string `toml:"logs"`
}

// ServerConfig holds JVM flags and lifecycle settings.
type ServerConfig struct {
	JarName        string   `toml:"jar_name"`
	StopCommand    string   `toml:"stop_command"`
	SessionName    string   `toml:"session_name"`
	JavaFlags      []string `toml:"java_flags"`
	MaxStopWait    int      `toml:"max_stop_wait"`
	StartupTimeout int      `toml:"startup_timeout"`
}

// ModsConfig controls mod update behavior.
type ModsConfig struct {
	ModrinthSources     []string `toml:"modrinth_sources"`
	ConcurrentDownloads int      `toml:"concurrent_downloads"`
	MaxRetries          int      `toml:"max_retries"`
	RetryDelay          float64  `toml:"retry_delay"`
	Timeout             int      `toml:"timeout"`
}

// BackupConfig controls backup creation and retention.
type BackupConfig struct {
	ExcludePatterns  []string `toml:"exclude_patterns"`
	MaxBackups       int      `toml:"max_backups"`
	CompressionLevel int      `toml:"compression_level"`
	Enabled          bool     `toml:"enabled"`
	IncludeLogs      bool     `toml:"include_logs"`
}

// NotificationConfig controls Discord webhook alerts.
type NotificationConfig struct {
	DiscordWebhook       string `toml:"discord_webhook"`
	WarningMessage       string `toml:"warning_message"`
	WarningIntervals     []int  `toml:"warning_intervals"`
	Timeout              int    `toml:"timeout"`
	SuccessNotifications bool   `toml:"success_notifications"`
	ErrorNotifications   bool   `toml:"error_notifications"`
}

// LoggingConfig controls log output.
type LoggingConfig struct {
	Level          string `toml:"level"`
	Format         string `toml:"format"`
	FileEnabled    bool   `toml:"file_enabled"`
	ConsoleEnabled bool   `toml:"console_enabled"`
}

// DefaultConfig returns production-ready defaults. Inside a container
// (detected via /.dockerenv) paths target the image layout under /minecraft.
func DefaultConfig() *Config {
	serverPath := filepath.Join(homeDir(), "minecraft", "server")
	if containerServerDir := containerServerDir(); containerServerDir != "" {
		serverPath = containerServerDir
	}

	return &Config{
		Minecraft: MinecraftConfig{
			Version:   "1.20.1",
			Modloader: "fabric",
		},
		Paths: PathsConfig{
			Server:  serverPath,
			Mods:    filepath.Join(serverPath, "mods"),
			Backups: filepath.Join(filepath.Dir(serverPath), "backups"),
			Logs:    filepath.Join(homeDir(), ".local", "share", "craftops", "logs"),
		},
		Server: ServerConfig{
			JarName: "server.jar",
			JavaFlags: []string{
				"-Xms4G", "-Xmx4G", "-XX:+UseG1GC",
				"-XX:+ParallelRefProcEnabled", "-XX:+UnlockExperimentalVMOptions",
				"-XX:+DisableExplicitGC", "-XX:+AlwaysPreTouch",
			},
			StopCommand:    "stop",
			MaxStopWait:    300,
			StartupTimeout: 120,
			SessionName:    domain.DefaultSessionName,
		},
		Mods: ModsConfig{
			ConcurrentDownloads: 5,
			MaxRetries:          3,
			RetryDelay:          2.0,
			Timeout:             30,
			ModrinthSources:     []string{},
		},
		Backup: BackupConfig{
			Enabled:          true,
			MaxBackups:       5,
			CompressionLevel: 6,
			ExcludePatterns: []string{
				"*.log", "*.log.*", "cache/", "temp/",
				".DS_Store", "Thumbs.db",
			},
		},
		Notifications: NotificationConfig{
			Timeout:              30,
			WarningIntervals:     []int{15, 10, 5, 1},
			WarningMessage:       "Server will restart in {minutes} minute(s) for mod updates",
			SuccessNotifications: true,
			ErrorNotifications:   true,
		},
		Logging: LoggingConfig{
			Level:          "INFO",
			Format:         "json",
			FileEnabled:    true,
			ConsoleEnabled: true,
		},
	}
}

// LoadConfig reads config from file (or defaults) and validates it.
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath == "" {
		configPath = findDefaultConfig()
	}
	if configPath != "" {
		if _, err := toml.DecodeFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// SaveConfig writes the configuration as TOML.
func (c *Config) SaveConfig(configPath string) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := os.WriteFile(configPath, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// Validate checks that all settings are within supported bounds and normalizes case.
func (c *Config) Validate() error {
	valid := []string{"fabric", "forge", "quilt", "neoforge"}
	modloader := strings.ToLower(c.Minecraft.Modloader)
	if !slices.Contains(valid, modloader) {
		return fmt.Errorf("unsupported modloader: %s. Must be one of %v", c.Minecraft.Modloader, valid)
	}
	c.Minecraft.Modloader = modloader

	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	level := strings.ToUpper(c.Logging.Level)
	if !slices.Contains(validLevels, level) {
		return fmt.Errorf("invalid log level: %s. Must be one of %v", c.Logging.Level, validLevels)
	}
	c.Logging.Level = level

	validFormats := []string{"json", "text"}
	format := strings.ToLower(c.Logging.Format)
	if !slices.Contains(validFormats, format) {
		return fmt.Errorf("invalid log format: %s. Must be one of %v", c.Logging.Format, validFormats)
	}
	c.Logging.Format = format

	for _, v := range []struct {
		name  string
		value int
		min   int
	}{
		{"mods.concurrent_downloads", c.Mods.ConcurrentDownloads, 1},
		{"mods.timeout", c.Mods.Timeout, 1},
		{"mods.max_retries", c.Mods.MaxRetries, 0},
		{"server.max_stop_wait", c.Server.MaxStopWait, 0},
		{"server.startup_timeout", c.Server.StartupTimeout, 0},
		{"backup.max_backups", c.Backup.MaxBackups, 0},
	} {
		if err := validateAtLeast(v.name, v.value, v.min); err != nil {
			return err
		}
	}

	if c.Mods.RetryDelay < 0 {
		return fmt.Errorf("mods.retry_delay must be at least 0, got %g", c.Mods.RetryDelay)
	}

	if err := validateRange("backup.compression_level", c.Backup.CompressionLevel, gzip.NoCompression, gzip.BestCompression); err != nil {
		return err
	}
	return nil
}

// validateAtLeast rejects values below min.
func validateAtLeast(name string, value, minValue int) error {
	if value < minValue {
		return fmt.Errorf("%s must be at least %d, got %d", name, minValue, value)
	}
	return nil
}

// validateRange rejects values outside [min, max].
func validateRange(name string, value, minValue, maxValue int) error {
	if value < minValue || value > maxValue {
		return fmt.Errorf("%s must be between %d and %d, got %d", name, minValue, maxValue, value)
	}
	return nil
}

func findDefaultConfig() string {
	candidates := []string{"config.toml"}
	if cfgDir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(cfgDir, "craftops", "config.toml"))
	}
	candidates = append(candidates, "/config/config.toml", "/etc/craftops/config.toml")

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// containerServerDir returns the container image's server root when running
// inside a container that provides it, or "" on a normal host.
func containerServerDir() string {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return ""
	}
	const root = "/minecraft/server"
	if _, err := os.Stat(root); err != nil {
		return ""
	}
	return root
}
