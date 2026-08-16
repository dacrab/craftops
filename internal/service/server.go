package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/domain"
)

// Server manages the Minecraft server process lifecycle.
type Server struct {
	cfg    *config.Config
	logger *zap.Logger
}

// NewServer creates a server manager.
func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{cfg: cfg, logger: logger}
}

// Status checks if the server screen session is running.
func (s *Server) Status(ctx context.Context) (*domain.ServerStatus, error) {
	session := s.sessionName()
	cmd := exec.CommandContext(ctx, "screen", "-ls", session)
	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("server.status: screen not found in PATH: %w", err)
		}
		// screen exits non-zero when no sockets match; that means "not running".
		s.logger.Debug("screen -ls returned no matching sessions", zap.Error(err))
	}

	return &domain.ServerStatus{
		IsRunning:   sessionRunning(string(output), session),
		SessionName: session,
		CheckedAt:   time.Now(),
	}, nil
}

// sessionRunning reports whether output contains a "pid.<name>" screen
// listing for the exact session name, avoiding prefix false-positives.
func sessionRunning(output, session string) bool {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		dot := strings.LastIndex(name, ".")
		if dot <= 0 || dot == len(name)-1 {
			continue
		}
		if _, err := strconv.Atoi(name[:dot]); err != nil {
			continue
		}
		if name[dot+1:] == session {
			return true
		}
	}
	return false
}

// Start launches the server in a detached screen session.
func (s *Server) Start(ctx context.Context) error {
	if s.cfg.DryRun {
		s.logger.Info("Dry run: Would start server")
		return nil
	}

	status, err := s.Status(ctx)
	if err != nil {
		return fmt.Errorf("server.start: %w", err)
	}
	if status.IsRunning {
		s.logger.Warn("Server is already running")
		return nil
	}

	serverJar := filepath.Join(s.cfg.Paths.Server, s.cfg.Server.JarName)
	if _, err := os.Stat(serverJar); errors.Is(err, os.ErrNotExist) {
		return domain.ErrServerJarNotFound
	}

	javaArgs := slices.Concat(s.cfg.Server.JavaFlags, []string{"-jar", s.cfg.Server.JarName, "nogui"})
	cmdArgs := append([]string{"-dmS", s.sessionName(), "java"}, javaArgs...)

	cmd := exec.CommandContext(ctx, "screen", cmdArgs...)
	cmd.Dir = s.cfg.Paths.Server
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("server.start: %w", err)
	}

	return s.waitForStatus(ctx, true, s.cfg.Server.StartupTimeout, "started")
}

// Stop sends the stop command and waits for exit.
func (s *Server) Stop(ctx context.Context) error {
	if s.cfg.DryRun {
		s.logger.Info("Dry run: Would stop server")
		return nil
	}

	status, err := s.Status(ctx)
	if err != nil {
		return fmt.Errorf("server.stop: %w", err)
	}
	if !status.IsRunning {
		s.logger.Warn("Server is not running")
		return nil
	}

	stopCmd := s.cfg.Server.StopCommand + "\n"
	cmd := exec.CommandContext(ctx, "screen", "-S", s.sessionName(), "-X", "stuff", stopCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("server.stop: %w", err)
	}

	return s.waitForStatus(ctx, false, s.cfg.Server.MaxStopWait, "stopped")
}

// Restart performs a sequential stop then start.
func (s *Server) Restart(ctx context.Context) error {
	s.logger.Info("Restarting server")
	if err := s.Stop(ctx); err != nil {
		return err
	}
	return s.Start(ctx)
}

// HealthCheck verifies server dependencies (Java, screen, paths).
func (s *Server) HealthCheck(_ context.Context) []domain.HealthCheck {
	checks := []domain.HealthCheck{
		domain.CheckPath("Server directory", s.cfg.Paths.Server),
	}

	serverJar := filepath.Join(s.cfg.Paths.Server, s.cfg.Server.JarName)
	if info, err := os.Stat(serverJar); err == nil && !info.IsDir() {
		checks = append(checks, domain.HealthCheck{
			Name:    "Server JAR",
			Status:  domain.StatusOK,
			Message: fmt.Sprintf("Found (%.1f MB)", float64(info.Size())/(1024*1024)),
		})
	} else {
		checks = append(checks, domain.HealthCheck{Name: "Server JAR", Status: domain.StatusError, Message: "Not found"})
	}

	for _, b := range []struct{ bin, name string }{{"java", "Java Runtime"}, {"screen", "GNU screen"}} {
		if _, err := exec.LookPath(b.bin); err == nil {
			checks = append(checks, domain.HealthCheck{Name: b.name, Status: domain.StatusOK, Message: "Available"})
		} else {
			checks = append(checks, domain.HealthCheck{Name: b.name, Status: domain.StatusError, Message: b.bin + " not found in PATH"})
		}
	}
	return checks
}

func (s *Server) sessionName() string {
	if s.cfg.Server.SessionName != "" {
		return s.cfg.Server.SessionName
	}
	return "minecraft"
}

// waitForStatus polls until the server reaches the target state or timeout.
func (s *Server) waitForStatus(ctx context.Context, target bool, timeout int, label string) error {
	if timeout <= 0 {
		timeout = 30
	}

	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
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
				s.logger.Info("Server "+label, zap.Duration("duration", time.Since(start)))
				return nil
			}
			if time.Since(start) > time.Duration(timeout)*time.Second {
				return fmt.Errorf("server failed to %s within %ds", label, timeout)
			}
		}
	}
}
