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
	checks := []HealthCheck{}

	// Check server directory
	serverDir := ss.config.Paths.Server
	if info, err := os.Stat(serverDir); err == nil && info.IsDir() {
		checks = append(checks, HealthCheck{
			Name:    "Server directory",
			Status:  "✅",
			Message: "OK",
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Server directory",
			Status:  "❌",
			Message: "Directory not found",
		})
	}

	// Check server JAR
	serverJar := filepath.Join(ss.config.Paths.Server, ss.config.Server.JarName)
	if info, err := os.Stat(serverJar); err == nil && !info.IsDir() {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		checks = append(checks, HealthCheck{
			Name:    "Server JAR",
			Status:  "✅",
			Message: fmt.Sprintf("Found (%.1f MB)", sizeMB),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Server JAR",
			Status:  "❌",
			Message: fmt.Sprintf("Not found: %s", ss.config.Server.JarName),
		})
	}

	// Check Java availability
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "java", "-version")
	if err := cmd.Run(); err == nil {
		checks = append(checks, HealthCheck{
			Name:    "Java Runtime",
			Status:  "✅",
			Message: "Available",
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Java Runtime",
			Status:  "❌",
			Message: "Java not found or not working",
		})
	}

	return checks
}

// GetStatus gets the current server status
func (ss *ServerService) GetStatus(ctx context.Context) (*ServerStatus, error) {
	// Check if server process is running via screen session
	cmd := exec.CommandContext(ctx, "screen", "-ls")
	output, err := cmd.Output()
	if err != nil {
		ss.logger.Error("Error checking server status", zap.Error(err))
		return &ServerStatus{IsRunning: false}, nil
	}

	// Check if minecraft screen session exists
	if strings.Contains(string(output), "minecraft") {
		return &ServerStatus{
			IsRunning: true,
			PID:       nil,
			Uptime:    "Unknown",
		}, nil
	}

	return &ServerStatus{IsRunning: false}, nil
}

// Start starts the Minecraft server
func (ss *ServerService) Start(ctx context.Context) error {
	if ss.config.DryRun {
		ss.logger.Info("Dry run: Would start server")
		return nil
	}

	// Check if server is already running
	status, err := ss.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check server status: %w", err)
	}

	if status.IsRunning {
		ss.logger.Warn("Server is already running")
		return nil
	}

	// Check if server JAR exists
	serverJar := filepath.Join(ss.config.Paths.Server, ss.config.Server.JarName)
	if _, err := os.Stat(serverJar); os.IsNotExist(err) {
		return fmt.Errorf("server JAR not found: %s", serverJar)
	}

	// Build the start command
	javaArgs := append(ss.config.Server.JavaFlags, "-jar", ss.config.Server.JarName, "nogui")
	cmdArgs := append([]string{"-dmS", "minecraft", "java"}, javaArgs...)

	// Start server in screen session
	cmd := exec.CommandContext(ctx, "screen", cmdArgs...)
	cmd.Dir = ss.config.Paths.Server

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait a moment for the process to start
	time.Sleep(2 * time.Second)

	// Check if server started successfully
	status, err = ss.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify server start: %w", err)
	}

	if status.IsRunning {
		ss.logger.Info("Server started successfully")
		return nil
	}

	return fmt.Errorf("server failed to start")
}

// Stop stops the Minecraft server
func (ss *ServerService) Stop(ctx context.Context) error {
	if ss.config.DryRun {
		ss.logger.Info("Dry run: Would stop server")
		return nil
	}

	// Check if server is running
	status, err := ss.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check server status: %w", err)
	}

	if !status.IsRunning {
		ss.logger.Warn("Server is not running")
		return nil
	}

	// Send stop command to screen session
	stopCmd := fmt.Sprintf("%s\n", ss.config.Server.StopCommand)
	cmd := exec.CommandContext(ctx, "screen", "-S", "minecraft", "-X", "stuff", stopCmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send stop command: %w", err)
	}

	// Wait for server to stop
	maxWait := time.Duration(ss.config.Server.MaxStopWait) * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		time.Sleep(1 * time.Second)

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

	return fmt.Errorf("server failed to stop within timeout (%d seconds)", ss.config.Server.MaxStopWait)
}

// Restart restarts the Minecraft server
func (ss *ServerService) Restart(ctx context.Context) error {
	ss.logger.Info("Restarting server")

	// Stop the server
	if err := ss.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	// Wait a moment between stop and start
	time.Sleep(2 * time.Second)

	// Start the server
	if err := ss.Start(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
