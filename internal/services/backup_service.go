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

const (
	backupFileTimeFormat = "20060102_150405"
	humanTimeFormat      = "2006-01-02 15:04:05"
	backupFilePrefix     = "minecraft_backup_"
	backupFileExt        = ".tar.gz"
)

func clampGzipLevel(level int) int {
	if level < gzip.NoCompression {
		return gzip.DefaultCompression
	}
	if level > gzip.BestCompression {
		return gzip.BestCompression
	}
	return level
}

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

	backupPath, err := bs.createBackupFile(ctx)
	if err != nil {
		return "", err
	}
	bs.cleanupOldBackups()
	return backupPath, nil
}

func (bs *BackupService) ListBackups() ([]BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(bs.config.Paths.Backups, "*"+backupFileExt))
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
			CreatedAt: info.ModTime().Format(humanTimeFormat),
			Size:      fmt.Sprintf("%.1f MB", float64(info.Size())/(1024*1024)),
			SizeBytes: info.Size(),
		})
	}

	return backups, nil
}

func (bs *BackupService) validateServerDirectory() error {
	info, err := os.Stat(bs.config.Paths.Server)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("server directory not found: %s", bs.config.Paths.Server)
		}
		return fmt.Errorf("failed to stat server directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("server path is not a directory: %s", bs.config.Paths.Server)
	}
	return nil
}

func (bs *BackupService) ensureBackupDir() error {
    return os.MkdirAll(bs.config.Paths.Backups, 0o750)
}

func (bs *BackupService) createBackupFile(ctx context.Context) (string, error) {
	timestamp := time.Now().Format(backupFileTimeFormat)
	backupName := fmt.Sprintf("%s%s%s", backupFilePrefix, timestamp, backupFileExt)
	backupPath := filepath.Join(bs.config.Paths.Backups, backupName)

	bs.logger.Info("Creating backup", zap.String("backup_name", backupName))

    // #nosec G304 -- backup path is constructed within configured backups directory
    backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
    defer func() { _ = backupFile.Close() }()

    gzipWriter, err := gzip.NewWriterLevel(backupFile, clampGzipLevel(bs.config.Backup.CompressionLevel))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip writer: %w", err)
	}
    defer func() { _ = gzipWriter.Close() }()

    tarWriter := tar.NewWriter(gzipWriter)
    defer func() { _ = tarWriter.Close() }()

    if err := bs.addFilesToBackup(ctx, tarWriter); err != nil {
        _ = os.Remove(backupPath)
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

    if err := bs.verifyBackup(backupPath); err != nil {
        _ = os.Remove(backupPath)
		return "", err
	}

	bs.logBackupSuccess(backupName, backupPath)
	return backupPath, nil
}

func (bs *BackupService) addFilesToBackup(ctx context.Context, tarWriter *tar.Writer) error {
	return filepath.Walk(bs.config.Paths.Server, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
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

		if info.Mode()&os.ModeSymlink != 0 {
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
    // #nosec G304 -- reading from path within server directory
    file, err := os.Open(path)
	if err != nil {
		return err
	}
    defer func() { _ = file.Close() }()

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
	if !bs.config.Backup.IncludeLogs {
		if strings.HasPrefix(relPath, "logs/") || relPath == "logs" {
			return true
		}
	}

	for _, pattern := range bs.config.Backup.ExcludePatterns {
		p := pattern
		if strings.HasSuffix(p, "/") {
			p = strings.TrimSuffix(p, "/")
			if relPath == p || strings.HasPrefix(relPath, p+"/") {
				return true
			}
			continue
		}
		if matched, _ := filepath.Match(p, filepath.Base(relPath)); matched {
			return true
		}
		if matched, _ := filepath.Match(p, relPath); matched {
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
	files, err := filepath.Glob(filepath.Join(bs.config.Paths.Backups, "*"+backupFileExt))
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

func (bs *BackupService) HealthCheck(ctx context.Context) []HealthCheck {
	return []HealthCheck{
		bs.checkBackupDirectory(),
		bs.checkBackupConfig(),
	}
}

func (bs *BackupService) checkBackupDirectory() HealthCheck {
	if !bs.config.Backup.Enabled {
		return HealthCheck{
			Name:    "Backup system",
			Status:  "WARN",
			Message: "Disabled in configuration",
		}
	}

	info, err := os.Stat(bs.config.Paths.Backups)
	if err != nil || !info.IsDir() {
		return HealthCheck{
			Name:    "Backup directory",
			Status:  "ERROR",
			Message: "Directory not found",
		}
	}

	testFile := filepath.Join(bs.config.Paths.Backups, ".health_check_test")
	if file, err := os.Create(testFile); err == nil {
		file.Close()
		os.Remove(testFile)
		return HealthCheck{
			Name:    "Backup directory",
			Status:  "OK",
			Message: "OK",
		}
	}

	return HealthCheck{
		Name:    "Backup directory",
		Status:  "ERROR",
		Message: "No write permission",
	}
}

func (bs *BackupService) checkBackupConfig() HealthCheck {
	if bs.config.Backup.MaxBackups <= 0 {
		return HealthCheck{
			Name:    "Backup retention",
			Status:  "WARN",
			Message: "Invalid max_backups setting",
		}
	}
	return HealthCheck{
		Name:    "Backup retention",
		Status:  "OK",
		Message: fmt.Sprintf("Keeping %d backups", bs.config.Backup.MaxBackups),
	}
}
