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

// Server manages the lifecycle and health of the Minecraft server process
type Server struct {
	cfg    *config.Config
	logger *zap.Logger
}

var _ ServerManager = (*Server)(nil)

// NewServer initializes a new server management service
func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{cfg: cfg, logger: logger}
}

// Status checks if the server is running by listing GNU screen sessions
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

// Start launches the server in a detached screen session
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

// Stop sends the stop command to the server session and waits for exit
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

// Restart performs a sequential stop and start of the server
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

// HealthCheck verifies all server dependencies (Java, Screen, paths)
func (s *Server) HealthCheck(ctx context.Context) []domain.HealthCheck {
	return []domain.HealthCheck{
		domain.CheckPath("Server directory", s.cfg.Paths.Server),
		s.checkServerJAR(),
		s.checkBinary(ctx, "java", "Java Runtime"),
		s.checkBinary(ctx, "screen", "GNU screen"),
	}
}

func (s *Server) checkBinary(ctx context.Context, binary, name string) domain.HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, binary, "-version").Run(); err == nil {
		return domain.HealthCheck{Name: name, Status: domain.StatusOK, Message: "Available"}
	}
	return domain.HealthCheck{Name: name, Status: domain.StatusError, Message: binary + " not found"}
}

func (s *Server) checkServerJAR() domain.HealthCheck {
	serverJar := filepath.Join(s.cfg.Paths.Server, s.cfg.Server.JarName)
	if info, err := os.Stat(serverJar); err == nil && !info.IsDir() {
		return domain.HealthCheck{
			Name:    "Server JAR",
			Status:  domain.StatusOK,
			Message: fmt.Sprintf("Found (%.1f MB)", float64(info.Size())/(1024*1024)),
		}
	}
	return domain.HealthCheck{Name: "Server JAR", Status: domain.StatusError, Message: "Not found"}
}

func (s *Server) sessionName() string {
	if s.cfg.Server.SessionName != "" {
		return s.cfg.Server.SessionName
	}
	return "minecraft"
}

// waitForStatus polls the server status until target is reached or timeout occurs
func (s *Server) waitForStatus(ctx context.Context, target bool, timeout int, label string) error {
	if timeout <= 0 {
		timeout = 30
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
			if status.IsRunning == target {
				s.logger.Info(fmt.Sprintf("Server %s", label), zap.Duration("duration", time.Since(start)))
				return nil
			}
			if time.Since(start) > time.Duration(timeout)*time.Second {
				return fmt.Errorf("server failed to %s within %ds", label, timeout)
			}
		}
	}
}

func (s *Server) waitForStart(ctx context.Context) error {
	return s.waitForStatus(ctx, true, s.cfg.Server.StartupTimeout, "started")
}

func (s *Server) waitForStop(ctx context.Context) error {
	return s.waitForStatus(ctx, false, s.cfg.Server.MaxStopWait, "stopped")
}
