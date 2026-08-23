package service

import (
	"context"
	"encoding/json"
	"errors"
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
	"craftops/internal/ui"
)

const (
	apiBase       = "https://api.modrinth.com/v2"
	apiHealthName = "Modrinth API"
)

// Version is set by ldflags during build.
var Version = "dev"

// Mods handles automated mod updates from Modrinth.
type Mods struct {
	cfg     *config.Config
	logger  *zap.Logger
	client  *http.Client
	baseURL string
}

// NewMods creates a mod manager. The HTTP client has no global timeout;
// API calls get a per-request deadline from cfg.Mods.Timeout while mod
// downloads are bounded only by the parent context (large jars on slow
// links must not be killed mid-transfer).
func NewMods(cfg *config.Config, logger *zap.Logger) *Mods {
	return &Mods{
		cfg:     cfg,
		logger:  logger,
		client:  &http.Client{},
		baseURL: apiBase,
	}
}

func (m *Mods) userAgent() string {
	return "craftops/" + Version
}

// requestCtx bounds a metadata/health API call by cfg.Mods.Timeout.
func (m *Mods) requestCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := m.cfg.Mods.Timeout
	if timeout <= 0 {
		timeout = 10
	}
	return context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
}

// UpdateAll downloads the latest versions of all configured mods concurrently.
func (m *Mods) UpdateAll(ctx context.Context, force bool) *domain.ModUpdateResult {
	m.logger.Info("Starting mod update", zap.Bool("force", force))
	res := &domain.ModUpdateResult{
		UpdatedMods: []string{},
		FailedMods:  make(map[string]string),
		SkippedMods: []string{},
	}

	sources := m.cfg.Mods.ModrinthSources
	if len(sources) == 0 {
		return res
	}

	workers := m.cfg.Mods.ConcurrentDownloads
	if workers < 1 {
		workers = 1
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	jobs := make(chan string)

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for src := range jobs {
				updated, name, err := m.updateMod(ctx, src, force)
				if name == "" {
					name = src
				}
				mu.Lock()
				switch {
				case err != nil:
					res.FailedMods[name] = err.Error()
				case updated:
					res.UpdatedMods = append(res.UpdatedMods, name)
				default:
					res.SkippedMods = append(res.SkippedMods, name)
				}
				mu.Unlock()
			}
		}()
	}

	for _, src := range sources {
		jobs <- src
	}
	close(jobs)
	wg.Wait()

	return res
}

// ListInstalled returns all .jar files in the mods directory.
func (m *Mods) ListInstalled() ([]domain.InstalledMod, error) {
	files, err := filepath.Glob(filepath.Join(m.cfg.Paths.Mods, "*.jar"))
	if err != nil {
		return nil, fmt.Errorf("failed to list mod files: %w", err)
	}

	mods := make([]domain.InstalledMod, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("failed to stat mod %s: %w", file, err)
		}
		filename := filepath.Base(file)
		mods = append(mods, domain.InstalledMod{
			Name:     strings.TrimSuffix(filename, filepath.Ext(filename)),
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}
	return mods, nil
}

// HealthCheck verifies mods directory and API connectivity.
func (m *Mods) HealthCheck(ctx context.Context) []domain.HealthCheck {
	total := len(m.cfg.Mods.ModrinthSources)
	var sourcesCheck domain.HealthCheck
	if total == 0 {
		sourcesCheck = domain.HealthCheck{Name: "Mod sources", Status: domain.StatusWarn, Message: "None configured"}
	} else {
		sourcesCheck = domain.HealthCheck{Name: "Mod sources", Status: domain.StatusOK, Message: fmt.Sprintf("%d sources", total)}
	}
	return []domain.HealthCheck{
		ui.CheckPath("Mods directory", m.cfg.Paths.Mods),
		sourcesCheck,
		m.checkAPI(ctx),
	}
}

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
	return err
}

// get performs a GET request with the Modrinth user agent.
func (m *Mods) get(ctx context.Context, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", m.userAgent())
	return m.client.Do(req)
}

func (m *Mods) downloadMod(ctx context.Context, info *domain.ModInfo, force bool) (bool, error) {
	if m.cfg.DryRun {
		m.logger.Info("Dry run: Would download mod", zap.String("filename", info.Filename))
		return true, nil
	}
	if err := os.MkdirAll(m.cfg.Paths.Mods, domain.DirPerm); err != nil {
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

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	err = m.withRetry(ctx, func() error {
		if _, err := tmpFile.Seek(0, 0); err != nil {
			return err
		}
		if err := tmpFile.Truncate(0); err != nil {
			return err
		}

		resp, err := m.get(ctx, info.DownloadURL)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return &domain.APIError{URL: info.DownloadURL, StatusCode: resp.StatusCode, Message: "download failed"}
		}

		_, err = io.Copy(tmpFile, resp.Body)
		return err
	})

	if closeErr := tmpFile.Close(); closeErr != nil {
		return false, fmt.Errorf("closing temporary file: %w", closeErr)
	}
	if err != nil {
		return false, err
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		return false, err
	}

	success = true
	m.logger.Info("Downloaded mod", zap.String("filename", info.Filename))
	return true, nil
}

func (m *Mods) updateMod(ctx context.Context, modURL string, force bool) (updated bool, name string, err error) {
	projectID, err := parseProjectID(modURL)
	if err != nil {
		return false, projectID, err
	}

	info, err := m.fetchLatestVersion(ctx, projectID)
	if err != nil {
		return false, projectID, err
	}

	updated, err = m.downloadMod(ctx, info, force)
	return updated, info.ProjectName, err
}

// parseProjectID extracts the Modrinth slug from a full URL or bare slug.
func parseProjectID(modURL string) (string, error) {
	if !strings.Contains(modURL, "/") {
		return modURL, nil
	}
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
	Files []modrinthFile `json:"files"`
}

func (m *Mods) fetchLatestVersion(ctx context.Context, projectID string) (*domain.ModInfo, error) {
	ctx, cancel := m.requestCtx(ctx)
	defer cancel()

	q := url.Values{}
	q.Set("game_versions", "[\""+m.cfg.Minecraft.Version+"\"]")
	q.Set("loaders", "[\""+m.cfg.Minecraft.Modloader+"\"]")
	apiURL := fmt.Sprintf("%s/project/%s/version?%s", m.baseURL, projectID, q.Encode())

	var versions []modrinthVersion
	if err := m.withRetry(ctx, func() error {
		resp, err := m.get(ctx, apiURL)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return &domain.APIError{URL: apiURL, StatusCode: resp.StatusCode, Message: "request failed"}
		}
		return json.NewDecoder(resp.Body).Decode(&versions)
	}); err != nil {
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
		DownloadURL: v.Files[0].URL,
		Filename:    v.Files[0].Filename,
		ProjectName: projectID,
	}, nil
}

func (m *Mods) checkAPI(ctx context.Context) domain.HealthCheck {
	ctx, cancel := m.requestCtx(ctx)
	defer cancel()

	resp, err := m.get(ctx, m.baseURL)
	if err != nil {
		return domain.HealthCheck{Name: apiHealthName, Status: domain.StatusError, Message: "Connection failed"}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return domain.HealthCheck{Name: apiHealthName, Status: domain.StatusWarn, Message: fmt.Sprintf("Status %d", resp.StatusCode)}
	}
	return domain.HealthCheck{Name: apiHealthName, Status: domain.StatusOK, Message: "Connected"}
}
