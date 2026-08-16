// Package service implements business logic for server, mods, backups, and notifications.
package service

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/domain"
)

const (
	backupTimeFormat = "20060102_150405.000000000"
	backupPrefix     = "minecraft_backup_"
	backupExt        = ".tar.gz"
)

// Backup manages compressed server archives with retention.
type Backup struct {
	cfg    *config.Config
	logger *zap.Logger
}

// NewBackup creates a backup manager.
func NewBackup(cfg *config.Config, logger *zap.Logger) *Backup {
	return &Backup{cfg: cfg, logger: logger}
}

// Create generates a compressed tarball of the server directory.
func (b *Backup) Create(ctx context.Context) (string, error) {
	if !b.cfg.Backup.Enabled {
		b.logger.Info("Backups are disabled")
		return "", domain.ErrBackupsDisabled
	}

	if b.cfg.DryRun {
		b.logger.Info("Dry run: Would create backup")
		return "", nil
	}

	if check := domain.CheckPath("Server", b.cfg.Paths.Server); check.Status != domain.StatusOK {
		return "", fmt.Errorf("%s: %s", check.Name, check.Message)
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

// List returns metadata for all backup archives, newest first.
func (b *Backup) List() ([]domain.BackupInfo, error) {
	files, err := os.ReadDir(b.cfg.Paths.Backups)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	backups := make([]domain.BackupInfo, 0, len(files))
	for _, entry := range files {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), backupExt) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, domain.BackupInfo{
			Name:      entry.Name(),
			Path:      filepath.Join(b.cfg.Paths.Backups, entry.Name()),
			CreatedAt: info.ModTime(),
			Size:      info.Size(),
		})
	}

	slices.SortFunc(backups, func(a, b domain.BackupInfo) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	return backups, nil
}

// Delete removes a backup archive by name.
func (b *Backup) Delete(name string) error {
	backups, err := b.List()
	if err != nil {
		return err
	}
	for _, bk := range backups {
		if bk.Name != name {
			continue
		}
		if err := os.Remove(bk.Path); err != nil {
			return fmt.Errorf("failed to delete backup: %w", err)
		}
		b.logger.Info("Deleted backup", zap.String("name", name))
		return nil
	}
	return fmt.Errorf("backup not found: %s", name)
}

// HealthCheck verifies backup directory and retention settings.
func (b *Backup) HealthCheck(_ context.Context) []domain.HealthCheck {
	if !b.cfg.Backup.Enabled {
		return []domain.HealthCheck{{Name: "Backup system", Status: domain.StatusWarn, Message: "Disabled"}}
	}
	var retentionCheck domain.HealthCheck
	if b.cfg.Backup.MaxBackups <= 0 {
		retentionCheck = domain.HealthCheck{Name: "Backup retention", Status: domain.StatusOK, Message: "Unlimited"}
	} else {
		retentionCheck = domain.HealthCheck{Name: "Backup retention", Status: domain.StatusOK, Message: fmt.Sprintf("Keeping %d backups", b.cfg.Backup.MaxBackups)}
	}
	return []domain.HealthCheck{
		domain.CheckPath("Backup directory", b.cfg.Paths.Backups),
		retentionCheck,
	}
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

	gzLevel := b.cfg.Backup.CompressionLevel
	if gzLevel < gzip.NoCompression || gzLevel > gzip.BestCompression {
		gzLevel = gzip.DefaultCompression
	}

	gzWriter, err := gzip.NewWriterLevel(file, gzLevel)
	if err != nil {
		return "", err
	}
	tarWriter := tar.NewWriter(gzWriter)

	if err := b.addFiles(ctx, tarWriter); err != nil {
		_ = tarWriter.Close()
		_ = gzWriter.Close()
		_ = file.Close()
		_ = os.Remove(backupPath)
		return "", err
	}

	if err := tarWriter.Close(); err != nil {
		_ = gzWriter.Close()
		_ = file.Close()
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("finalizing tar: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		_ = file.Close()
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("finalizing gzip: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("closing backup file: %w", err)
	}

	info, err := os.Stat(backupPath)
	if err != nil || info.Size() == 0 {
		_ = os.Remove(backupPath)
		return "", errors.New("backup file empty or not created")
	}

	b.logger.Info("Backup created", zap.String("name", backupName), zap.Int64("size", info.Size()))
	return backupPath, nil
}

func (b *Backup) addFiles(ctx context.Context, tw *tar.Writer) error {
	root, err := os.OpenRoot(b.cfg.Paths.Server)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	return fs.WalkDir(root.FS(), ".", func(relPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if b.shouldExclude(relPath, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
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
		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := root.Open(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, f)
		_ = f.Close()
		return err
	})
}

// shouldExclude checks patterns using doublestar glob. Appends trailing slash
// for directories so patterns like "cache/" match correctly.
func (b *Backup) shouldExclude(relPath string, isDir bool) bool {
	if !b.cfg.Backup.IncludeLogs && (relPath == "logs" || strings.HasPrefix(relPath, "logs/")) {
		return true
	}
	matchPath := relPath
	if isDir && !strings.HasSuffix(matchPath, "/") {
		matchPath += "/"
	}
	for _, pattern := range b.cfg.Backup.ExcludePatterns {
		if matched, _ := doublestar.Match(pattern, matchPath); matched {
			return true
		}
		if isDir && matchPath != relPath {
			if matched, _ := doublestar.Match(pattern, relPath); matched {
				return true
			}
		}
	}
	return false
}

func (b *Backup) cleanup() {
	if b.cfg.Backup.MaxBackups <= 0 {
		return
	}
	backups, err := b.List()
	if err != nil {
		b.logger.Warn("Failed to list backups for cleanup", zap.Error(err))
		return
	}
	if len(backups) <= b.cfg.Backup.MaxBackups {
		return
	}
	for _, old := range backups[b.cfg.Backup.MaxBackups:] {
		if err := os.Remove(old.Path); err != nil {
			b.logger.Warn("Failed to remove old backup", zap.String("name", old.Name), zap.Error(err))
		} else {
			b.logger.Info("Removed old backup", zap.String("name", old.Name))
		}
	}
}
