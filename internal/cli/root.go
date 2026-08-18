// Package cli wires Cobra commands to service layer operations.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"craftops/internal/config"
	"craftops/internal/domain"
	"craftops/internal/service"
	"craftops/internal/ui"
)

// Version is set by ldflags during build.
var Version = "dev"

// app bundles the services shared by all commands for a single run.
type app struct {
	Config       *config.Config
	Logger       *zap.Logger
	Terminal     *ui.Terminal
	Server       *service.Server
	Mods         *service.Mods
	Backup       *service.Backup
	Notification *service.Notification
}

func newApp(cfg *config.Config) *app {
	logger := newLogger(cfg)
	return &app{
		Config:       cfg,
		Logger:       logger,
		Terminal:     ui.NewTerminal(),
		Server:       service.NewServer(cfg, logger),
		Mods:         service.NewMods(cfg, logger),
		Backup:       service.NewBackup(cfg, logger),
		Notification: service.NewNotification(cfg, logger),
	}
}

func (a *app) Close() {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}

func newLogger(cfg *config.Config) *zap.Logger {
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if cfg.Logging.Level == "DEBUG" {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	if cfg.Logging.Format == "text" {
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stderr),
		level,
	)

	var cores []zapcore.Core
	if cfg.Logging.FileEnabled && cfg.Paths.Logs != "" {
		if err := os.MkdirAll(cfg.Paths.Logs, domain.DirPerm); err != nil {
			loggerWarnf("failed to create log directory %s: %v", cfg.Paths.Logs, err)
		} else {
			logPath := filepath.Join(cfg.Paths.Logs, "craftops.log")
			f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
			if err != nil {
				loggerWarnf("failed to open log file %s: %v", logPath, err)
			} else {
				enc := zapcore.NewJSONEncoder(encoderCfg)
				if cfg.Logging.Format == "text" {
					enc = zapcore.NewConsoleEncoder(encoderCfg)
				}
				cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(f), level))
			}
		}
	}

	// Fall back to the console when nothing else is configured.
	if cfg.Logging.ConsoleEnabled || len(cores) == 0 {
		cores = append(cores, consoleCore)
	}
	return zap.New(zapcore.NewTee(cores...))
}

func loggerWarnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "craftops: warning: "+format+"\n", args...)
}

// activeApp holds the initialized app for the duration of a command run.
// It is set in PersistentPreRunE and cleared in PersistentPostRun.
var activeApp *app

var rootCmd = &cobra.Command{
	Use:               "craftops",
	Short:             "Modern Minecraft server operations and mod management",
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: initApp,
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		if activeApp != nil {
			activeApp.Close()
			activeApp = nil
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file path")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
	rootCmd.PersistentFlags().Bool("dry-run", false, "show what would be done")
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("CraftOps v{{.Version}}\n")
	rootCmd.Run = func(cmd *cobra.Command, _ []string) { _ = cmd.Help() }

	rootCmd.AddCommand(serverCmd, modsCmd, backupCmd, healthCmd, initCmd)
	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd, serverStatusCmd)
	modsCmd.AddCommand(modsUpdateCmd, modsListCmd)
	backupCmd.AddCommand(backupCreateCmd, backupListCmd, backupDeleteCmd)

	modsUpdateCmd.Flags().Bool("force", false, "force update even if mod is current")
	modsUpdateCmd.Flags().Bool("no-backup", false, "skip pre-update backup")
	initCmd.Flags().StringP("output", "o", "", "config file output path")
	initCmd.Flags().Bool("force", false, "overwrite existing config file")
}

// Execute runs the root command.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func initApp(cmd *cobra.Command, _ []string) error {
	cfgFile, _ := cmd.Flags().GetString("config")
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		cfg.Logging.Level = "DEBUG"
	}
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		cfg.DryRun = true
	}

	activeApp = newApp(cfg)
	return nil
}

// appFrom returns the initialized app, or nil if initApp was skipped.
func appFrom() *app {
	return activeApp
}
