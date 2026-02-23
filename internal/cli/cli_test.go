package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// resetGlobals resets all global CLI state between tests.
// Cobra registers flags as globals, so tests must restore them to avoid bleed-through.
func resetGlobals(t *testing.T) {
	t.Helper()
	origArgs := os.Args
	origCfgFile := cfgFile
	origOutputPath := outputPath
	origForce := force
	origDebug := debug
	origDryRun := dryRun
	t.Cleanup(func() {
		os.Args = origArgs
		cfgFile = origCfgFile
		outputPath = origOutputPath
		force = origForce
		debug = origDebug
		dryRun = origDryRun
	})
}

func TestInitConfig_CreatesFile(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "config.toml")

	cfgFile = ""
	outputPath = out
	force = false
	os.Args = []string{"craftops", "init-config", "-o", out}

	if err := Execute(context.Background()); err != nil {
		t.Fatalf("Execute(init-config) error: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestInitConfig_ForceOverwrite(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "config.toml")

	// Write a sentinel value first
	_ = os.WriteFile(out, []byte("sentinel"), 0o600)

	cfgFile = ""
	outputPath = out
	force = true
	os.Args = []string{"craftops", "init-config", "-o", out, "--force"}

	if err := Execute(context.Background()); err != nil {
		t.Fatalf("Execute(init-config --force) error: %v", err)
	}
	data, err := os.ReadFile(out) //nolint:gosec
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}
	if string(data) == "sentinel" {
		t.Error("--force should have overwritten the existing file")
	}
}

func TestInitConfig_NoForce_ExistingFile(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "config.toml")

	// Pre-create the file
	_ = os.WriteFile(out, []byte("original"), 0o600)

	cfgFile = ""
	outputPath = out
	force = false
	os.Args = []string{"craftops", "init-config", "-o", out}

	if err := Execute(context.Background()); err != nil {
		t.Fatalf("Execute should not error when file exists without --force: %v", err)
	}
	// File should be unchanged
	data, _ := os.ReadFile(out) //nolint:gosec
	if string(data) != "original" {
		t.Error("file should not be overwritten without --force")
	}
}

func TestInitConfig_OutputIsDirectory(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir()

	cfgFile = ""
	outputPath = tmp // directory, not a file
	force = true
	os.Args = []string{"craftops", "init-config", "-o", tmp, "--force"}

	if err := Execute(context.Background()); err == nil {
		t.Error("expected error when output path is a directory")
	}
}

func TestInitConfig_DefaultPath(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Skipf("cannot chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	cfgFile = ""
	outputPath = "" // let it default to "config.toml"
	force = false
	os.Args = []string{"craftops", "init-config"}

	if err := Execute(context.Background()); err != nil {
		t.Fatalf("Execute(init-config) with default path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "config.toml")); err != nil {
		t.Fatalf("default config.toml not created: %v", err)
	}
}
