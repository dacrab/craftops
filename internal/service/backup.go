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
		if errors.Is(err, os.ErrNotExist) {
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
	slices.SortFunc(backups, func(a, b domain.BackupInfo) int {
		return b.CreatedAt.Compare(a.CreatedAt)
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

	// Ensure compression level is within valid gzip range
	gzLevel := b.cfg.Backup.CompressionLevel
	if gzLevel < gzip.NoCompression || gzLevel > gzip.BestCompression {
		gzLevel = gzip.DefaultCompression
	}

	gzWriter, err := gzip.NewWriterLevel(file, gzLevel)
	if err != nil {
		return "", err
	}
	tarWriter := tar.NewWriter(gzWriter)

	// Walk and archive — if anything fails, remove the partial file.
	if err := b.addFiles(ctx, tarWriter); err != nil {
		_ = tarWriter.Close()
		_ = gzWriter.Close()
		_ = file.Close()
		_ = os.Remove(backupPath)
		return "", err
	}

	// Close in reverse order: tar → gzip → file so each layer flushes correctly.
	// All close errors are propagated — a failed close means a corrupt archive.
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

	// Verify file was created and isn't empty
	info, err := os.Stat(backupPath)
	if err != nil || info.Size() == 0 {
		_ = os.Remove(backupPath)
		return "", errors.New("backup file empty or not created")
	}

	b.logger.Info("Backup created", zap.String("name", backupName), zap.Int64("size", info.Size()))
	return backupPath, nil
}

// addFiles walks the server directory and adds eligible files to the archive
func (b *Backup) addFiles(ctx context.Context, tw *tar.Writer) error {
	return filepath.WalkDir(b.cfg.Paths.Server, func(path string, d fs.DirEntry, err error) error {
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

		// Skip symlinks
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

		f, err := os.Open(path) //nolint:gosec // path is from server directory walk
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = io.Copy(tw, f)
		return err
	})
}

// shouldExclude checks if a file/dir should be skipped based on config patterns using doublestar for glob support
func (b *Backup) shouldExclude(relPath string, isDir bool) bool {
	if !b.cfg.Backup.IncludeLogs && (relPath == "logs" || strings.HasPrefix(relPath, "logs/")) {
		return true
	}
	// Append trailing slash for directories so patterns like "cache/" match correctly
	matchPath := relPath
	if isDir && !strings.HasSuffix(matchPath, "/") {
		matchPath += "/"
	}
	for _, pattern := range b.cfg.Backup.ExcludePatterns {
		if matched, _ := doublestar.Match(pattern, matchPath); matched {
			return true
		}
		if matched, _ := doublestar.Match(pattern, relPath); matched {
			return true
		}
	}
	return false
}

// cleanup enforces the max_backups retention policy
func (b *Backup) cleanup() {
	backups, err := b.List()
	if err != nil {
		b.logger.Warn("Failed to list backups for cleanup", zap.Error(err))
		return
	}
	if len(backups) <= b.cfg.Backup.MaxBackups {
		return
	}
	// List is sorted newest first, so delete from index max_backups onwards
	for _, old := range backups[b.cfg.Backup.MaxBackups:] {
		if err := os.Remove(old.Path); err != nil {
			b.logger.Warn("Failed to remove old backup", zap.String("name", old.Name), zap.Error(err))
		} else {
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
