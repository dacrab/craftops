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

// HealthCheck represents a health check result
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewModService creates a new mod service instance
func NewModService(cfg *config.Config, logger *zap.Logger) *ModService {
	client := &http.Client{
		Timeout: time.Duration(cfg.Mods.Timeout) * time.Second,
	}

	return &ModService{
		config: cfg,
		logger: logger,
		client: client,
	}
}

// HealthCheck performs health checks for the mod service
func (ms *ModService) HealthCheck(ctx context.Context) []HealthCheck {
	checks := []HealthCheck{}

	// Check mods directory
	modsDir := ms.config.Paths.Mods
	if info, err := os.Stat(modsDir); err == nil && info.IsDir() {
		// Count JAR files
		jarCount := 0
		if files, err := filepath.Glob(filepath.Join(modsDir, "*.jar")); err == nil {
			jarCount = len(files)
		}
		checks = append(checks, HealthCheck{
			Name:    "Mods directory",
			Status:  "✅",
			Message: fmt.Sprintf("OK (%d mods found)", jarCount),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Mods directory",
			Status:  "❌",
			Message: "Directory not found or not accessible",
		})
	}

	// Check mod sources configuration
	totalSources := len(ms.config.Mods.Sources.Modrinth)

	if totalSources > 0 {
		checks = append(checks, HealthCheck{
			Name:    "Mod sources",
			Status:  "✅",
			Message: fmt.Sprintf("%d sources configured", totalSources),
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Mod sources",
			Status:  "⚠️",
			Message: "No mod sources configured",
		})
	}

	// Test API connectivity
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.modrinth.com/v2/", nil)
	if err == nil {
		resp, err := ms.client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				checks = append(checks, HealthCheck{
					Name:    "Modrinth API",
					Status:  "✅",
					Message: "API accessible",
				})
			} else {
				checks = append(checks, HealthCheck{
					Name:    "Modrinth API",
					Status:  "⚠️",
					Message: fmt.Sprintf("API returned status %d", resp.StatusCode),
				})
			}
		} else {
			checks = append(checks, HealthCheck{
				Name:    "Modrinth API",
				Status:  "❌",
				Message: fmt.Sprintf("Connection failed: %v", err),
			})
		}
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Modrinth API",
			Status:  "❌",
			Message: fmt.Sprintf("Request creation failed: %v", err),
		})
	}

	return checks
}

// UpdateAllMods updates all configured mods to their latest versions
func (ms *ModService) UpdateAllMods(ctx context.Context, force bool) (*ModUpdateResult, error) {
	ms.logger.Info("Starting mod update process", zap.Bool("force", force))

	if ms.config.DryRun {
		ms.logger.Info("Dry run mode - no actual changes will be made")
		return &ModUpdateResult{
			UpdatedMods: []string{"example-mod (dry-run)"},
			FailedMods:  make(map[string]string),
			SkippedMods: []string{},
		}, nil
	}

	result := &ModUpdateResult{
		UpdatedMods: []string{},
		FailedMods:  make(map[string]string),
		SkippedMods: []string{},
	}

	// Collect all mod sources
	var allSources []modSource

	// Add Modrinth sources
	for _, url := range ms.config.Mods.Sources.Modrinth {
		allSources = append(allSources, modSource{
			Type: "modrinth",
			URL:  url,
		})
	}

	// Note: CurseForge and GitHub support coming in future releases
	// For now, only Modrinth is fully supported

	if len(allSources) == 0 {
		ms.logger.Info("No mod sources configured")
		return result, nil
	}

	// Process mods concurrently
	semaphore := make(chan struct{}, ms.config.Mods.ConcurrentDownloads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, source := range allSources {
		wg.Add(1)
		go func(src modSource) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := ms.updateSingleMod(ctx, src, force); err != nil {
				mu.Lock()
				result.FailedMods[src.URL] = err.Error()
				mu.Unlock()
				ms.logger.Error("Failed to update mod", zap.String("url", src.URL), zap.Error(err))
			} else {
				mu.Lock()
				result.UpdatedMods = append(result.UpdatedMods, src.URL)
				mu.Unlock()
			}
		}(source)
	}

	wg.Wait()

	ms.logger.Info("Mod update process completed",
		zap.Int("updated", len(result.UpdatedMods)),
		zap.Int("failed", len(result.FailedMods)),
		zap.Int("skipped", len(result.SkippedMods)))

	return result, nil
}

