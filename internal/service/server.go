package service

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
	"craftops/internal/domain"
)

// Server implements ServerManager
type Server struct {
	cfg    *config.Config
	logger *zap.Logger
}

var _ ServerManager = (*Server)(nil)

// NewServer creates a new server service
func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{cfg: cfg, logger: logger}
}

// Status returns the current server status
func (s *Server) Status(ctx context.Context) (*domain.ServerStatus, error) {
	cmd := exec.CommandContext(ctx, "screen", "-ls")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Debug("screen -ls returned error (may be normal)", zap.Error(err))
	}

	session := s.sessionName()
	isRunning := strings.Contains(string(output), "."+session)

	return &domain.ServerStatus{
		IsRunning:   isRunning,
		SessionName: session,
		CheckedAt:   time.Now(),
	}, nil
}

// Start starts the Minecraft server
func (s *Server) Start(ctx context.Context) error {
	if s.cfg.DryRun {
		s.logger.Info("Dry run: Would start server")
		return nil
	}

	status, err := s.Status(ctx)
	if err != nil {
		return domain.NewServiceError("server", "start", err)
	}
	if status.IsRunning {
		s.logger.Warn("Server is already running")
		return nil
	}

	serverJar := filepath.Join(s.cfg.Paths.Server, s.cfg.Server.JarName)
	if _, err := os.Stat(serverJar); os.IsNotExist(err) {
		return domain.ErrServerJarNotFound
	}

	javaArgs := append(s.cfg.Server.JavaFlags, "-jar", s.cfg.Server.JarName, "nogui")
	cmdArgs := append([]string{"-dmS", s.sessionName(), "java"}, javaArgs...)

	cmd := exec.CommandContext(ctx, "screen", cmdArgs...)
	cmd.Dir = s.cfg.Paths.Server
	if err := cmd.Start(); err != nil {
		return domain.NewServiceError("server", "start", err)
	}

	return s.waitForStart(ctx)
}

// Stop stops the Minecraft server
func (s *Server) Stop(ctx context.Context) error {
	if s.cfg.DryRun {
		s.logger.Info("Dry run: Would stop server")
		return nil
	}

	status, err := s.Status(ctx)
	if err != nil {
		return domain.NewServiceError("server", "stop", err)
	}
	if !status.IsRunning {
		s.logger.Warn("Server is not running")
		return nil
	}

	stopCmd := fmt.Sprintf("%s\n", s.cfg.Server.StopCommand)
	cmd := exec.CommandContext(ctx, "screen", "-S", s.sessionName(), "-X", "stuff", stopCmd)
	if err := cmd.Run(); err != nil {
		return domain.NewServiceError("server", "stop", err)
	}

	return s.waitForStop(ctx)
}

// Restart restarts the Minecraft server
func (s *Server) Restart(ctx context.Context) error {
	s.logger.Info("Restarting server")

	if err := s.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	time.Sleep(2 * time.Second)

	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// HealthCheck performs health checks
func (s *Server) HealthCheck(ctx context.Context) []domain.HealthCheck {
	checks := make([]domain.HealthCheck, 0, 4)
	checks = append(checks, s.checkServerDirectory())
	checks = append(checks, s.checkServerJAR())
	checks = append(checks, s.checkJava(ctx))
	checks = append(checks, s.checkScreen(ctx))
	return checks
}

func (s *Server) sessionName() string {
	if s.cfg.Server.SessionName != "" {
		return s.cfg.Server.SessionName
	}
	return "minecraft"
}

func (s *Server) waitForStart(ctx context.Context) error {
	timeout := s.cfg.Server.StartupTimeout
	if timeout <= 0 {
		timeout = 15
	}

	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := s.Status(ctx)
			if err != nil {
				return err
			}
			if status.IsRunning {
				s.logger.Info("Server started", zap.Duration("startup_time", time.Since(start)))
				return nil
			}
			if time.Since(start) > time.Duration(timeout)*time.Second {
				return fmt.Errorf("server failed to start within %ds", timeout)
			}
		}
	}
}

func (s *Server) waitForStop(ctx context.Context) error {
	maxWait := time.Duration(s.cfg.Server.MaxStopWait) * time.Second
	ctx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server failed to stop within %ds", s.cfg.Server.MaxStopWait)
		case <-ticker.C:
			status, err := s.Status(ctx)
			if err != nil {
				return err
			}
			if !status.IsRunning {
				s.logger.Info("Server stopped", zap.Duration("wait_time", time.Since(start)))
				return nil
			}
		}
	}
}

func (s *Server) checkServerDirectory() domain.HealthCheck {
	if info, err := os.Stat(s.cfg.Paths.Server); err == nil && info.IsDir() {
		return domain.HealthCheck{Name: "Server directory", Status: domain.StatusOK, Message: "OK"}
	}
	return domain.HealthCheck{Name: "Server directory", Status: domain.StatusError, Message: "Directory not found"}
}

func (s *Server) checkServerJAR() domain.HealthCheck {
	serverJar := filepath.Join(s.cfg.Paths.Server, s.cfg.Server.JarName)
	if info, err := os.Stat(serverJar); err == nil && !info.IsDir() {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		return domain.HealthCheck{
			Name:    "Server JAR",
			Status:  domain.StatusOK,
			Message: fmt.Sprintf("Found (%.1f MB)", sizeMB),
		}
	}
	return domain.HealthCheck{
		Name:    "Server JAR",
		Status:  domain.StatusError,
		Message: fmt.Sprintf("Not found: %s", s.cfg.Server.JarName),
	}
}

func (s *Server) checkJava(ctx context.Context) domain.HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "java", "-version")
	if err := cmd.Run(); err == nil {
		return domain.HealthCheck{Name: "Java Runtime", Status: domain.StatusOK, Message: "Available"}
	}
	return domain.HealthCheck{Name: "Java Runtime", Status: domain.StatusError, Message: "Java not found"}
}

func (s *Server) checkScreen(ctx context.Context) domain.HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "screen", "-version")
	if err := cmd.Run(); err == nil {
		return domain.HealthCheck{Name: "GNU screen", Status: domain.StatusOK, Message: "Available"}
	}
	return domain.HealthCheck{Name: "GNU screen", Status: domain.StatusError, Message: "screen not found"}
}
