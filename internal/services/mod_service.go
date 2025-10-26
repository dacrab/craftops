package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

// ModService handles mod management operations
type ModService struct {
	config *config.Config
	logger *zap.Logger
	client *http.Client
}

// ModUpdateResult represents the result of a mod update operation
type ModUpdateResult struct {
	UpdatedMods []string          `json:"updated_mods"`
	FailedMods  map[string]string `json:"failed_mods"`
	SkippedMods []string          `json:"skipped_mods"`
}

// ModInfo represents information about a mod
type ModInfo struct {
	VersionID   string `json:"version_id"`
	Version     string `json:"version_number"`
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ProjectName string `json:"project_name"`
}

// Minimal types for Modrinth API decoding
type mrFile struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

type mrVersion struct {
	ID            string   `json:"id"`
	VersionNumber string   `json:"version_number"`
	Files         []mrFile `json:"files"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewModService creates a new mod service instance
func NewModService(cfg *config.Config, logger *zap.Logger) *ModService {
	return &ModService{
		config: cfg,
		logger: logger,
		client: &http.Client{
			Timeout: time.Duration(cfg.Mods.Timeout) * time.Second,
		},
	}
}

// HealthCheck performs health checks for the mod service
func (ms *ModService) HealthCheck(ctx context.Context) []HealthCheck {
	return []HealthCheck{
		ms.checkModsDirectory(),
		ms.checkModSources(),
		ms.checkAPIConnectivity(ctx),
	}
}

// UpdateAllMods updates all configured mods to their latest versions
func (ms *ModService) UpdateAllMods(ctx context.Context, force bool) (*ModUpdateResult, error) {
	ms.logger.Info("Starting mod update process", zap.Bool("force", force))

	result := &ModUpdateResult{
		UpdatedMods: []string{},
		FailedMods:  make(map[string]string),
		SkippedMods: []string{},
	}

	if ms.config.DryRun {
		ms.logger.Info("Dry run mode - no actual changes will be made")
		result.UpdatedMods = []string{"example-mod (dry-run)"}
		return result, nil
	}

	sources := ms.config.Mods.ModrinthSources
	if len(sources) == 0 {
		ms.logger.Info("No mod sources configured")
		return result, nil
	}

	ms.processModsParallel(ctx, sources, force, result)

	ms.logger.Info("Mod update process completed",
		zap.Int("updated", len(result.UpdatedMods)),
		zap.Int("failed", len(result.FailedMods)),
		zap.Int("skipped", len(result.SkippedMods)))

	return result, nil
}

// ListInstalledMods lists all currently installed mods
func (ms *ModService) ListInstalledMods() ([]map[string]interface{}, error) {
	files, err := filepath.Glob(filepath.Join(ms.config.Paths.Mods, "*.jar"))
	if err != nil {
		return nil, fmt.Errorf("failed to list mod files: %w", err)
	}

	mods := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		filename := filepath.Base(file)
		name := strings.TrimSuffix(filename, filepath.Ext(filename))

		mods = append(mods, map[string]interface{}{
			"name":     name,
			"filename": filename,
			"size":     info.Size(),
			"modified": info.ModTime().Unix(),
		})
	}

	return mods, nil
}

// checkModsDirectory checks if the mods directory exists and counts JAR files
func (ms *ModService) checkModsDirectory() HealthCheck {
	modsDir := ms.config.Paths.Mods
	info, err := os.Stat(modsDir)
	if err != nil || !info.IsDir() {
		return HealthCheck{
			Name:    "Mods directory",
			Status:  "ERROR",
			Message: "Directory not found or not accessible",
		}
	}

	jarCount := 0
	if files, err := filepath.Glob(filepath.Join(modsDir, "*.jar")); err == nil {
		jarCount = len(files)
	}
	return HealthCheck{
		Name:    "Mods directory",
		Status:  "OK",
		Message: fmt.Sprintf("OK (%d mods found)", jarCount),
	}
}

// checkModSources checks the mod sources configuration
func (ms *ModService) checkModSources() HealthCheck {
	totalSources := len(ms.config.Mods.ModrinthSources)
	if totalSources == 0 {
		return HealthCheck{
			Name:    "Mod sources",
			Status:  "WARN",
			Message: "No mod sources configured",
		}
	}

	return HealthCheck{
		Name:    "Mod sources",
		Status:  "OK",
		Message: fmt.Sprintf("%d sources configured", totalSources),
	}
}

// checkAPIConnectivity tests connectivity to the Modrinth API
func (ms *ModService) checkAPIConnectivity(ctx context.Context) HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.modrinth.com/v2/", nil)
	if err != nil {
		return HealthCheck{
			Name:    "Modrinth API",
			Status:  "ERROR",
			Message: fmt.Sprintf("Request creation failed: %v", err),
		}
	}
	resp, err := ms.client.Do(req)
	if err != nil {
		return HealthCheck{
			Name:    "Modrinth API",
			Status:  "ERROR",
			Message: fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return HealthCheck{
			Name:    "Modrinth API",
			Status:  "WARN",
			Message: fmt.Sprintf("API returned status %d", resp.StatusCode),
		}
	}

	return HealthCheck{
		Name:    "Modrinth API",
		Status:  "OK",
		Message: "API accessible",
	}
}

// processModsParallel processes mods concurrently
func (ms *ModService) processModsParallel(ctx context.Context, sources []string, force bool, result *ModUpdateResult) {
	semaphore := make(chan struct{}, ms.config.Mods.ConcurrentDownloads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, source := range sources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			updated, name, err := ms.updateModrinthMod(ctx, src, force)
			display := name
			if display == "" {
				display = src
			}
			if err != nil {
				mu.Lock()
				result.FailedMods[display] = err.Error()
				mu.Unlock()
				ms.logger.Error("Failed to update mod", zap.String("mod", display), zap.Error(err))
				return
			}
			mu.Lock()
			if updated {
				result.UpdatedMods = append(result.UpdatedMods, display)
			} else {
				result.SkippedMods = append(result.SkippedMods, display)
			}
			mu.Unlock()
		}(source)
	}

	wg.Wait()
}

// updateModrinthMod updates a mod from Modrinth
func (ms *ModService) updateModrinthMod(ctx context.Context, modURL string, force bool) (bool, string, error) {
	projectID, err := ms.parseModrinthProjectID(modURL)
	if err != nil {
		return false, projectID, fmt.Errorf("failed to parse Modrinth project ID: %w", err)
	}

	versionInfo, err := ms.fetchModrinthLatestVersion(ctx, projectID)
	if err != nil {
		return false, projectID, fmt.Errorf("failed to fetch latest version: %w", err)
	}
	updated, err := ms.downloadMod(ctx, versionInfo.DownloadURL, versionInfo.Filename, force)
	return updated, versionInfo.ProjectName, err
}

// parseModrinthProjectID parses project ID from Modrinth URL
func (ms *ModService) parseModrinthProjectID(modURL string) (string, error) {
	u, err := url.Parse(modURL)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`/mod/([^/]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Modrinth URL format: %s", modURL)
	}

	return matches[1], nil
}

