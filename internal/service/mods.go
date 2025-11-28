package service

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
	"craftops/internal/domain"
)

// Mods implements ModManager
type Mods struct {
	cfg    *config.Config
	logger *zap.Logger
	client *http.Client
}

var _ ModManager = (*Mods)(nil)

// NewMods creates a new mods service
func NewMods(cfg *config.Config, logger *zap.Logger) *Mods {
	return &Mods{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: time.Duration(cfg.Mods.Timeout) * time.Second},
	}
}

// UpdateAll updates all configured mods
func (m *Mods) UpdateAll(ctx context.Context, force bool) (*domain.ModUpdateResult, error) {
	m.logger.Info("Starting mod update", zap.Bool("force", force))

	result := &domain.ModUpdateResult{
		UpdatedMods: []string{},
		FailedMods:  make(map[string]string),
		SkippedMods: []string{},
	}

	if m.cfg.DryRun {
		m.logger.Info("Dry run mode - no changes")
		result.UpdatedMods = []string{"example-mod (dry-run)"}
		return result, nil
	}

	sources := m.cfg.Mods.ModrinthSources
	if len(sources) == 0 {
		m.logger.Info("No mod sources configured")
		return result, nil
	}

	m.processParallel(ctx, sources, force, result)

	m.logger.Info("Mod update completed",
		zap.Int("updated", len(result.UpdatedMods)),
		zap.Int("failed", len(result.FailedMods)),
		zap.Int("skipped", len(result.SkippedMods)))

	return result, nil
}

// ListInstalled lists all installed mods
func (m *Mods) ListInstalled() ([]domain.InstalledMod, error) {
	files, err := filepath.Glob(filepath.Join(m.cfg.Paths.Mods, "*.jar"))
	if err != nil {
		return nil, fmt.Errorf("failed to list mod files: %w", err)
	}

	mods := make([]domain.InstalledMod, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		filename := filepath.Base(file)
		name := strings.TrimSuffix(filename, filepath.Ext(filename))

		mods = append(mods, domain.InstalledMod{
			Name:     name,
			Filename: filename,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	return mods, nil
}

// HealthCheck performs health checks
func (m *Mods) HealthCheck(ctx context.Context) []domain.HealthCheck {
	return []domain.HealthCheck{
		m.checkModsDirectory(),
		m.checkModSources(),
		m.checkAPI(ctx),
	}
}

func (m *Mods) processParallel(ctx context.Context, sources []string, force bool, result *domain.ModUpdateResult) {
	semaphore := make(chan struct{}, m.cfg.Mods.ConcurrentDownloads)
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

			updated, name, err := m.updateMod(ctx, src, force)
			display := name
			if display == "" {
				display = src
			}

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.FailedMods[display] = err.Error()
				m.logger.Error("Failed to update mod", zap.String("mod", display), zap.Error(err))
			} else if updated {
				result.UpdatedMods = append(result.UpdatedMods, display)
			} else {
				result.SkippedMods = append(result.SkippedMods, display)
			}
		}(source)
	}

	wg.Wait()
}

func (m *Mods) updateMod(ctx context.Context, modURL string, force bool) (bool, string, error) {
	projectID, err := m.parseProjectID(modURL)
	if err != nil {
		return false, projectID, err
	}

	info, err := m.fetchLatestVersion(ctx, projectID)
	if err != nil {
		return false, projectID, err
	}

	updated, err := m.downloadMod(ctx, info, force)
	return updated, info.ProjectName, err
}

