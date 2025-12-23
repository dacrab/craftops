// Package cli provides the command-line interface for craftops
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd groups all lifecycle-related commands
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Minecraft server management",
}

// serverStartCmd launches the server process
var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Minecraft server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := App(cmd)
		a.Terminal.Info("Starting server...")
		if err := a.Server.Start(context.Background()); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to start server: %v", err))
			return err
		}
		a.Terminal.Success("Server is now running")
		return nil
	},
}

// serverStopCmd shuts down the server process
var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Minecraft server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := App(cmd)
		a.Terminal.Info("Stopping server...")
		if err := a.Server.Stop(context.Background()); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to stop server: %v", err))
			return err
		}
		a.Terminal.Success("Server has been stopped")
		return nil
	},
}

// serverRestartCmd triggers a full restart with warnings if configured
var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Minecraft server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, a := context.Background(), App(cmd)
		if len(a.Config.Notifications.WarningIntervals) > 0 {
			a.Terminal.Info("Sending restart warnings...")
			if err := a.Notification.SendRestartWarnings(ctx); err != nil {
				a.Terminal.Warning(fmt.Sprintf("Warning notifications failed: %v", err))
			}
		}
		a.Terminal.Info("Restarting server...")
		if err := a.Server.Restart(ctx); err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to restart: %v", err))
			_ = a.Notification.SendError(ctx, fmt.Sprintf("Server restart failed: %v", err))
			return err
		}
		a.Terminal.Success("Server restarted")
		_ = a.Notification.SendSuccess(ctx, "Server restarted successfully")
		return nil
	},
}

// serverStatusCmd reports whether the server process is alive
var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check server status",
	RunE: func(cmd *cobra.Command, _ []string) error {
		a := App(cmd)
		status, err := a.Server.Status(context.Background())
		if err != nil {
			a.Terminal.Error(fmt.Sprintf("Failed to get status: %v", err))
			return err
		}
		if status.IsRunning {
			a.Terminal.Success("Server is running")
		} else {
			a.Terminal.Error("Server is not running")
		}
		return nil
	},
}