// fetchModrinthLatestVersion fetches the latest compatible version from Modrinth API
func (ms *ModService) fetchModrinthLatestVersion(ctx context.Context, projectID string) (*ModInfo, error) {
	url := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version?game_versions=[\"%s\"]&loaders=[\"%s\"]",
		projectID, ms.config.Minecraft.Version, ms.config.Minecraft.Modloader)

	var versions []mrVersion
	var lastErr error
	for attempt := 0; attempt <= ms.config.Mods.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "craftops/"+time.Now().Format("20060102"))

		resp, err := ms.client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			func() {
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode == 200 {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						lastErr = err
						return
					}
					if err := json.Unmarshal(body, &versions); err != nil {
						lastErr = err
						return
					}
					lastErr = nil
				} else if resp.StatusCode >= 500 || resp.StatusCode == 429 {
					lastErr = fmt.Errorf("API request failed with status %d", resp.StatusCode)
				} else {
					lastErr = fmt.Errorf("API request failed with status %d", resp.StatusCode)
					// non-retriable
					attempt = ms.config.Mods.MaxRetries
				}
			}()
		}

		if lastErr == nil {
			break
		}
		// backoff
		delay := time.Duration(ms.config.Mods.RetryDelay * float64(time.Second))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no compatible versions found")
	}

	v := versions[0]
	if len(v.Files) == 0 {
		return nil, fmt.Errorf("no files found in version")
	}
	f := v.Files[0]

	return &ModInfo{
		VersionID:   v.ID,
		Version:     v.VersionNumber,
		DownloadURL: f.URL,
		Filename:    f.Filename,
		ProjectName: projectID,
	}, nil
}

// downloadMod downloads a mod file
func (ms *ModService) downloadMod(ctx context.Context, downloadURL, filename string, force bool) (bool, error) {
	modsDir := ms.config.Paths.Mods

	if err := os.MkdirAll(modsDir, 0o750); err != nil {
		return false, fmt.Errorf("failed to create mods directory: %w", err)
	}

	finalPath := filepath.Join(modsDir, filename)
	if !force {
		if _, err := os.Stat(finalPath); err == nil {
			// Already have this file; skip
			ms.logger.Info("Mod already up-to-date, skipping", zap.String("filename", filename))
			return false, nil
		}
	}
	tmpFile, err := os.CreateTemp(modsDir, ".tmp-*")
	if err != nil {
		return false, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	var lastErr error
	for attempt := 0; attempt <= ms.config.Mods.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			return false, err
		}
		req.Header.Set("User-Agent", "craftops/"+time.Now().Format("20060102"))

		resp, err := ms.client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			func() {
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode == 200 {
					if _, err = io.Copy(tmpFile, resp.Body); err != nil {
						lastErr = fmt.Errorf("failed to write temp file: %w", err)
						return
					}
					lastErr = nil
				} else if resp.StatusCode >= 500 || resp.StatusCode == 429 {
					lastErr = fmt.Errorf("download failed with status %d", resp.StatusCode)
				} else {
					lastErr = fmt.Errorf("download failed with status %d", resp.StatusCode)
					// non-retriable
					attempt = ms.config.Mods.MaxRetries
				}
			}()
		}

		if lastErr == nil {
			break
		}
		// backoff
		delay := time.Duration(ms.config.Mods.RetryDelay * float64(time.Second))
		select {
		case <-ctx.Done():
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			return false, ctx.Err()
		case <-time.After(delay):
		}
	}
	if lastErr != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return false, lastErr
	}
	_ = tmpFile.Close()

	// Replace existing file atomically when possible
	if _, err := os.Stat(finalPath); err == nil {
		_ = os.Remove(finalPath)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return false, fmt.Errorf("failed to move temp file into place: %w", err)
	}

	ms.logger.Info("Downloaded mod", zap.String("filename", filename))
	return true, nil
}
