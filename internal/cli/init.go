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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init-config",
	Short: "📝 Initialize a new configuration file with defaults",
	Long: `Initialize a new configuration file with default settings.

This command creates a new configuration file with sensible defaults that you can
customize for your Minecraft server setup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner("Configuration Setup")

		// Determine output path
		if outputPath == "" {
			outputPath = "config.toml"
		}

		printStep(1, 4, fmt.Sprintf("Checking output path: %s", outputPath))

		// Check if file exists and force flag
		if _, err := os.Stat(outputPath); err == nil && !force {
			printWarning(fmt.Sprintf("Configuration file already exists: %s", outputPath))
			printInfo("💡 Use --force to overwrite the existing file")
			return nil
		}

		printStep(2, 4, "Creating directory structure...")
		// Create directory if needed
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		printStep(3, 4, "Generating default configuration...")
		// Create default configuration
		defaultConfig := config.DefaultConfig()

		printStep(4, 4, "Saving configuration file...")
		// Save configuration
		if err := defaultConfig.SaveConfig(outputPath); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		printSection("Setup Complete!")
		printSuccess(fmt.Sprintf("📝 Configuration file created: %s", outputPath))

		printSection("Next Steps")
		printInfo("1. 📝 Edit the configuration file with your server details:")
		fmt.Printf("   %s\n", accentColor.Sprintf("nano %s", outputPath))
		fmt.Println()
		printInfo("2. 🎮 Add your Modrinth mod URLs to the [mods.sources] section")
		fmt.Println()
		printInfo("3. 🏥 Run a health check to validate your setup:")
		fmt.Printf("   %s\n", accentColor.Sprintf("craftops health-check"))
		fmt.Println()
		printInfo("4. 🚀 Start managing your server:")
		fmt.Printf("   %s\n", accentColor.Sprintf("craftops update-mods"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path for configuration file")
	initCmd.Flags().BoolVar(&force, "force", false, "overwrite existing configuration file")
}
