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

	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/domain"
)

const (
	backupTimeFormat = "20060102_150405"
	backupPrefix     = "minecraft_backup_"
	backupExt        = ".tar.gz"
)

// Backup provides methods to create and manage server backups
type Backup struct {
	cfg    *config.Config
	logger *zap.Logger
}

var _ BackupManager = (*Backup)(nil)

// NewBackup initializes a new backup service
func NewBackup(cfg *config.Config, logger *zap.Logger) *Backup {
	return &Backup{cfg: cfg, logger: logger}
}

// Create generates a new compressed tarball of the server directory
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

// List scans the backup directory and returns metadata for all archives
func (b *Backup) List() ([]domain.BackupInfo, error) {
	files, err := os.ReadDir(b.cfg.Paths.Backups)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	backups := make([]domain.BackupInfo, 0)
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

	// Sort backups by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// HealthCheck verifies backup directory availability and retention settings
func (b *Backup) HealthCheck(_ context.Context) []domain.HealthCheck {
	if !b.cfg.Backup.Enabled {
		return []domain.HealthCheck{{Name: "Backup system", Status: domain.StatusWarn, Message: "Disabled"}}
	}
	return []domain.HealthCheck{
		domain.CheckPath("Backup directory", b.cfg.Paths.Backups),
		b.checkRetention(),
	}
}

func (b *Backup) validateServerDir() error {
	check := domain.CheckPath("Server", b.cfg.Paths.Server)
	if check.Status != domain.StatusOK {
		return fmt.Errorf("%s: %s", check.Name, check.Message)
	}
	return nil
}

// createArchive performs the actual file walking and compression
func (b *Backup) createArchive(ctx context.Context) (string, error) {
	timestamp := time.Now().Format(backupTimeFormat)
	backupName := backupPrefix + timestamp + backupExt
	backupPath := filepath.Join(b.cfg.Paths.Backups, backupName)

	b.logger.Info("Creating backup", zap.String("name", backupName))

	file, err := os.Create(backupPath) //nolint:gosec // backup path is from config
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			b.logger.Warn("Failed to close backup file", zap.Error(closeErr))
		}
	}()

	// Ensure compression level is within valid gzip range
	gzLevel := b.cfg.Backup.CompressionLevel
	if gzLevel < gzip.NoCompression || gzLevel > gzip.BestCompression {
		gzLevel = gzip.DefaultCompression
	}

	gzWriter, err := gzip.NewWriterLevel(file, gzLevel)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := gzWriter.Close(); closeErr != nil {
			b.logger.Warn("Failed to close gzip writer", zap.Error(closeErr))
		}
	}()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		if closeErr := tarWriter.Close(); closeErr != nil {
			b.logger.Warn("Failed to close tar writer", zap.Error(closeErr))
		}
	}()

	if err := b.addFiles(ctx, tarWriter); err != nil {
		_ = os.Remove(backupPath)
		return "", err
	}

	// Verify file was created and isn't empty
	info, err := os.Stat(backupPath)
	if err != nil || info.Size() == 0 {
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("backup file empty or not created")
	}

	b.logger.Info("Backup created", zap.String("name", backupName), zap.Int64("size", info.Size()))
	return backupPath, nil
}

// addFiles walks the server directory and adds eligible files to the archive
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
		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path) //nolint:gosec // path is from server directory walk
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close() // Close errors are non-critical after successful read
		}()
		_, err = io.Copy(tw, f)
		return err
	})
}

// shouldExclude checks if a file/dir should be skipped based on config patterns using doublestar for glob support
func (b *Backup) shouldExclude(relPath string, _ os.FileInfo) bool {
	if !b.cfg.Backup.IncludeLogs && (relPath == "logs" || strings.HasPrefix(relPath, "logs/")) {
		return true
	}
	for _, pattern := range b.cfg.Backup.ExcludePatterns {
		if matched, _ := doublestar.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}

// cleanup enforces the max_backups retention policy
func (b *Backup) cleanup() {
	backups, err := b.List()
	if err != nil || len(backups) <= b.cfg.Backup.MaxBackups {
		return
	}

	// List is sorted newest first, so we delete from index max_backups onwards
	for _, old := range backups[b.cfg.Backup.MaxBackups:] {
		if err := os.Remove(old.Path); err == nil {
			b.logger.Info("Removed old backup", zap.String("name", old.Name))
		}
	}
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
