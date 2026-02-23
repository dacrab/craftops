package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"craftops/internal/service"
)

// modrinthVersionFixture returns a minimal Modrinth API version response.
func modrinthVersionFixture(filename, downloadURL string) []map[string]any {
	return []map[string]any{
		{
			"id":             "AABBccDD",
			"version_number": "1.0.0",
			"files": []map[string]any{
				{"filename": filename, "url": downloadURL},
			},
		},
	}
}

// newMockModrinth spins up a test HTTP server simulating the Modrinth API.
// versionPath is the path prefix that returns versions (e.g. "/v2/project/fabric-api/version").
// downloadPath is the path that serves the jar bytes.
func newMockModrinth(t *testing.T, versionPath, downloadPath string, jarContent []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, versionPath):
			filename := "mod-1.0.0.jar"
			dlURL := "http://" + r.Host + downloadPath
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(modrinthVersionFixture(filename, dlURL))

		case r.URL.Path == downloadPath:
			w.Header().Set("Content-Type", "application/java-archive")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jarContent)

		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestMods_UpdateAll_Downloads(t *testing.T) {
	cfg, logger, ctx := setup(t)

	srv := newMockModrinth(t,
		"/v2/project/fabric-api/version",
		"/files/mod-1.0.0.jar",
		[]byte("FAKE_JAR_CONTENT"),
	)

	// Point the mods service at the mock server by using a slug and patching
	// the API base. We do this by using the test server URL as the source —
	// parseProjectID accepts bare slugs, and fetchLatestVersion builds the URL
	// from the configured API base. We override via the config source list and
	// patch the client's transport to redirect to the mock.
	cfg.Mods.ModrinthSources = []string{"fabric-api"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL)

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("UpdateAll error: %v", err)
	}

	if len(result.FailedMods) > 0 {
		t.Errorf("unexpected failures: %v", result.FailedMods)
	}
	if len(result.UpdatedMods) != 1 {
		t.Errorf("expected 1 updated mod, got %d (skipped=%v failed=%v)",
			len(result.UpdatedMods), result.SkippedMods, result.FailedMods)
	}

	// Verify the jar was written to disk
	jar := filepath.Join(cfg.Paths.Mods, "mod-1.0.0.jar")
	data, err := os.ReadFile(jar) //nolint:gosec
	if err != nil {
		t.Fatalf("jar not written to disk: %v", err)
	}
	if string(data) != "FAKE_JAR_CONTENT" {
		t.Errorf("jar content mismatch: got %q", data)
	}
}

func TestMods_UpdateAll_SkipsExisting(t *testing.T) {
	cfg, logger, ctx := setup(t)

	srv := newMockModrinth(t,
		"/v2/project/sodium/version",
		"/files/mod-1.0.0.jar",
		[]byte("FAKE"),
	)

	cfg.Mods.ModrinthSources = []string{"sodium"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	// Pre-place the jar so it appears "already installed"
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "mod-1.0.0.jar"), []byte("OLD"), 0o600)

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL)

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("UpdateAll error: %v", err)
	}
	if len(result.SkippedMods) != 1 {
		t.Errorf("expected 1 skipped mod, got updated=%v skipped=%v failed=%v",
			result.UpdatedMods, result.SkippedMods, result.FailedMods)
	}
}

func TestMods_UpdateAll_ForceRedownload(t *testing.T) {
	cfg, logger, ctx := setup(t)

	srv := newMockModrinth(t,
		"/v2/project/sodium/version",
		"/files/mod-1.0.0.jar",
		[]byte("NEW_CONTENT"),
	)

	cfg.Mods.ModrinthSources = []string{"sodium"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	// Pre-place old jar
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "mod-1.0.0.jar"), []byte("OLD"), 0o600)

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL)

	result, err := svc.UpdateAll(ctx, true) // force=true
	if err != nil {
		t.Fatalf("UpdateAll(force) error: %v", err)
	}
	if len(result.UpdatedMods) != 1 {
		t.Errorf("expected 1 updated mod with force, got skipped=%v updated=%v failed=%v",
			result.SkippedMods, result.UpdatedMods, result.FailedMods)
	}

	data, _ := os.ReadFile(filepath.Join(cfg.Paths.Mods, "mod-1.0.0.jar"))
	if string(data) != "NEW_CONTENT" {
		t.Errorf("expected NEW_CONTENT after force update, got %q", data)
	}
}

func TestMods_UpdateAll_API404(t *testing.T) {
	cfg, logger, ctx := setup(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	t.Cleanup(srv.Close)

	cfg.Mods.ModrinthSources = []string{"nonexistent-mod"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL)

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("UpdateAll should not return top-level error: %v", err)
	}
	if len(result.FailedMods) != 1 {
		t.Errorf("expected 1 failed mod for 404, got %v", result.FailedMods)
	}
}

func TestMods_UpdateAll_NoCompatibleVersions(t *testing.T) {
	cfg, logger, ctx := setup(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{}) // empty versions
	}))
	t.Cleanup(srv.Close)

	cfg.Mods.ModrinthSources = []string{"some-mod"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL)

	result, err := svc.UpdateAll(ctx, false)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(result.FailedMods) != 1 {
		t.Errorf("expected 1 failed mod for empty versions, got %v", result.FailedMods)
	}
}
