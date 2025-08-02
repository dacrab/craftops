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

type BackupService struct {
	config *config.Config
	logger *zap.Logger
}

type BackupInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
}

func NewBackupService(cfg *config.Config, logger *zap.Logger) *BackupService {
	return &BackupService{
		config: cfg,
		logger: logger,
	}
}

func (bs *BackupService) CreateBackup(ctx context.Context) (string, error) {
	if !bs.config.Backup.Enabled {
		bs.logger.Info("Backups are disabled")
		return "", nil
	}

	if bs.config.DryRun {
		bs.logger.Info("Dry run: Would create backup")
		return "dry-run-backup.tar.gz", nil
	}

	if err := bs.validateServerDirectory(); err != nil {
		return "", err
	}

	if err := bs.ensureBackupDir(); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath, err := bs.createBackupFile()
	if err != nil {
		return "", err
	}
	bs.cleanupOldBackups()
	return backupPath, nil
}

func (bs *BackupService) ListBackups() ([]BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(bs.config.Paths.Backups, "*.tar.gz"))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	bs.sortBackupsByModTime(files)

	backups := make([]BackupInfo, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Name:      filepath.Base(file),
			Path:      file,
			CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			Size:      fmt.Sprintf("%.1f MB", float64(info.Size())/(1024*1024)),
			SizeBytes: info.Size(),
		})
	}

	return backups, nil
}

func (bs *BackupService) validateServerDirectory() error {
	if _, err := os.Stat(bs.config.Paths.Server); os.IsNotExist(err) {
		return fmt.Errorf("server directory not found: %s", bs.config.Paths.Server)
	}
	return nil
}

func (bs *BackupService) ensureBackupDir() error {
	return os.MkdirAll(bs.config.Paths.Backups, 0755)
}

func (bs *BackupService) createBackupFile() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("minecraft_backup_%s.tar.gz", timestamp)
	backupPath := filepath.Join(bs.config.Paths.Backups, backupName)

	bs.logger.Info("Creating backup", zap.String("backup_name", backupName))

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	gzipWriter, err := gzip.NewWriterLevel(backupFile, bs.config.Backup.CompressionLevel)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if err := bs.addFilesToBackup(tarWriter); err != nil {
		os.Remove(backupPath)
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	if err := bs.verifyBackup(backupPath); err != nil {
		os.Remove(backupPath)
		return "", err
	}

	bs.logBackupSuccess(backupName, backupPath)
	return backupPath, nil
}

func (bs *BackupService) addFilesToBackup(tarWriter *tar.Writer) error {
	return filepath.Walk(bs.config.Paths.Server, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(bs.config.Paths.Server, path)
		if err != nil {
			return err
		}

		if bs.shouldExclude(relPath, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			return bs.copyFileToTar(path, tarWriter)
		}

		return nil
	})
}

func (bs *BackupService) copyFileToTar(path string, tarWriter *tar.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tarWriter, file)
	return err
}

func (bs *BackupService) verifyBackup(backupPath string) error {
	info, err := os.Stat(backupPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("backup file was not created or is empty")
	}
	return nil
}

func (bs *BackupService) logBackupSuccess(backupName, backupPath string) {
	if info, err := os.Stat(backupPath); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		bs.logger.Info("Backup created successfully",
			zap.String("backup_name", backupName),
			zap.Float64("size_mb", sizeMB))
	}
}

func (bs *BackupService) shouldExclude(relPath string, _ os.FileInfo) bool {
	if !bs.config.Backup.IncludeLogs && strings.Contains(relPath, "logs") {
		return true
	}

	for _, pattern := range bs.config.Backup.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
		if strings.Contains(relPath, strings.TrimSuffix(pattern, "/")) {
			return true
		}
	}
	return false
}

func (bs *BackupService) sortBackupsByModTime(files []string) {
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})
}

func (bs *BackupService) cleanupOldBackups() {
	files, err := filepath.Glob(filepath.Join(bs.config.Paths.Backups, "*.tar.gz"))
	if err != nil {
		bs.logger.Error("Failed to list backup files for cleanup", zap.Error(err))
		return
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})
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

// HealthCheck performs health checks for the backup service
func (bs *BackupService) HealthCheck(ctx context.Context) []HealthCheck {
	checks := make([]HealthCheck, 0, 2)

	// Check backup directory
	checks = append(checks, bs.checkBackupDirectory())
	// Check backup configuration
	checks = append(checks, bs.checkBackupConfig())

	return checks
}

func (bs *BackupService) checkBackupDirectory() HealthCheck {
	if !bs.config.Backup.Enabled {
		return HealthCheck{
			Name:    "Backup system",
			Status:  "⚠️",
			Message: "Disabled in configuration",
		}
	}

	if info, err := os.Stat(bs.config.Paths.Backups); err == nil && info.IsDir() {
		// Test write permissions
		testFile := filepath.Join(bs.config.Paths.Backups, ".health_check_test")
		if file, err := os.Create(testFile); err == nil {
			file.Close()
			os.Remove(testFile)
			return HealthCheck{
				Name:    "Backup directory",
				Status:  "✅",
				Message: "OK",
			}
		} else {
			return HealthCheck{
				Name:    "Backup directory",
				Status:  "❌",
				Message: "No write permission",
			}
		}
	}
	return HealthCheck{
		Name:    "Backup directory",
		Status:  "❌",
		Message: "Directory not found",
	}
}

func (bs *BackupService) checkBackupConfig() HealthCheck {
	if bs.config.Backup.MaxBackups <= 0 {
		return HealthCheck{
			Name:    "Backup retention",
			Status:  "⚠️",
			Message: "Invalid max_backups setting",
		}
	}
	return HealthCheck{
		Name:    "Backup retention",
		Status:  "✅",
		Message: fmt.Sprintf("Keeping %d backups", bs.config.Backup.MaxBackups),
	}
}