func (m *Mods) parseProjectID(modURL string) (string, error) {
	u, err := url.Parse(modURL)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`/mod/([^/]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid Modrinth URL: %s", modURL)
	}

	return matches[1], nil
}

type modrinthFile struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

type modrinthVersion struct {
	ID            string         `json:"id"`
	VersionNumber string         `json:"version_number"`
	Files         []modrinthFile `json:"files"`
}

func (m *Mods) fetchLatestVersion(ctx context.Context, projectID string) (*domain.ModInfo, error) {
	apiURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version?game_versions=[\"%s\"]&loaders=[\"%s\"]",
		projectID, m.cfg.Minecraft.Version, m.cfg.Minecraft.Modloader)

	var versions []modrinthVersion
	if err := m.apiRequest(ctx, apiURL, &versions); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no compatible versions found")
	}

	v := versions[0]
	if len(v.Files) == 0 {
		return nil, fmt.Errorf("no files in version")
	}

	return &domain.ModInfo{
		VersionID:   v.ID,
		Version:     v.VersionNumber,
		DownloadURL: v.Files[0].URL,
		Filename:    v.Files[0].Filename,
		ProjectName: projectID,
	}, nil
}

func (m *Mods) apiRequest(ctx context.Context, apiURL string, result interface{}) error {
	var lastErr error
	for attempt := 0; attempt <= m.cfg.Mods.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "craftops/2.0")

		resp, err := m.client.Do(req)
		if err != nil {
			lastErr = err
			m.backoff(ctx, attempt)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 {
			return json.Unmarshal(body, result)
		}

		lastErr = &domain.APIError{URL: apiURL, StatusCode: resp.StatusCode, Message: "request failed"}
		if resp.StatusCode < 500 && resp.StatusCode != 429 {
			break
		}
		m.backoff(ctx, attempt)
	}
	return lastErr
}

func (m *Mods) downloadMod(ctx context.Context, info *domain.ModInfo, force bool) (bool, error) {
	if err := os.MkdirAll(m.cfg.Paths.Mods, 0o750); err != nil {
		return false, err
	}

	finalPath := filepath.Join(m.cfg.Paths.Mods, info.Filename)
	if !force {
		if _, err := os.Stat(finalPath); err == nil {
			m.logger.Info("Mod up-to-date, skipping", zap.String("filename", info.Filename))
			return false, nil
		}
	}

	tmpFile, err := os.CreateTemp(m.cfg.Paths.Mods, ".tmp-*")
	if err != nil {
		return false, err
	}
	tmpPath := tmpFile.Name()

	var lastErr error
	for attempt := 0; attempt <= m.cfg.Mods.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", info.DownloadURL, nil)
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return false, err
		}
		req.Header.Set("User-Agent", "craftops/2.0")

		resp, err := m.client.Do(req)
		if err != nil {
			lastErr = err
			m.backoff(ctx, attempt)
			continue
		}

		if resp.StatusCode == 200 {
			_, err = io.Copy(tmpFile, resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
				m.backoff(ctx, attempt)
				continue
			}
			lastErr = nil
			break
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("download failed: status %d", resp.StatusCode)
		if resp.StatusCode < 500 && resp.StatusCode != 429 {
			break
		}
		m.backoff(ctx, attempt)
	}

	tmpFile.Close()
	if lastErr != nil {
		os.Remove(tmpPath)
		return false, lastErr
	}

	os.Remove(finalPath)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return false, err
	}

	m.logger.Info("Downloaded mod", zap.String("filename", info.Filename))
	return true, nil
}

func (m *Mods) backoff(ctx context.Context, attempt int) {
	delay := time.Duration(m.cfg.Mods.RetryDelay*float64(attempt+1)) * time.Second
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

func (m *Mods) checkModsDirectory() domain.HealthCheck {
	info, err := os.Stat(m.cfg.Paths.Mods)
	if err != nil || !info.IsDir() {
		return domain.HealthCheck{Name: "Mods directory", Status: domain.StatusError, Message: "Not found"}
	}

	jarCount := 0
	if files, err := filepath.Glob(filepath.Join(m.cfg.Paths.Mods, "*.jar")); err == nil {
		jarCount = len(files)
	}
	return domain.HealthCheck{
		Name:    "Mods directory",
		Status:  domain.StatusOK,
		Message: fmt.Sprintf("OK (%d mods)", jarCount),
	}
}

func (m *Mods) checkModSources() domain.HealthCheck {
	total := len(m.cfg.Mods.ModrinthSources)
	if total == 0 {
		return domain.HealthCheck{Name: "Mod sources", Status: domain.StatusWarn, Message: "None configured"}
	}
	return domain.HealthCheck{Name: "Mod sources", Status: domain.StatusOK, Message: fmt.Sprintf("%d sources", total)}
}

func (m *Mods) checkAPI(ctx context.Context) domain.HealthCheck {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.modrinth.com/v2/", nil)
	resp, err := m.client.Do(req)
	if err != nil {
		return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusError, Message: "Connection failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusWarn, Message: fmt.Sprintf("Status %d", resp.StatusCode)}
	}
	return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusOK, Message: "Connected"}
}
