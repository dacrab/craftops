package service

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
	"craftops/internal/domain"
)

const (
	backupTimeFormat = "20060102_150405"
	backupPrefix     = "minecraft_backup_"
	backupExt        = ".tar.gz"
)

// Backup implements BackupManager
type Backup struct {
	cfg    *config.Config
	logger *zap.Logger
}

var _ BackupManager = (*Backup)(nil)

// NewBackup creates a new backup service
func NewBackup(cfg *config.Config, logger *zap.Logger) *Backup {
	return &Backup{cfg: cfg, logger: logger}
}

// Create creates a new backup
func (b *Backup) Create(ctx context.Context) (string, error) {
	if !b.cfg.Backup.Enabled {
		b.logger.Info("Backups are disabled")
		return "", domain.ErrBackupsDisabled
	}

	if b.cfg.DryRun {
		b.logger.Info("Dry run: Would create backup")
		return "dry-run-backup.tar.gz", nil
	}

	if err := b.validateServerDir(); err != nil {
		return "", err
	}

	if err := os.MkdirAll(b.cfg.Paths.Backups, 0o750); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupPath, err := b.createArchive(ctx)
	if err != nil {
		return "", err
	}

	b.cleanup()
	return backupPath, nil
}

// List returns all available backups
func (b *Backup) List() ([]domain.BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(b.cfg.Paths.Backups, "*"+backupExt))
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	backups := make([]domain.BackupInfo, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		backups = append(backups, domain.BackupInfo{
			Name:      filepath.Base(file),
			Path:      file,
			CreatedAt: info.ModTime(),
			Size:      info.Size(),
		})
	}

	return backups, nil
}

// HealthCheck performs health checks
func (b *Backup) HealthCheck(_ context.Context) []domain.HealthCheck {
	return []domain.HealthCheck{
		b.checkBackupDir(),
		b.checkRetention(),
	}
}

func (b *Backup) validateServerDir() error {
	info, err := os.Stat(b.cfg.Paths.Server)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("server directory not found: %s", b.cfg.Paths.Server)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("server path is not a directory: %s", b.cfg.Paths.Server)
	}
	return nil
}

func (b *Backup) createArchive(ctx context.Context) (string, error) {
	timestamp := time.Now().Format(backupTimeFormat)
	backupName := backupPrefix + timestamp + backupExt
	backupPath := filepath.Join(b.cfg.Paths.Backups, backupName)

	b.logger.Info("Creating backup", zap.String("name", backupName))

	file, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzLevel := b.cfg.Backup.CompressionLevel
	if gzLevel < gzip.NoCompression {
		gzLevel = gzip.DefaultCompression
	}
	if gzLevel > gzip.BestCompression {
		gzLevel = gzip.BestCompression
	}

	gzWriter, err := gzip.NewWriterLevel(file, gzLevel)
	if err != nil {
		return "", err
	}
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	if err := b.addFiles(ctx, tarWriter); err != nil {
		os.Remove(backupPath)
		return "", err
	}

	info, err := os.Stat(backupPath)
	if err != nil || info.Size() == 0 {
		os.Remove(backupPath)
		return "", fmt.Errorf("backup file empty or not created")
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	b.logger.Info("Backup created", zap.String("name", backupName), zap.Float64("size_mb", sizeMB))

	return backupPath, nil
}

func (b *Backup) addFiles(ctx context.Context, tw *tar.Writer) error {
	return filepath.Walk(b.cfg.Paths.Server, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		relPath, err := filepath.Rel(b.cfg.Paths.Server, path)
		if err != nil {
			return err
		}

		if b.shouldExclude(relPath, info) {
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

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(tw, f)
			return err
		}

		return nil
	})
}

func (b *Backup) shouldExclude(relPath string, _ os.FileInfo) bool {
	if !b.cfg.Backup.IncludeLogs {
		if strings.HasPrefix(relPath, "logs/") || relPath == "logs" {
			return true
		}
	}

	for _, pattern := range b.cfg.Backup.ExcludePatterns {
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

func (b *Backup) cleanup() {
	files, err := filepath.Glob(filepath.Join(b.cfg.Paths.Backups, "*"+backupExt))
	if err != nil {
		b.logger.Error("Failed to list backups for cleanup", zap.Error(err))
		return
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	if len(files) > b.cfg.Backup.MaxBackups {
		for _, file := range files[:len(files)-b.cfg.Backup.MaxBackups] {
			if err := os.Remove(file); err != nil {
				b.logger.Error("Failed to remove old backup", zap.String("file", file), zap.Error(err))
			} else {
				b.logger.Info("Removed old backup", zap.String("name", filepath.Base(file)))
			}
		}
	}
}

func (b *Backup) checkBackupDir() domain.HealthCheck {
	if !b.cfg.Backup.Enabled {
		return domain.HealthCheck{Name: "Backup system", Status: domain.StatusWarn, Message: "Disabled"}
	}

	info, err := os.Stat(b.cfg.Paths.Backups)
	if err != nil || !info.IsDir() {
		return domain.HealthCheck{Name: "Backup directory", Status: domain.StatusError, Message: "Not found"}
	}

	testFile := filepath.Join(b.cfg.Paths.Backups, ".health_test")
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
		return domain.HealthCheck{Name: "Backup directory", Status: domain.StatusOK, Message: "OK"}
	}

	return domain.HealthCheck{Name: "Backup directory", Status: domain.StatusError, Message: "No write permission"}
}

func (b *Backup) checkRetention() domain.HealthCheck {
	if b.cfg.Backup.MaxBackups <= 0 {
		return domain.HealthCheck{Name: "Backup retention", Status: domain.StatusWarn, Message: "Invalid max_backups"}
	}
	return domain.HealthCheck{
		Name:    "Backup retention",
		Status:  domain.StatusOK,
		Message: fmt.Sprintf("Keeping %d backups", b.cfg.Backup.MaxBackups),
	}
}
