package services

import (
	"context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "net/url"
    "os"
    "path/filepath"
    "strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"craftops/internal/config"
)

func TestServerService_StatusDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DryRun = true
	s := NewServerService(cfg, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("start (dry run) error: %v", err)
	}
	if err := s.Stop(ctx); err != nil {
		t.Fatalf("stop (dry run) error: %v", err)
	}
	if _, err := s.GetStatus(ctx); err != nil {
		t.Fatalf("get status error: %v", err)
	}
}

func TestBackupService_ListBackupsEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	b := NewBackupService(cfg, zap.NewNop())
	if list, err := b.ListBackups(); err != nil || list == nil {
		t.Fatalf("list backups err=%v nil=%v", err, list == nil)
	}
}

func TestNotificationService_NoWebhook(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Notifications.DiscordWebhook = ""
	n := NewNotificationService(cfg, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := n.SendSuccessNotification(ctx, "test"); err != nil {
		t.Fatalf("unexpected error without webhook: %v", err)
	}
}

func TestBackup_InternalHelpersAndRetention(t *testing.T) {
    // validateServerDirectory errors on nonexistent and non-directory
    cfg := config.DefaultConfig()
    cfg.Paths.Server = filepath.Join(t.TempDir(), "nope")
    bs := NewBackupService(cfg, zap.NewNop())
    if err := bs.validateServerDirectory(); err == nil {
        t.Fatalf("expected error for missing server dir")
    }
    tmp := t.TempDir()
    f := filepath.Join(tmp, "file")
    if err := os.WriteFile(f, []byte("x"), 0o644); err != nil { t.Fatal(err) }
    cfg.Paths.Server = f
    bs = NewBackupService(cfg, zap.NewNop())
    if err := bs.validateServerDirectory(); err == nil {
        t.Fatalf("expected error for non-directory server path")
    }

    // cleanupOldBackups keeps only MaxBackups most recent
    backups := filepath.Join(tmp, "backups")
    if err := os.MkdirAll(backups, 0o755); err != nil { t.Fatal(err) }
    cfg.Paths.Backups = backups
    cfg.Backup.MaxBackups = 1
    bs = NewBackupService(cfg, zap.NewNop())
    f1 := filepath.Join(backups, "a.tar.gz")
    f2 := filepath.Join(backups, "b.tar.gz")
    if err := os.WriteFile(f1, []byte("1"), 0o644); err != nil { t.Fatal(err) }
    time.Sleep(20 * time.Millisecond)
    if err := os.WriteFile(f2, []byte("2"), 0o644); err != nil { t.Fatal(err) }
    bs.cleanupOldBackups()
    files, err := filepath.Glob(filepath.Join(backups, "*.tar.gz"))
    if err != nil { t.Fatal(err) }
    if len(files) != 1 || filepath.Base(files[0]) != "b.tar.gz" {
        t.Fatalf("cleanupOldBackups kept %v", files)
    }
}

// rewriteTransport rewrites api.modrinth.com requests to a local httptest server
type rewriteTransport struct{ base *url.URL }

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    if strings.Contains(req.URL.Host, "api.modrinth.com") {
        req.URL.Scheme = rt.base.Scheme
        req.URL.Host = rt.base.Host
    }
    return http.DefaultTransport.RoundTrip(req)
}

func TestMod_FetchLatestVersion_And_Download(t *testing.T) {
    // Fake Modrinth API
    api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/v2/project/") && strings.HasSuffix(r.URL.Path, "/version") {
            resp := []map[string]interface{}{
                {
                    "id":             "v1",
                    "version_number": "1.0.0",
                    "files": []map[string]string{
                        {"url": "", "filename": "mod.jar"},
                    },
                },
            }
            _ = json.NewEncoder(w).Encode(resp)
            return
        }
        w.WriteHeader(404)
    }))
    defer api.Close()

    cfg := config.DefaultConfig()
    cfg.Minecraft.Version = "1.20.1"
    cfg.Minecraft.Modloader = "fabric"
    cfg.Paths.Mods = t.TempDir()
    ms := NewModService(cfg, zap.NewNop())
    base, _ := url.Parse(api.URL)
    ms.client.Transport = rewriteTransport{base: base}

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    info, err := ms.fetchModrinthLatestVersion(ctx, "proj")
    if err != nil { t.Fatalf("fetchModrinthLatestVersion error: %v", err) }
    if info == nil || info.Filename == "" { t.Fatalf("invalid info: %+v", info) }

    // Now test downloadMod: serve file content
    fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        _, _ = w.Write([]byte("jar-bytes"))
    }))
    defer fileSrv.Close()

    // Download new file
    updated, err := ms.downloadMod(ctx, fileSrv.URL+"/mod.jar", "mod.jar", false)
    if err != nil || !updated { t.Fatalf("downloadMod new err=%v updated=%v", err, updated) }
    // Skip when exists and not forced
    updated, err = ms.downloadMod(ctx, fileSrv.URL+"/mod.jar", "mod.jar", false)
    if err != nil || updated { t.Fatalf("downloadMod skip err=%v updated=%v", err, updated) }
    // Force overwrite
    updated, err = ms.downloadMod(ctx, fileSrv.URL+"/mod.jar", "mod.jar", true)
    if err != nil || !updated { t.Fatalf("downloadMod force err=%v updated=%v", err, updated) }
}

func TestServer_WaitForStop(t *testing.T) {
    cfg := config.DefaultConfig()
    cfg.Server.MaxStopWait = 1
    s := NewServerService(cfg, zap.NewNop())
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if err := s.waitForStop(ctx); err != nil {
        t.Fatalf("waitForStop unexpected error: %v", err)
    }
}

