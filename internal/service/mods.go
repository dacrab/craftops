package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

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
	res := &domain.ModUpdateResult{
		UpdatedMods: []string{},
		FailedMods:  make(map[string]string),
		SkippedMods: []string{},
	}

	sources := m.cfg.Mods.ModrinthSources
	if len(sources) == 0 {
		return res, nil
	}

	var mu sync.Mutex
	sem := semaphore.NewWeighted(int64(m.cfg.Mods.ConcurrentDownloads))
	g, gctx := errgroup.WithContext(ctx)

	for _, src := range sources {
		if err := sem.Acquire(gctx, 1); err != nil {
			continue
		}
		g.Go(func() error {
			defer sem.Release(1)
			updated, name, err := m.updateMod(gctx, src, force)
			if name == "" {
				name = src
			}
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				res.FailedMods[name] = err.Error()
			case updated:
				res.UpdatedMods = append(res.UpdatedMods, name)
			default:
				res.SkippedMods = append(res.SkippedMods, name)
			}
			return nil
		})
	}
	_ = g.Wait() // errors are captured in res.FailedMods
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

// withRetry retries op up to MaxRetries times with a fixed delay between attempts.
// Non-retryable API errors are returned immediately.
func (m *Mods) withRetry(ctx context.Context, op func() error) error {
	maxRetries := m.cfg.Mods.MaxRetries
	delay := time.Duration(m.cfg.Mods.RetryDelay * float64(time.Second))
	var apiErr *domain.APIError
	var err error
	for attempt := range maxRetries + 1 {
		if err = op(); err == nil {
			return nil
		}
		if errors.As(err, &apiErr) && !apiErr.IsRetryable() {
			return err
		}
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return err // return the actual last error, not a synthetic one
}

// apiRequest performs a JSON GET request with retries
func (m *Mods) apiRequest(ctx context.Context, apiURL string, result any) error {
	return m.withRetry(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "craftops/2.0")

		resp, err := m.client.Do(req) //nolint:gosec // URL is constructed from Modrinth API base URL + config values, not raw user input
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
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
	defer func() {
		if removeErr := os.Remove(tmpPath); removeErr != nil {
			m.logger.Warn("Failed to remove temporary file", zap.String("path", tmpPath), zap.Error(removeErr))
		}
	}()

	err = m.withRetry(ctx, func() error {
		if _, err := tmpFile.Seek(0, 0); err != nil {
			return err
		}
		if err := tmpFile.Truncate(0); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.DownloadURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "craftops/2.0")

		resp, err := m.client.Do(req) //nolint:gosec // URL comes from Modrinth API response, not user input
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed: status %d", resp.StatusCode)
		}

		_, err = io.Copy(tmpFile, resp.Body)
		return err
	})

	if closeErr := tmpFile.Close(); closeErr != nil {
		m.logger.Warn("Failed to close temporary file", zap.Error(closeErr))
	}
	if err != nil {
		return false, err
	}

	_ = os.Remove(finalPath)
	if err := os.Rename(tmpPath, finalPath); err != nil { //nolint:gosec // path derived from validated config + Modrinth API slug
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

// parseProjectID extracts the Modrinth slug from a full URL or bare slug.
func (m *Mods) parseProjectID(modURL string) (string, error) {
	// Handle direct slug
	if !strings.Contains(modURL, "/") {
		return modURL, nil
	}
	// Extract from URL path (e.g., https://modrinth.com/mod/fabric-api -> fabric-api)
	if idx := strings.LastIndex(modURL, "/mod/"); idx != -1 {
		slug := strings.TrimPrefix(modURL[idx+5:], "/")
		if idx := strings.Index(slug, "/"); idx != -1 {
			slug = slug[:idx]
		}
		if slug != "" {
			return slug, nil
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
		return nil, errors.New("no compatible versions found")
	}

	v := versions[0]
	if len(v.Files) == 0 {
		return nil, errors.New("no files in version")
	}

	return &domain.ModInfo{
		VersionID:   v.ID,
		Version:     v.VersionNumber,
		DownloadURL: v.Files[0].URL,
		Filename:    v.Files[0].Filename,
		ProjectName: projectID,
	}, nil
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.modrinth.com/v2/", nil)
	if err != nil {
		return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusError, Message: "Failed to build request"}
	}
	resp, err := m.client.Do(req) //nolint:gosec // fixed known-good URL
	if err != nil {
		return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusError, Message: "Connection failed"}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusWarn, Message: fmt.Sprintf("Status %d", resp.StatusCode)}
	}
	return domain.HealthCheck{Name: "Modrinth API", Status: domain.StatusOK, Message: "Connected"}
}
