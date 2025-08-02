package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"craftops/internal/config"
)

var (
	cfgFile string
	debug   bool
	dryRun  bool
	cfg     *config.Config
	logger  *zap.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "craftops",
	Short: "🎮 Modern Minecraft server operations and mod management",
	Long: `CraftOps is a comprehensive CLI tool for Minecraft server operations and mod management.

Features:
• Server lifecycle management (start, stop, restart) - Linux/macOS only
• Automated backups with retention policies
• Discord notifications and player warnings
• Health checks and configuration validation`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Override config with CLI flags
		if debug {
			cfg.Debug = true
			cfg.Logging.Level = "DEBUG"
		}
		if dryRun {
			cfg.DryRun = true
		}

		// Initialize logger
		logger = initLogger(cfg)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "show version information")
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if version, _ := cmd.Flags().GetBool("version"); version {
			fmt.Println("CraftOps v2.0.0")
			return
		}
		cmd.Help()
	}
}

// initLogger initializes the logger based on configuration
func initLogger(cfg *config.Config) *zap.Logger {
	// Parse log level
	var level zapcore.Level
	switch cfg.Logging.Level {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARNING":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "CRITICAL":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create encoder config
	var encoderConfig zapcore.EncoderConfig
	if cfg.Logging.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Create cores
	var cores []zapcore.Core

	// Console core
	if cfg.Logging.ConsoleEnabled {
		var encoder zapcore.Encoder
		if cfg.Logging.Format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stderr),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File core
	if cfg.Logging.FileEnabled {
		// Ensure log directory exists
		logDir := filepath.Dir(cfg.Paths.Logs)
		os.MkdirAll(logDir, 0755)

		logFile := filepath.Join(cfg.Paths.Logs, "craftops.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			var encoder zapcore.Encoder
			if cfg.Logging.Format == "json" {
				encoder = zapcore.NewJSONEncoder(encoderConfig)
			} else {
				encoder = zapcore.NewConsoleEncoder(encoderConfig)
			}

			fileCore := zapcore.NewCore(
				encoder,
				zapcore.AddSync(file),
				level,
			)
			cores = append(cores, fileCore)
		}
	}

	// Create logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}

// getContext returns a context for operations
func getContext() context.Context {
	return context.Background()
}

// Color scheme
var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	infoColor    = color.New(color.FgCyan, color.Bold)
	headerColor  = color.New(color.FgMagenta, color.Bold)
	accentColor  = color.New(color.FgBlue, color.Bold)
	dimColor     = color.New(color.FgHiBlack)
)

// printBanner prints a beautiful banner
func printBanner(title string) {
	width := 60
	padding := (width - len(title) - 4) / 2

	headerColor.Println(strings.Repeat("═", width))
	headerColor.Printf("║%s🎮 %s 🎮%s║\n",
		strings.Repeat(" ", padding),
		title,
		strings.Repeat(" ", padding))
	headerColor.Println(strings.Repeat("═", width))
	fmt.Println()
}

// printSection prints a section header
func printSection(title string) {
	accentColor.Printf("\n▶ %s\n", title)
	dimColor.Println(strings.Repeat("─", len(title)+2))
}

// printSuccess prints a success message
func printSuccess(message string) {
	successColor.Printf("✅ %s\n", message)
}

// printError prints an error message
func printError(message string) {
	errorColor.Printf("❌ %s\n", message)
}

// printWarning prints a warning message
func printWarning(message string) {
	warningColor.Printf("⚠️  %s\n", message)
}

// printInfo prints an info message
func printInfo(message string) {
	infoColor.Printf("ℹ️  %s\n", message)
}

// printStep prints a step in a process
func printStep(step int, total int, message string) {
	accentColor.Printf("[%d/%d] ", step, total)
	fmt.Printf("%s\n", message)
}

// printTable prints a formatted table
func printTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	accentColor.Print("┌")
	for i, width := range widths {
		accentColor.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			accentColor.Print("┬")
		}
	}
	accentColor.Println("┐")

	accentColor.Print("│")
	for i, header := range headers {
		fmt.Printf(" %-*s ", widths[i], header)
		accentColor.Print("│")
	}
	fmt.Println()

	accentColor.Print("├")
	for i, width := range widths {
		accentColor.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			accentColor.Print("┼")
		}
	}
	accentColor.Println("┤")

	// Print rows
	for _, row := range rows {
		accentColor.Print("│")
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf(" %-*s ", widths[i], cell)
				accentColor.Print("│")
			}
		}
		fmt.Println()
	}

	accentColor.Print("└")
	for i, width := range widths {
		accentColor.Print(strings.Repeat("─", width+2))
		if i < len(widths)-1 {
			accentColor.Print("┴")
		}
	}
	accentColor.Println("┘")
}
