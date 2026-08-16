package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	// Point the mods service at the mock server via NewModsWithBaseURL.
	cfg.Mods.ModrinthSources = []string{"fabric-api"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")

	result := svc.UpdateAll(ctx, false)

	if len(result.FailedMods) > 0 {
		t.Errorf("unexpected failures: %v", result.FailedMods)
	}
	if len(result.UpdatedMods) != 1 {
		t.Errorf("expected 1 updated mod, got %d (skipped=%v failed=%v)",
			len(result.UpdatedMods), result.SkippedMods, result.FailedMods)
	}

	// Verify the jar was written to disk
	jar := filepath.Join(cfg.Paths.Mods, "mod-1.0.0.jar")
	data, err := os.ReadFile(jar)
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

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")

	result := svc.UpdateAll(ctx, false)
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

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")

	result := svc.UpdateAll(ctx, true) // force=true
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

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")

	result := svc.UpdateAll(ctx, false)
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

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")

	result := svc.UpdateAll(ctx, false)
	if len(result.FailedMods) != 1 {
		t.Errorf("expected 1 failed mod for empty versions, got %v", result.FailedMods)
	}
}

func TestMods_ListInstalled_Empty(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	mods, err := svc.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled error: %v", err)
	}
	if len(mods) != 0 {
		t.Errorf("expected 0 mods, got %d", len(mods))
	}
}

func TestMods_ListInstalled(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "fabric-api.jar"), []byte("jar"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "sodium.jar"), []byte("jar"), 0o600)
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "readme.txt"), []byte("text"), 0o600)

	mods, err := svc.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled error: %v", err)
	}
	if len(mods) != 2 {
		t.Errorf("expected 2 mods, got %d", len(mods))
	}
}

func TestParseProjectID(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"fabric-api", "fabric-api", false},
		{"https://modrinth.com/mod/fabric-api", "fabric-api", false},
		{"https://modrinth.com/mod/sodium/versions", "sodium", false},
		{"https://invalid.com/notamod", "", true},
	}
	for _, tt := range tests {
		got, err := service.ParseProjectID(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseProjectID(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseProjectID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMods_UpdateAll_NoSources(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Mods.ModrinthSources = []string{}
	svc := service.NewMods(cfg, logger)

	result := svc.UpdateAll(ctx, false)
	if len(result.UpdatedMods) != 0 || len(result.FailedMods) != 0 || len(result.SkippedMods) != 0 {
		t.Error("expected empty results with no sources")
	}
}

func TestMods_ListInstalled_Metadata(t *testing.T) {
	cfg, logger, _ := setup(t)
	svc := service.NewMods(cfg, logger)

	content := []byte("fake jar content")
	_ = os.WriteFile(filepath.Join(cfg.Paths.Mods, "fabric-api.jar"), content, 0o600)

	mods, err := svc.ListInstalled()
	if err != nil {
		t.Fatalf("ListInstalled error: %v", err)
	}
	if len(mods) != 1 {
		t.Fatalf("expected 1 mod, got %d", len(mods))
	}
	m := mods[0]
	if m.Name != "fabric-api" {
		t.Errorf("Name = %q, want %q", m.Name, "fabric-api")
	}
	if m.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", m.Size, len(content))
	}
}

func TestMods_UpdateAll_SendsJSONArrayFilters(t *testing.T) {
	cfg, logger, ctx := setup(t)
	cfg.Minecraft.Version = "1.20.1"
	cfg.Minecraft.Modloader = "fabric"
	cfg.Mods.ModrinthSources = []string{"fabric-api"}
	cfg.Mods.MaxRetries = 0
	cfg.Mods.Timeout = 5

	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v2/project/") {
			query = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			dlURL := "http://" + r.Host + "/files/mod-1.0.0.jar"
			_ = json.NewEncoder(w).Encode(modrinthVersionFixture("mod-1.0.0.jar", dlURL))
			return
		}
		if r.URL.Path == "/files/mod-1.0.0.jar" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("JAR"))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")
	if result := svc.UpdateAll(ctx, false); len(result.FailedMods) > 0 {
		t.Fatalf("unexpected failures: %v", result.FailedMods)
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("failed to parse query %q: %v", query, err)
	}
	for _, tc := range []struct{ key, want string }{
		{"game_versions", "1.20.1"},
		{"loaders", "fabric"},
	} {
		var decoded []string
		if err := json.Unmarshal([]byte(values.Get(tc.key)), &decoded); err != nil {
			t.Errorf("%s=%q does not decode as JSON array: %v", tc.key, values.Get(tc.key), err)
			continue
		}
		if len(decoded) != 1 || decoded[0] != tc.want {
			t.Errorf("%s decoded to %v, want [%q]", tc.key, decoded, tc.want)
		}
	}
}

func TestMods_HealthCheck(t *testing.T) {
	cfg, logger, ctx := setup(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	svc := service.NewModsWithBaseURL(cfg, logger, srv.URL+"/v2")

	checks := svc.HealthCheck(ctx)
	if len(checks) < 2 {
		t.Fatalf("expected at least 2 health checks, got %d", len(checks))
	}
	names := make(map[string]bool)
	for _, c := range checks {
		names[c.Name] = true
	}
	if !names["Mods directory"] {
		t.Error("expected 'Mods directory' health check")
	}
	if !names["Mod sources"] {
		t.Error("expected 'Mod sources' health check")
	}
}
