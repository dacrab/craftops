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
	IsRunning bool `json:"is_running"`
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
	checks := make([]HealthCheck, 0, 4)

	// Check server directory
	checks = append(checks, ss.checkServerDirectory())
	// Check server JAR
	checks = append(checks, ss.checkServerJAR())
	// Check Java availability
	checks = append(checks, ss.checkJavaRuntime(ctx))
	// Check screen availability
	checks = append(checks, ss.checkScreen(ctx))
	return checks
}

func (ss *ServerService) checkServerDirectory() HealthCheck {
	if info, err := os.Stat(ss.config.Paths.Server); err == nil && info.IsDir() {
		return HealthCheck{
			Name:    "Server directory",
            Status:  "OK",
			Message: "OK",
		}
	}
	return HealthCheck{
		Name:    "Server directory",
        Status:  "ERROR",
		Message: "Directory not found",
	}
}

func (ss *ServerService) checkServerJAR() HealthCheck {
	serverJar := filepath.Join(ss.config.Paths.Server, ss.config.Server.JarName)
	if info, err := os.Stat(serverJar); err == nil && !info.IsDir() {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		return HealthCheck{
			Name:    "Server JAR",
            Status:  "OK",
			Message: fmt.Sprintf("Found (%.1f MB)", sizeMB),
		}
	}
	return HealthCheck{
		Name:    "Server JAR",
        Status:  "ERROR",
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
            Status:  "OK",
			Message: "Available",
		}
	}
	return HealthCheck{
		Name:    "Java Runtime",
        Status:  "ERROR",
		Message: "Java not found or not working",
	}
}

func (ss *ServerService) checkScreen(ctx context.Context) HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "screen", "-version")
	if err := cmd.Run(); err == nil {
		return HealthCheck{
			Name:    "GNU screen",
            Status:  "OK",
			Message: "Available",
		}
	}
	return HealthCheck{
		Name:    "GNU screen",
        Status:  "ERROR",
		Message: "screen not found; required for server control",
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

	session := ss.config.Server.SessionName
	if session == "" {
		session = "minecraft"
	}
	// screen -ls output typically includes lines like: "\t12345.<session>\t(Detached)"
	isRunning := strings.Contains(string(output), "."+session)
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
	session := ss.config.Server.SessionName
	if session == "" {
		session = "minecraft"
	}
	cmdArgs := append([]string{"-dmS", session, "java"}, javaArgs...)

	cmd := exec.CommandContext(ctx, "screen", cmdArgs...)
	cmd.Dir = ss.config.Paths.Server
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Verify server started within StartupTimeout
	timeoutSec := ss.config.Server.StartupTimeout
	if timeoutSec <= 0 {
		timeoutSec = 15
	}
	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for server start: %w", ctx.Err())
		case <-ticker.C:
			status, err := ss.GetStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to verify server start: %w", err)
			}
			if status.IsRunning {
				ss.logger.Info("Server started successfully", zap.Duration("startup_time", time.Since(start)))
				return nil
			}
			if time.Since(start) > time.Duration(timeoutSec)*time.Second {
				return fmt.Errorf("server failed to start within timeout (%ds)", timeoutSec)
			}
		}
	}
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
	session := ss.config.Server.SessionName
	if session == "" {
		session = "minecraft"
	}
	cmd := exec.CommandContext(ctx, "screen", "-S", session, "-X", "stuff", stopCmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send stop command: %w", err)
	}

	// Wait for server to stop
	return ss.waitForStop(ctx)
}

func (ss *ServerService) waitForStop(ctx context.Context) error {
    maxWait := time.Duration(ss.config.Server.MaxStopWait) * time.Second
    ctx, cancel := context.WithTimeout(ctx, maxWait)
    defer cancel()
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    startTime := time.Now()

    for {
        select {
        case <-ctx.Done():
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
