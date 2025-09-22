package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"craftops/internal/config"
)

// New creates a new logger based on the provided configuration.
func New(cfg *config.Config) *zap.Logger {
	level := parseLogLevel(cfg.Logging.Level)
	encoderConfig := createEncoderConfig(cfg.Logging.Format)
	cores := createCores(cfg, level, encoderConfig)

	return zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

// parseLogLevel converts string log level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARNING":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "CRITICAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// createEncoderConfig creates encoder configuration based on format
func createEncoderConfig(format string) zapcore.EncoderConfig {
	if format == "json" {
		return zap.NewProductionEncoderConfig()
	}

	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return config
}

// createCores creates logging cores based on configuration
func createCores(cfg *config.Config, level zapcore.Level, encoderConfig zapcore.EncoderConfig) []zapcore.Core {
	var cores []zapcore.Core

	// Console core
	if cfg.Logging.ConsoleEnabled {
		encoder := createEncoder(cfg.Logging.Format, encoderConfig)
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stderr),
			level,
		))
	}

	// File core
	if cfg.Logging.FileEnabled {
		if file := createLogFile(cfg.Paths.Logs); file != nil {
			encoder := createEncoder(cfg.Logging.Format, encoderConfig)
			cores = append(cores, zapcore.NewCore(
				encoder,
				zapcore.AddSync(file),
				level,
			))
		}
	}

	return cores
}

// createEncoder creates the appropriate encoder based on format
func createEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}

// createLogFile creates the log file and ensures the directory exists
func createLogFile(logsPath string) *os.File {
	// logsPath is expected to be a directory path; ensure it exists
	if err := os.MkdirAll(logsPath, 0755); err != nil {
		return nil
	}

	logFile := filepath.Join(logsPath, "craftops.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil
	}

	return file
}
