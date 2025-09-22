package cli

import (
	"context"

	"github.com/spf13/cobra"

	"craftops/internal/view"
)

func newServerCmd(factory ServiceFactory) *cobra.Command {
	handler := NewCommandHandler(factory)

	serverCmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"srv"},
		Short:   "🎮 Minecraft server management commands",
		Long:    `Commands for managing the Minecraft server lifecycle (start, stop, restart, status).`,
	}

	// Start server command
	startCmd := createSimpleCommand("start", "🚀 Start the Minecraft server", nil, func() error {
		view.PrintInfo("Starting server...")
		return handler.ExecuteWithSpinner("Starting server...", func(ctx context.Context) error {
			if err := factory.GetServerService().Start(ctx); err != nil {
				return handleError(err, "Failed to start server")
			}
			view.PrintSuccess("Server is now running")
			return nil
		})
	})

	// Stop server command
	stopCmd := createSimpleCommand("stop", "🛑 Stop the Minecraft server", nil, func() error {
		view.PrintInfo("Stopping server...")
		return handler.ExecuteWithSpinner("Stopping server...", func(ctx context.Context) error {
			if err := factory.GetServerService().Stop(ctx); err != nil {
				return handleError(err, "Failed to stop server")
			}
			view.PrintSuccess("Server has been stopped")
			return nil
		})
	})

	// Restart server command
	restartCmd := createSimpleCommand("restart", "🔄 Restart the Minecraft server", nil, func() error {
		return handler.ExecuteWithContext(func(ctx context.Context) error {
			serverService := factory.GetServerService()
			notificationService := factory.GetNotificationService()

			// Send warnings if configured
			if len(factory.GetConfig().Notifications.WarningIntervals) > 0 {
				view.PrintInfo("Sending restart warnings to players...")
				if err := notificationService.SendRestartWarnings(ctx); err != nil {
					view.PrintWarning("Failed to send warnings: " + err.Error())
				}
			}

			view.PrintInfo("Restarting server...")
			return runWithSpinner("Restarting server...", func() error {
				if err := serverService.Restart(ctx); err != nil {
					_ = notificationService.SendErrorNotification(ctx, "Server restart failed: "+err.Error())
					return handleError(err, "Failed to restart server")
				}
				view.PrintSuccess("Server has been restarted")
				_ = notificationService.SendSuccessNotification(ctx, "Server restarted successfully")
				return nil
			})
		})
	})

	// Status command
	statusCmd := createSimpleCommand("status", "📊 Check Minecraft server status", nil, func() error {
		return handler.ExecuteWithContext(func(ctx context.Context) error {
			status, err := factory.GetServerService().GetStatus(ctx)
			if err != nil {
				return handleError(err, "Failed to get server status")
			}
			printServerStatus(status)
			return nil
		})
	})

	serverCmd.AddCommand(startCmd, stopCmd, restartCmd, statusCmd)
	return serverCmd
}
