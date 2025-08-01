package services

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// BackupService handles backup operations
type BackupService struct {
	config *config.Config
	logger *zap.Logger
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
}

// NewBackupService creates a new backup service instance
func NewBackupService(cfg *config.Config, logger *zap.Logger) *BackupService {
	return &BackupService{
		config: cfg,
		logger: logger,
	}
}

// HealthCheck performs health checks for the backup service
func (bs *BackupService) HealthCheck(ctx context.Context) []HealthCheck {
	checks := []HealthCheck{}

	// Check backup directory
	backupDir := bs.config.Paths.Backups
	if info, err := os.Stat(backupDir); err == nil && info.IsDir() {
		// Count backup files
		backupCount := 0
		if files, err := filepath.Glob(filepath.Join(backupDir, "*.tar.gz")); err == nil {
			backupCount = len(files)
		}
		checks = append(checks, HealthCheck{
			Name:    "Backup directory",
			Status:  "✅",
			Message: fmt.Sprintf("OK (%d backups found)", backupCount),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Backup directory",
			Status:  "❌",
			Message: "Directory not found",
		})
	}

	// Check backup configuration
	if bs.config.Backup.Enabled {
		checks = append(checks, HealthCheck{
			Name:    "Backup configuration",
			Status:  "✅",
			Message: fmt.Sprintf("Enabled (max: %d)", bs.config.Backup.MaxBackups),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Backup configuration",
			Status:  "⚠️",
			Message: "Disabled",
		})
	}

	// Check backup storage accessibility
	if err := bs.ensureBackupDir(); err != nil {
		checks = append(checks, HealthCheck{
			Name:    "Backup storage",
			Status:  "❌",
			Message: fmt.Sprintf("Error: %v", err),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Backup storage",
			Status:  "✅",
			Message: "Accessible",
		})
	}

	return checks
}

// CreateBackup creates a new backup of the server
func (bs *BackupService) CreateBackup(ctx context.Context) (string, error) {
	if !bs.config.Backup.Enabled {
		bs.logger.Info("Backups are disabled")
		return "", nil
	}

	if bs.config.DryRun {
		bs.logger.Info("Dry run: Would create backup")
		return "dry-run-backup.tar.gz", nil
	}

	serverDir := bs.config.Paths.Server
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		return "", fmt.Errorf("server directory not found: %s", serverDir)
	}

	// Ensure backup directory exists
	if err := bs.ensureBackupDir(); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("minecraft_backup_%s.tar.gz", timestamp)
	backupPath := filepath.Join(bs.config.Paths.Backups, backupName)

	bs.logger.Info("Creating backup", zap.String("backup_name", backupName))

	// Create the backup file
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	// Create gzip writer
	gzipWriter, err := gzip.NewWriterLevel(backupFile, bs.config.Backup.CompressionLevel)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through server directory and add files to backup
	err = filepath.Walk(serverDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(serverDir, path)
		if err != nil {
			return err
		}

		// Skip if should be excluded
		if bs.shouldExclude(relPath, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a regular file, write its content
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		os.Remove(backupPath) // Clean up on error
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Verify backup was created successfully
	if info, err := os.Stat(backupPath); err != nil || info.Size() == 0 {
		os.Remove(backupPath)
		return "", fmt.Errorf("backup file was not created or is empty")
	}

	// Get backup size
	info, _ := os.Stat(backupPath)
	sizeMB := float64(info.Size()) / (1024 * 1024)

	bs.logger.Info("Backup created successfully",
		zap.String("backup_name", backupName),
		zap.Float64("size_mb", sizeMB))

	// Clean up old backups
	bs.cleanupOldBackups()

	return backupPath, nil
}

// ListBackups lists all available backups
func (bs *BackupService) ListBackups() ([]BackupInfo, error) {
	backupDir := bs.config.Paths.Backups
	backups := []BackupInfo{}

	files, err := filepath.Glob(filepath.Join(backupDir, "*.tar.gz"))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		sizeMB := float64(info.Size()) / (1024 * 1024)
		createdAt := info.ModTime().Format("2006-01-02 15:04:05")

		backups = append(backups, BackupInfo{
			Name:      filepath.Base(file),
			Path:      file,
			CreatedAt: createdAt,
			Size:      fmt.Sprintf("%.1f MB", sizeMB),
			SizeBytes: info.Size(),
		})
	}

	return backups, nil
}

// ensureBackupDir ensures the backup directory exists
func (bs *BackupService) ensureBackupDir() error {
	return os.MkdirAll(bs.config.Paths.Backups, 0755)
}

// shouldExclude checks if a path should be excluded from backup
func (bs *BackupService) shouldExclude(relPath string, _ os.FileInfo) bool {
	// Skip logs if configured
	if !bs.config.Backup.IncludeLogs && strings.Contains(relPath, "logs") {
		return true
	}

	// Check exclude patterns
	for _, pattern := range bs.config.Backup.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}

		// Also check if any part of the path matches
		if strings.Contains(relPath, strings.TrimSuffix(pattern, "/")) {
			return true
		}
	}

	return false
}

// cleanupOldBackups removes old backups beyond the retention limit
func (bs *BackupService) cleanupOldBackups() {
	backupDir := bs.config.Paths.Backups

	files, err := filepath.Glob(filepath.Join(backupDir, "*.tar.gz"))
	if err != nil {
		bs.logger.Error("Failed to list backup files for cleanup", zap.Error(err))
		return
	}

	// Sort by modification time (oldest first for removal)
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Remove old backups beyond the limit
	if len(files) > bs.config.Backup.MaxBackups {
		for _, file := range files[:len(files)-bs.config.Backup.MaxBackups] {
			if err := os.Remove(file); err != nil {
				bs.logger.Error("Failed to remove old backup",
					zap.String("file", file),
					zap.Error(err))
			} else {
				bs.logger.Info("Removed old backup", zap.String("backup_name", filepath.Base(file)))
			}
		}
	}
}
