package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// ServerService handles server management operations
type ServerService struct {
	config *config.Config
	logger *zap.Logger
}

// ServerStatus represents server status information
type ServerStatus struct {
	IsRunning   bool   `json:"is_running"`
	PID         *int   `json:"pid,omitempty"`
	Uptime      string `json:"uptime,omitempty"`
	MemoryUsage string `json:"memory_usage,omitempty"`
}

// NewServerService creates a new server service instance
func NewServerService(cfg *config.Config, logger *zap.Logger) *ServerService {
	return &ServerService{
		config: cfg,
		logger: logger,
	}
}

// HealthCheck performs health checks for the server service
func (ss *ServerService) HealthCheck(ctx context.Context) []HealthCheck {
	checks := make([]HealthCheck, 0, 3)

	// Check server directory
	checks = append(checks, ss.checkServerDirectory())
	// Check server JAR
	checks = append(checks, ss.checkServerJAR())
	// Check Java availability
	checks = append(checks, ss.checkJavaRuntime(ctx))
	return checks
}

func (ss *ServerService) checkServerDirectory() HealthCheck {
	if info, err := os.Stat(ss.config.Paths.Server); err == nil && info.IsDir() {
		return HealthCheck{
			Name:    "Server directory",
			Status:  "✅",
			Message: "OK",
		}
	}
	return HealthCheck{
		Name:    "Server directory",
		Status:  "❌",
		Message: "Directory not found",
	}
}

func (ss *ServerService) checkServerJAR() HealthCheck {
	serverJar := filepath.Join(ss.config.Paths.Server, ss.config.Server.JarName)
	if info, err := os.Stat(serverJar); err == nil && !info.IsDir() {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		return HealthCheck{
			Name:    "Server JAR",
			Status:  "✅",
			Message: fmt.Sprintf("Found (%.1f MB)", sizeMB),
		}
	}
	return HealthCheck{
		Name:    "Server JAR",
		Status:  "❌",
		Message: fmt.Sprintf("Not found: %s", ss.config.Server.JarName),
	}
}

func (ss *ServerService) checkJavaRuntime(ctx context.Context) HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "java", "-version")
	if err := cmd.Run(); err == nil {
		return HealthCheck{
			Name:    "Java Runtime",
			Status:  "✅",
			Message: "Available",
		}
	}
	return HealthCheck{
		Name:    "Java Runtime",
		Status:  "❌",
		Message: "Java not found or not working",
	}
}

// GetStatus gets the current server status
func (ss *ServerService) GetStatus(ctx context.Context) (*ServerStatus, error) {
	cmd := exec.CommandContext(ctx, "screen", "-ls")
	output, err := cmd.Output()
	if err != nil {
		ss.logger.Error("Error checking server status", zap.Error(err))
		return &ServerStatus{IsRunning: false}, nil
	}

	isRunning := strings.Contains(string(output), "minecraft")
	return &ServerStatus{IsRunning: isRunning}, nil
}

// Start starts the Minecraft server
func (ss *ServerService) Start(ctx context.Context) error {
	if ss.config.DryRun {
		ss.logger.Info("Dry run: Would start server")
		return nil
	}

	// Check if server is already running
	if status, err := ss.GetStatus(ctx); err != nil {
		return fmt.Errorf("failed to check server status: %w", err)
	} else if status.IsRunning {
		ss.logger.Warn("Server is already running")
		return nil
	}

	// Validate server JAR exists
	serverJar := filepath.Join(ss.config.Paths.Server, ss.config.Server.JarName)
	if _, err := os.Stat(serverJar); os.IsNotExist(err) {
		return fmt.Errorf("server JAR not found: %s", serverJar)
	}
	// Start server in screen session
	javaArgs := append(ss.config.Server.JavaFlags, "-jar", ss.config.Server.JarName, "nogui")
	cmdArgs := append([]string{"-dmS", "minecraft", "java"}, javaArgs...)

	cmd := exec.CommandContext(ctx, "screen", cmdArgs...)
	cmd.Dir = ss.config.Paths.Server
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Verify server started
	time.Sleep(2 * time.Second)
	if status, err := ss.GetStatus(ctx); err != nil {
		return fmt.Errorf("failed to verify server start: %w", err)
	} else if !status.IsRunning {
		return fmt.Errorf("server failed to start")
	}

	ss.logger.Info("Server started successfully")
	return nil
}

// Stop stops the Minecraft server
func (ss *ServerService) Stop(ctx context.Context) error {
	if ss.config.DryRun {
		ss.logger.Info("Dry run: Would stop server")
		return nil
	}

	// Check if server is running
	if status, err := ss.GetStatus(ctx); err != nil {
		return fmt.Errorf("failed to check server status: %w", err)
	} else if !status.IsRunning {
		ss.logger.Warn("Server is not running")
		return nil
	}

	// Send stop command
	stopCmd := fmt.Sprintf("%s\n", ss.config.Server.StopCommand)
	cmd := exec.CommandContext(ctx, "screen", "-S", "minecraft", "-X", "stuff", stopCmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send stop command: %w", err)
	}

	// Wait for server to stop
	return ss.waitForStop(ctx)
}

func (ss *ServerService) waitForStop(ctx context.Context) error {
	maxWait := time.Duration(ss.config.Server.MaxStopWait) * time.Second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	timeout := time.After(maxWait)
	startTime := time.Now()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("server failed to stop within timeout (%d seconds)", ss.config.Server.MaxStopWait)
		case <-ticker.C:
			status, err := ss.GetStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to check server status: %w", err)
			}
			if !status.IsRunning {
				waitTime := time.Since(startTime)
				ss.logger.Info("Server stopped successfully", zap.Duration("wait_time", waitTime))
				return nil
			}
		}
	}
}

// Restart restarts the Minecraft server
func (ss *ServerService) Restart(ctx context.Context) error {
	ss.logger.Info("Restarting server")

	if err := ss.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	time.Sleep(2 * time.Second)

	if err := ss.Start(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
