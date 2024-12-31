"""Main application module."""

import logging
from typing import Final, Optional

from .config.config import load_config
from .controllers.server import ServerController, ServerControllerProtocol
from .managers import BackupManager, ModManager, NotificationManager

logger = logging.getLogger(__name__)

DEFAULT_CONFIG_FILE: Final[str] = "config.toml"

class MinecraftModManager:
    """Main application class that orchestrates all the functionality."""

    def __init__(self, config_path: Optional[str] = None):
        """
        Initialize the Minecraft Mod Manager.

        Args:
            config_path: Optional path to the configuration file
        """
        self.logger = logging.getLogger(f"{__name__}.MinecraftModManager")
        self.config = load_config(config_path or DEFAULT_CONFIG_FILE)
        self.mod_manager = ModManager(self.config, self.logger)
        self.backup_manager = BackupManager(self.config, self.logger)
        self.notification_manager = NotificationManager(self.config, self.logger)
        self.server_controller: ServerControllerProtocol = (
            ServerController(self.config, self.logger)
        )

    async def run_automated_update(self) -> None:
        """Run automated update process with player warnings."""
        try:
            # Check server status
            if not self.server_controller.verify_status():
                self.logger.error("Server must be running for automated updates")
                return

            # Send initial notification
            self.notification_manager.send_discord_notification(
                "Update Started",
                "Starting automated mod update process..."
            )

            # Warn players
            self.notification_manager.warn_players()

            # Stop server
            if not self.server_controller.stop():
                raise RuntimeError("Failed to stop server")

            # Create backup
            if not self.backup_manager.create_backup():
                raise RuntimeError("Failed to create backup")

            # Update mods
            await self.mod_manager.update_mods()

            # Start server
            if not self.server_controller.start():
                raise RuntimeError("Failed to start server")

            # Send completion notification
            self.notification_manager.send_discord_notification(
                "Update Complete",
                "✅ Server updated and restarted successfully!"
            )

        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            self.notification_manager.send_discord_notification(
                "Update Failed",
                f"❌ Error during update process: {str(e)}",
                True
            )
            raise

    async def run_maintenance(self) -> None:
        """Run manual maintenance process."""
        try:
            # Send initial notification
            self.notification_manager.send_discord_notification(
                "Maintenance Started",
                "Starting manual maintenance process..."
            )

            # Create backup
            if not self.backup_manager.create_backup():
                raise RuntimeError("Failed to create backup")

            # Update mods
            await self.mod_manager.update_mods()

            # Cleanup old backups
            self.backup_manager.cleanup_old_backups()

            # Send completion notification
            self.notification_manager.send_discord_notification(
                "Maintenance Complete",
                "✅ Server maintenance completed successfully!"
            )

        except Exception as e:
            self.logger.error(f"Maintenance failed: {str(e)}")
            self.notification_manager.send_discord_notification(
                "Maintenance Failed",
                f"❌ Error during maintenance: {str(e)}",
                True
            )
            raise
