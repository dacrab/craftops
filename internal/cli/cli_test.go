package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitConfigCreatesFile verifies the init-config command creates a valid config file
func TestInitConfigCreatesFile(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "config.toml")

	cfgFile = ""
	orig := os.Args
	defer func() { os.Args = orig }()
	os.Args = []string{"craftops", "init-config", "-o", out, "--force"}

	if err := Execute(); err != nil {
		t.Fatalf("Execute(init-config) error: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("config not created: %v", err)
	}
}
