package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"craftops/internal/services"
)

// serverCmd represents the server command group
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Minecraft server management commands",
	Long:  `Commands for managing the Minecraft server lifecycle (start, stop, restart, status).`,
}

// serverStartCmd represents the server start command
var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Minecraft server",
	Long:  `Start the Minecraft server in a screen session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()

		serverService := services.NewServerService(cfg, logger)

		printInfo("Starting server...")
		err := serverService.Start(ctx)

		if err != nil {
			printError(fmt.Sprintf("Failed to start server: %v", err))
			return err
		}

		printSuccess("Server is now running")
		return nil
	},
}

// serverStopCmd represents the server stop command
var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Minecraft server",
	Long:  `Stop the Minecraft server gracefully.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()

		serverService := services.NewServerService(cfg, logger)

		printInfo("Stopping server...")
		err := serverService.Stop(ctx)

		if err != nil {
			printError(fmt.Sprintf("Failed to stop server: %v", err))
			return err
		}

		printSuccess("Server has been stopped")
		return nil
	},
}

// serverRestartCmd represents the server restart command
var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Minecraft server",
	Long: `Restart the Minecraft server with optional player warnings.

This command will:
• Send warning notifications to players (if configured)
• Stop the server gracefully
• Start the server again
• Send completion notification`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()

		serverService := services.NewServerService(cfg, logger)
		notificationService := services.NewNotificationService(cfg, logger)

		// Send warning notifications
		if len(cfg.Notifications.WarningIntervals) > 0 {
			printInfo("Sending restart warnings to players...")
			if err := notificationService.SendRestartWarnings(ctx); err != nil {
				printWarning(fmt.Sprintf("Failed to send warnings: %v", err))
			}
		}

		// Restart server
		printInfo("Restarting server...")
		err := serverService.Restart(ctx)

		if err != nil {
			printError(fmt.Sprintf("Failed to restart server: %v", err))
			_ = notificationService.SendErrorNotification(ctx, fmt.Sprintf("Server restart failed: %v", err))
			return err
		}

		printSuccess("Server has been restarted")
		_ = notificationService.SendSuccessNotification(ctx, "Server restarted successfully")

		return nil
	},
}

// serverStatusCmd represents the server status command
var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Minecraft server status",
	Long:  `Check the current status of the Minecraft server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := getContext()

		serverService := services.NewServerService(cfg, logger)

		status, err := serverService.GetStatus(ctx)
		if err != nil {
			printError(fmt.Sprintf("Failed to get server status: %v", err))
			return err
		}

		fmt.Println()
		fmt.Println("Server Status")
		fmt.Println("─────────────────")

		if status.IsRunning {
			printSuccess("Server is running")
		} else {
			printError("Server is not running")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Add subcommands
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
	serverCmd.AddCommand(serverRestartCmd)
	serverCmd.AddCommand(serverStatusCmd)
}
