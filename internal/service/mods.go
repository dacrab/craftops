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
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
	"craftops/internal/domain"
)

// Mods handles automated mod updates from Modrinth API
type Mods struct {
	cfg    *config.Config
	logger *zap.Logger
	client *http.Client
}

var _ ModManager = (*Mods)(nil)

// NewMods initializes a new mod management service
func NewMods(cfg *config.Config, logger *zap.Logger) *Mods {
	return &Mods{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{Timeout: time.Duration(cfg.Mods.Timeout) * time.Second},
	}
}

// UpdateAll identifies and downloads the latest versions of all configured mods
func (m *Mods) UpdateAll(ctx context.Context, force bool) (*domain.ModUpdateResult, error) {
	m.logger.Info("Starting mod update", zap.Bool("force", force))
	res := &domain.ModUpdateResult{UpdatedMods: []string{}, FailedMods: make(map[string]string), SkippedMods: []string{}}

	sources := m.cfg.Mods.ModrinthSources
	if len(sources) == 0 {
		return res, nil
	}

	// Use semaphore to limit concurrent downloads
	sem := make(chan struct{}, m.cfg.Mods.ConcurrentDownloads)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, src := range sources {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			updated, name, err := m.updateMod(ctx, s, force)
			if name == "" {
				name = s
			}

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				res.FailedMods[name] = err.Error()
			} else if updated {
				res.UpdatedMods = append(res.UpdatedMods, name)
			} else {
				res.SkippedMods = append(res.SkippedMods, name)
			}
		}(src)
	}
	wg.Wait()
	return res, nil
}

// ListInstalled returns a list of all .jar files in the mods directory
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
		mods = append(mods, domain.InstalledMod{
			Name:     strings.TrimSuffix(filename, filepath.Ext(filename)),
			Filename: filename,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	return mods, nil
}

// HealthCheck verifies mods directory availability and API connectivity
func (m *Mods) HealthCheck(ctx context.Context) []domain.HealthCheck {
	return []domain.HealthCheck{
		domain.CheckPath("Mods directory", m.cfg.Paths.Mods),
		m.checkModSources(),
		m.checkAPI(ctx),
	}
}

// withRetry wraps an operation with a backoff-based retry loop
func (m *Mods) withRetry(ctx context.Context, op func() error) error {
	var lastErr error
	for attempt := 0; attempt <= m.cfg.Mods.MaxRetries; attempt++ {
		if err := op(); err == nil {
			return nil
		} else {
			lastErr = err
			if apiErr, ok := err.(*domain.APIError); ok && !apiErr.IsRetryable() {
				break
			}
		}
		if attempt < m.cfg.Mods.MaxRetries {
			m.backoff(ctx, attempt)
		}
	}
	return lastErr
}

// apiRequest performs a JSON GET request with retries
func (m *Mods) apiRequest(ctx context.Context, apiURL string, result interface{}) error {
	return m.withRetry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "craftops/2.0")

		resp, err := m.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return &domain.APIError{URL: apiURL, StatusCode: resp.StatusCode, Message: "request failed"}
		}

		return json.NewDecoder(resp.Body).Decode(result)
	})
}

// downloadMod handles the actual file transfer and persistence
func (m *Mods) downloadMod(ctx context.Context, info *domain.ModInfo, force bool) (bool, error) {
	if m.cfg.DryRun {
		m.logger.Info("Dry run: Would download mod", zap.String("filename", info.Filename))
		return true, nil
	}
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
	defer os.Remove(tmpPath)

	err = m.withRetry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", info.DownloadURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "craftops/2.0")

		resp, err := m.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("download failed: status %d", resp.StatusCode)
		}

		if _, err := tmpFile.Seek(0, 0); err != nil {
			return err
		}
		if err := tmpFile.Truncate(0); err != nil {
			return err
		}

		_, err = io.Copy(tmpFile, resp.Body)
		return err
	})

	tmpFile.Close()
	if err != nil {
		return false, err
	}

	os.Remove(finalPath)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return false, err
	}

	m.logger.Info("Downloaded mod", zap.String("filename", info.Filename))
	return true, nil
}

// updateMod performs high-level mod update flow for a single URL
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

// parseProjectID extracts the Modrinth slug from a full URL
func (m *Mods) parseProjectID(modURL string) (string, error) {
	u, err := url.Parse(modURL)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, p := range parts {
		if p == "mod" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("invalid Modrinth URL: %s", modURL)
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

// fetchLatestVersion queries the Modrinth API for the newest compatible release
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

func (m *Mods) backoff(ctx context.Context, attempt int) {
	delay := time.Duration(m.cfg.Mods.RetryDelay*float64(attempt+1)) * time.Second
	select {
	case <-ctx.Done():
	case <-time.After(delay):
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
