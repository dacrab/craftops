package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"craftops/internal/config"
	"craftops/internal/view"
)

func newInitCmd() *cobra.Command {
	var (
		outputPath string
		force      bool
	)

	cmd := &cobra.Command{
		Use:     "init-config",
		Aliases: []string{"init", "configure"},
		Short:   "📝 Initialize a new configuration file with defaults",
		Long: `Initialize a new configuration file with default settings.
	
This command creates a new configuration file with sensible defaults that you can
customize for your Minecraft server setup.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return initializeConfig(outputPath, force)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path for configuration file")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing configuration file")

	return cmd
}

// initializeConfig handles the configuration initialization logic
func initializeConfig(outputPath string, force bool) error {
	view.PrintBanner("Configuration Setup")

	// Determine output path
	if outputPath == "" {
		outputPath = "config.toml"
	}

	view.PrintStep(1, 4, fmt.Sprintf("Checking output path: %s", outputPath))

	// Check if file exists and force flag
	if _, err := os.Stat(outputPath); err == nil && !force {
		view.PrintWarning(fmt.Sprintf("Configuration file already exists: %s", outputPath))
		view.PrintInfo("💡 Use --force to overwrite the existing file")
		return nil
	}

	view.PrintStep(2, 4, "Creating directory structure...")
	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return handleError(err, "Failed to create directory")
	}

	view.PrintStep(3, 4, "Generating default configuration...")
	// Create default configuration
	defaultConfig := config.DefaultConfig()

	view.PrintStep(4, 4, "Saving configuration file...")
	// Save configuration
	if err := defaultConfig.SaveConfig(outputPath); err != nil {
		return handleError(err, "Failed to save configuration")
	}

	printInitSuccess(outputPath)
	return nil
}

// printInitSuccess displays the success message and next steps
func printInitSuccess(outputPath string) {
	view.PrintSection("Setup Complete!")
	view.PrintSuccess(fmt.Sprintf("📝 Configuration file created: %s", outputPath))

	view.PrintSection("Next Steps")
	view.PrintInfo("1. 📝 Edit the configuration file with your server details:")
	fmt.Printf("   %s\n", view.AccentColor.Sprintf("nano %s", outputPath))
	fmt.Println()
	view.PrintInfo("2. 🎮 Add your Modrinth mod URLs to the [mods.sources] section")
	fmt.Println()
	view.PrintInfo("3. 🏥 Run a health check to validate your setup:")
	fmt.Printf("   %s\n", view.AccentColor.Sprintf("craftops health-check"))
	fmt.Println()
	view.PrintInfo("4. 🚀 Start managing your server:")
	fmt.Printf("   %s\n", view.AccentColor.Sprintf("craftops update-mods"))
}