// ListInstalledMods lists all currently installed mods
func (ms *ModService) ListInstalledMods() ([]map[string]interface{}, error) {
	modsDir := ms.config.Paths.Mods
	mods := []map[string]interface{}{}

	files, err := filepath.Glob(filepath.Join(modsDir, "*.jar"))
	if err != nil {
		return nil, fmt.Errorf("failed to list mod files: %w", err)
	}

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

// modSource represents a mod source
type modSource struct {
	Type string
	URL  string
}

// updateSingleMod updates a single mod
func (ms *ModService) updateSingleMod(ctx context.Context, source modSource, force bool) error {
	switch source.Type {
	case "modrinth":
		return ms.updateModrinthMod(ctx, source.URL, force)
	case "curseforge":
		return ms.updateCurseForgeMod(ctx, source.URL, force)
	case "github":
		return ms.updateGitHubMod(ctx, source.URL, force)
	default:
		return fmt.Errorf("unsupported mod source type: %s", source.Type)
	}
}

// updateModrinthMod updates a mod from Modrinth
func (ms *ModService) updateModrinthMod(ctx context.Context, modURL string, _ bool) error {
	// Parse project ID from URL
	projectID, err := ms.parseModrinthProjectID(modURL)
	if err != nil {
		return fmt.Errorf("failed to parse Modrinth project ID: %w", err)
	}

	// Get project info (for future use)
	_, err = ms.fetchModrinthProjectInfo(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch project info: %w", err)
	}

	// Get latest version
	versionInfo, err := ms.fetchModrinthLatestVersion(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %w", err)
	}

	// Download the mod
	return ms.downloadMod(ctx, versionInfo.DownloadURL, versionInfo.Filename)
}

// updateCurseForgeMod updates a mod from CurseForge
// Note: CurseForge support is planned for a future release
func (ms *ModService) updateCurseForgeMod(_ context.Context, _ string, _ bool) error {
	return fmt.Errorf("CurseForge mod updating is not yet implemented - coming in v2.1.0")
}

// updateGitHubMod updates a mod from GitHub
// Note: GitHub support is planned for a future release
func (ms *ModService) updateGitHubMod(_ context.Context, _ string, _ bool) error {
	return fmt.Errorf("GitHub mod updating is not yet implemented - coming in v2.1.0")
}

// parseModrinthProjectID parses project ID from Modrinth URL
func (ms *ModService) parseModrinthProjectID(modURL string) (string, error) {
	u, err := url.Parse(modURL)
	if err != nil {
		return "", err
	}

	// Match pattern like /mod/project-id
	re := regexp.MustCompile(`/mod/([^/]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Modrinth URL format: %s", modURL)
	}

	return matches[1], nil
}

// fetchModrinthProjectInfo fetches project information from Modrinth API
func (ms *ModService) fetchModrinthProjectInfo(ctx context.Context, projectID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.modrinth.com/v2/project/%s", projectID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ms.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// fetchModrinthLatestVersion fetches the latest compatible version from Modrinth API
func (ms *ModService) fetchModrinthLatestVersion(ctx context.Context, projectID string) (*ModInfo, error) {
	url := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version?game_versions=[\"%s\"]&loaders=[\"%s\"]",
		projectID, ms.config.Minecraft.Version, ms.config.Minecraft.Modloader)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ms.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions []map[string]interface{}
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no compatible versions found")
	}

	// Get the first (latest) version
	version := versions[0]
	files, ok := version["files"].([]interface{})
	if !ok || len(files) == 0 {
		return nil, fmt.Errorf("no files found in version")
	}

	file := files[0].(map[string]interface{})

	return &ModInfo{
		VersionID:   version["id"].(string),
		Version:     version["version_number"].(string),
		DownloadURL: file["url"].(string),
		Filename:    file["filename"].(string),
		ProjectName: projectID, // We'll use project ID as name for now
	}, nil
}

// downloadMod downloads a mod file
func (ms *ModService) downloadMod(ctx context.Context, downloadURL, filename string) error {
	modsDir := ms.config.Paths.Mods

	// Ensure mods directory exists
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mods directory: %w", err)
	}

	// Create the file
	filePath := filepath.Join(modsDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Download the file
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := ms.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ms.logger.Info("Downloaded mod", zap.String("filename", filename))
	return nil
}
