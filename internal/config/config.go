// Package config provides configuration management for craftops
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the main configuration object
type Config struct {
	Debug  bool `toml:"debug"`
	DryRun bool `toml:"dry_run"`

	Minecraft     MinecraftConfig    `toml:"minecraft"`
	Paths         PathsConfig        `toml:"paths"`
	Server        ServerConfig       `toml:"server"`
	Mods          ModsConfig         `toml:"mods"`
	Backup        BackupConfig       `toml:"backup"`
	Notifications NotificationConfig `toml:"notifications"`
	Logging       LoggingConfig      `toml:"logging"`
}

// MinecraftConfig contains settings for the Minecraft version and loader
type MinecraftConfig struct {
	Version   string `toml:"version"`
	Modloader string `toml:"modloader"`
}

// PathsConfig defines core directory locations
type PathsConfig struct {
	Server  string `toml:"server"`
	Mods    string `toml:"mods"`
	Backups string `toml:"backups"`
	Logs    string `toml:"logs"`
}

// ServerConfig contains JVM and lifecycle settings
type ServerConfig struct {
	JarName        string   `toml:"jar_name"`
	JavaFlags      []string `toml:"java_flags"`
	StopCommand    string   `toml:"stop_command"`
	MaxStopWait    int      `toml:"max_stop_wait"`
	StartupTimeout int      `toml:"startup_timeout"`
	SessionName    string   `toml:"session_name"`
}

// ModsConfig contains mod update settings
type ModsConfig struct {
	ConcurrentDownloads int      `toml:"concurrent_downloads"`
	MaxRetries          int      `toml:"max_retries"`
	RetryDelay          float64  `toml:"retry_delay"`
	Timeout             int      `toml:"timeout"`
	ModrinthSources     []string `toml:"modrinth_sources"`
}

// BackupConfig contains backup and retention settings
type BackupConfig struct {
	Enabled          bool     `toml:"enabled"`
	MaxBackups       int      `toml:"max_backups"`
	CompressionLevel int      `toml:"compression_level"`
	IncludeLogs      bool     `toml:"include_logs"`
	ExcludePatterns  []string `toml:"exclude_patterns"`
}

// NotificationConfig contains webhook and alert settings
type NotificationConfig struct {
	DiscordWebhook       string `toml:"discord_webhook"`
	WarningIntervals     []int  `toml:"warning_intervals"`
	WarningMessage       string `toml:"warning_message"`
	SuccessNotifications bool   `toml:"success_notifications"`
	ErrorNotifications   bool   `toml:"error_notifications"`
}

// LoggingConfig defines log output levels and formats
type LoggingConfig struct {
	Level          string `toml:"level"`
	Format         string `toml:"format"`
	FileEnabled    bool   `toml:"file_enabled"`
	ConsoleEnabled bool   `toml:"console_enabled"`
}

// DefaultConfig returns a configuration with production-ready defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	serverPath := filepath.Join(homeDir, "minecraft", "server")

	return &Config{
		Debug:  false,
		DryRun: false,
		Minecraft: MinecraftConfig{
			Version:   "1.20.1",
			Modloader: "fabric",
		},
		Paths: PathsConfig{
			Server:  serverPath,
			Mods:    filepath.Join(serverPath, "mods"),
			Backups: filepath.Join(homeDir, "minecraft", "backups"),
			Logs:    filepath.Join(homeDir, ".local", "share", "craftops", "logs"),
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
			SessionName:    "minecraft",
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
			IncludeLogs:      false,
			ExcludePatterns: []string{
				"*.log", "*.log.*", "cache/", "temp/",
				".DS_Store", "Thumbs.db",
			},
		},
		Notifications: NotificationConfig{
			DiscordWebhook:       "",
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

// LoadConfig loads configuration from a file or fallback paths
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if configPath == "" {
		configPath = findDefaultConfig()
	}
	if configPath != "" {
		if _, err := toml.DecodeFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// SaveConfig writes the configuration to a TOML file
func (c *Config) SaveConfig(configPath string) error {
	file, err := os.Create(configPath) //nolint:gosec // config path is user-controlled
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		_ = file.Close() // Close errors are non-critical after successful encoding
	}()

	return toml.NewEncoder(file).Encode(c)
}

// Validate ensures settings are within supported bounds
func (c *Config) Validate() error {
	if err := c.validateModloader(); err != nil {
		return err
	}
	return c.validateLogging()
}

func findDefaultConfig() string {
	candidates := []string{"config.toml"}

	if cfgDir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(cfgDir, "craftops", "config.toml"))
	}
	candidates = append(candidates, "/etc/craftops/config.toml")

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (c *Config) validateModloader() error {
	valid := []string{"fabric", "forge", "quilt", "neoforge"}
	modloader := strings.ToLower(c.Minecraft.Modloader)
	if !slices.Contains(valid, modloader) {
		return fmt.Errorf("unsupported modloader: %s. Must be one of %v", c.Minecraft.Modloader, valid)
	}
	c.Minecraft.Modloader = modloader
	return nil
}

func (c *Config) validateLogging() error {
	validLevels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
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
	return nil
}
