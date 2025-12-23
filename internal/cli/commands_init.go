package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"craftops/internal/config"
)

var (
	outputPath string
	force      bool
)

// initCmd scaffolds a new configuration file with default settings
var initCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Initialize a new configuration file",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return nil // Skip app initialization - config may not exist yet
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		if outputPath == "" {
			outputPath = "config.toml"
		}
		fmt.Printf("[1/4] Checking output path: %s\n", outputPath)
		if info, err := os.Stat(outputPath); err == nil && !force {
			if info.IsDir() {
				return fmt.Errorf("output path is a directory")
			}
			fmt.Printf("WARNING: Config already exists: %s\n", outputPath)
			fmt.Println("Use --force to overwrite")
			return nil
		}
		fmt.Println("[2/4] Creating directory structure...")
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		fmt.Println("[3/4] Generating default configuration...")
		cfg := config.DefaultConfig()
		fmt.Println("[4/4] Saving configuration file...")
		if err := cfg.SaveConfig(outputPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("\nSUCCESS: Configuration created: %s\n\n", outputPath)
		fmt.Println("Next steps:")
		fmt.Printf("  1. Edit the config: nano %s\n", outputPath)
		fmt.Println("  2. Add Modrinth mod URLs to [mods.modrinth_sources]")
		fmt.Println("  3. Run health check: craftops health-check")
		fmt.Println("  4. Start managing: craftops update-mods")
		return nil
	},
}

