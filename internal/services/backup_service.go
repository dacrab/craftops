package services

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

type BackupService struct {
	*BaseService
}

type BackupInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
}

func NewBackupService(cfg *config.Config, logger *zap.Logger) BackupServiceInterface {
	return &BackupService{
		BaseService: NewBaseService(cfg, logger),
	}
}

func (bs *BackupService) CreateBackup(ctx context.Context) (string, error) {
	if !bs.GetConfig().Backup.Enabled {
		bs.GetLogger().Info("Backups are disabled")
		return "", nil
	}

	if bs.HandleDryRun("create backup") {
		return "dry-run-backup.tar.gz", nil
	}

	var backupPath string
	err := bs.LogOperation("backup creation", func() error {
		if err := bs.validateServerDirectory(); err != nil {
			return err
		}

		if err := bs.ensureBackupDir(); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}

		var err error
		backupPath, err = bs.createBackupFile()
		if err != nil {
			return err
		}

		bs.cleanupOldBackups()
		return nil
	})

	return backupPath, err
}

func (bs *BackupService) ListBackups() ([]BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(bs.GetConfig().Paths.Backups, "*.tar.gz"))
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

// RestoreBackup restores a backup archive into the server directory.
// If force is false and the server directory is non-empty, restoration is aborted.
func (bs *BackupService) RestoreBackup(ctx context.Context, backupPath string, force bool) error {
	if bs.HandleDryRun("restore backup", backupPath) {
		return nil
	}

	if err := bs.validateServerDirectory(); err != nil {
		return err
	}

	return bs.LogOperation("backup restoration", func() error {
		// Safety: refuse to restore into a non-empty server dir unless force
		isEmpty, err := isDirEmpty(bs.GetConfig().Paths.Server)
		if err != nil {
			return fmt.Errorf("failed to check server directory: %w", err)
		}
		if !isEmpty && !force {
			return fmt.Errorf("server directory is not empty; use force to overwrite")
		}

		return bs.extractBackupArchive(ctx, backupPath)
	})
}

// extractBackupArchive extracts a backup archive to the server directory
func (bs *BackupService) extractBackupArchive(ctx context.Context, backupPath string) error {
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		if err := bs.extractTarEntry(tarReader, header); err != nil {
			return err
		}
	}

	return nil
}

// extractTarEntry extracts a single tar entry
func (bs *BackupService) extractTarEntry(tarReader *tar.Reader, header *tar.Header) error {
	targetPath := filepath.Join(bs.GetConfig().Paths.Server, header.Name)

	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(targetPath, fs.FileMode(header.Mode))
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		return bs.writeFile(targetPath, tarReader, header.Mode)
	default:
		// skip other types
		return nil
	}
}

// writeFile writes a file from tar reader
func (bs *BackupService) writeFile(targetPath string, tarReader *tar.Reader, mode int64) error {
	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(mode))
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, tarReader)
	return err
}

func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	// read at most 1 entry
	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (bs *BackupService) validateServerDirectory() error {
	if _, err := os.Stat(bs.GetConfig().Paths.Server); os.IsNotExist(err) {
		return fmt.Errorf("server directory not found: %s", bs.GetConfig().Paths.Server)
	}
	return nil
}

func (bs *BackupService) ensureBackupDir() error {
	return os.MkdirAll(bs.GetConfig().Paths.Backups, 0755)
}

func (bs *BackupService) createBackupFile() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("minecraft_backup_%s.tar.gz", timestamp)
	backupPath := filepath.Join(bs.GetConfig().Paths.Backups, backupName)

	bs.GetLogger().Info("Creating backup", zap.String("backup_name", backupName))

	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	gzipWriter, err := gzip.NewWriterLevel(backupFile, bs.GetConfig().Backup.CompressionLevel)
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
	return filepath.Walk(bs.GetConfig().Paths.Server, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(bs.GetConfig().Paths.Server, path)
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
		bs.GetLogger().Info("Backup created successfully",
			zap.String("backup_name", backupName),
			zap.Float64("size_mb", sizeMB))
	}
}

func (bs *BackupService) shouldExclude(relPath string, _ os.FileInfo) bool {
	if !bs.GetConfig().Backup.IncludeLogs && strings.Contains(relPath, "logs") {
		return true
	}

	for _, pattern := range bs.GetConfig().Backup.ExcludePatterns {
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
	files, err := filepath.Glob(filepath.Join(bs.GetConfig().Paths.Backups, "*.tar.gz"))
	if err != nil {
		bs.GetLogger().Error("Failed to list backup files for cleanup", zap.Error(err))
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
	if len(files) > bs.GetConfig().Backup.MaxBackups {
		for _, file := range files[:len(files)-bs.GetConfig().Backup.MaxBackups] {
			if err := os.Remove(file); err != nil {
				bs.GetLogger().Error("Failed to remove old backup",
					zap.String("file", file),
					zap.Error(err))
			} else {
				bs.GetLogger().Info("Removed old backup", zap.String("backup_name", filepath.Base(file)))
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
	if !bs.GetConfig().Backup.Enabled {
		return bs.CreateHealthCheck("Backup system", "⚠️", "Disabled in configuration")
	}

	path := bs.GetConfig().Paths.Backups
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		// Test write permissions
		testFile := filepath.Join(path, ".health_check_test")
		if file, err := os.Create(testFile); err == nil {
			file.Close()
			os.Remove(testFile)
			return bs.CreateHealthCheck("Backup directory", "✅", "OK")
		}
		return bs.CreateHealthCheck("Backup directory", "❌", "No write permission")
	}
	return bs.CreateHealthCheck("Backup directory", "❌", "Directory not found")
}

func (bs *BackupService) checkBackupConfig() HealthCheck {
	if bs.GetConfig().Backup.MaxBackups <= 0 {
		return bs.CreateHealthCheck("Backup retention", "⚠️", "Invalid max_backups setting")
	}
	return bs.CreateHealthCheck("Backup retention", "✅",
		fmt.Sprintf("Keeping %d backups", bs.GetConfig().Backup.MaxBackups))
}
