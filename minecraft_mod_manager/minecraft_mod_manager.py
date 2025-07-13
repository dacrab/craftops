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
            if not await self.server_controller.verify_status():
                self.logger.error("Server must be running for automated updates")
                return

            # Send initial notification
            await self.notification_manager.send_discord_notification(
                "Update Started",
                "Starting automated mod update process..."
            )

            # Warn players
            await self.notification_manager.warn_players()

            # Stop server
            if not await self.server_controller.stop():
                raise RuntimeError("Failed to stop server")


def parse_args() -> argparse.Namespace:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="Minecraft Server Mod Manager",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    %(prog)s --config custom_config.toml --auto-update
    %(prog)s --maintenance
        """
    )

    parser.add_argument(
        "--config",
        type=str,
        default=DEFAULT_CONFIG_PATH,
        help="Path to configuration file (default: %(default)s)",
    )

    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument(
        "--auto-update",
        action="store_true",
        help="Run automated update process",
    )
    group.add_argument(
        "--maintenance",
        action="store_true",
        help="Run manual maintenance process",
    )

    return parser.parse_args()

async def main() -> None:
    """Main entry point."""
    args = parse_args()

    try:
        manager = MinecraftModManager(args.config)

        if args.auto_update:
            await manager.run_automated_update()
        elif args.maintenance:
            await manager.run_maintenance()

    except Exception as e:
        logging.error(f"Error: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
 RuntimeError("Failed to stop server")

            # Create backup
            if not await self.backup_manager.create_backup():
                raise RuntimeError("Failed to create backup")

            # Update mods
            await self.mod_manager.update_mods()

            # Start server
            if not await self.server_controller.start():
                raise RuntimeError("Failed to start server")

            # Send completion notification
            await self.notification_manager.send_discord_notification(
                "Update Complete",
                "✅ Server updated and restarted successfully!"
            )

        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            await self.notification_manager.send_discord_notification(
                "Update Failed",
                f"❌ Error during update process: {str(e)}",
                True
            )
            raise

    async def run_maintenance(self) -> None:
        """Run manual maintenance process."""
        try:
            # Send initial notification
            await self.notification_manager.send_discord_notification(
                "Maintenance Started",
                "Starting manual maintenance process..."
            )

            # Create backup
            if not await self.backup_manager.create_backup():
                raise RuntimeError("Failed to create backup")

            # Update mods
            await self.mod_manager.update_mods()

            # Cleanup old backups
            self.backup_manager.cleanup_old_backups()

            # Send completion notification
            await self.notification_manager.send_discord_notification(
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
