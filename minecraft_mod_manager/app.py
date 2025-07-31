"""Main application module."""

import argparse
import asyncio
import logging
import sys
from pathlib import Path
from typing import Final, Optional

from .config.config import load_config
from .core import BackupManager, ModManager, NotificationManager, ServerController

logger = logging.getLogger(__name__)

# Default config path
CONFIG_DIR = Path.home() / ".config" / "minecraft-mod-manager"
DEFAULT_CONFIG_PATH: Final[str] = str(CONFIG_DIR / "config.toml")
DEFAULT_CONFIG_FILE: Final[str] = "config.toml"


class MinecraftModManager:
    """Main application class that orchestrates all the functionality."""

    def __init__(self, config_path: Optional[str] = None) -> None:
        """Initialize the Minecraft Mod Manager."""
        self.config = load_config(config_path or DEFAULT_CONFIG_FILE)
        self.logger = logging.getLogger(f"{__name__}.MinecraftModManager")

        # Initialize components
        self.mod_manager = ModManager(self.config, self.logger)
        self.backup_manager = BackupManager(self.config, self.logger)
        self.notification_manager = NotificationManager(self.config, self.logger)
        self.server_controller = ServerController(self.config, self.logger)

    async def run_automated_update(self) -> None:
        """Run automated update process with player warnings."""
        try:
            # Check server status
            if not await self.server_controller.verify_status():
                self.logger.error("Server must be running for automated updates")
                return

            # Send initial notification
            await self.notification_manager.send_discord_notification(
                "Update Started", "Starting automated mod update process..."
            )

            # Warn players
            await self.notification_manager.warn_players()

            # Stop server
            if not await self.server_controller.stop():
                raise RuntimeError("Failed to stop server")

            # Create backup
            if not await self.backup_manager.create_backup():
                raise RuntimeError("Failed to create backup")

            # Update mods
            async with self.mod_manager:
                await self.mod_manager.update_mods()

            # Start server
            if not await self.server_controller.start():
                raise RuntimeError("Failed to start server")

            # Send completion notification
            await self.notification_manager.send_discord_notification(
                "Update Complete", "✅ Server updated and restarted successfully!"
            )

        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            await self.notification_manager.send_discord_notification(
                "Update Failed", f"❌ Error during update process: {str(e)}", True
            )
            raise

    async def run_maintenance(self) -> None:
        """Run manual maintenance process."""
        try:
            # Send initial notification
            await self.notification_manager.send_discord_notification(
                "Maintenance Started", "Starting manual maintenance process..."
            )

            # Create backup
            if not await self.backup_manager.create_backup():
                raise RuntimeError("Failed to create backup")

            # Update mods
            async with self.mod_manager:
                await self.mod_manager.update_mods()

            # Cleanup old backups
            self.backup_manager.cleanup_old_backups()

            # Send completion notification
            await self.notification_manager.send_discord_notification(
                "Maintenance Complete", "✅ Server maintenance completed successfully!"
            )

        except Exception as e:
            self.logger.error(f"Maintenance failed: {str(e)}")
            await self.notification_manager.send_discord_notification(
                "Maintenance Failed", f"❌ Error during maintenance: {str(e)}", True
            )
            raise

    async def run_health_check(self) -> bool:
        """Run basic health checks."""
        try:
            self.logger.info("Running health checks...")

            # Check filesystem access
            paths_to_check = [
                Path(self.config.server_path),
                Path(self.config.mods_path),
                Path(self.config.backups_path),
            ]

            for path in paths_to_check:
                if not path.exists():
                    self.logger.error(f"Path does not exist: {path}")
                    return False

                # Test write permissions
                test_file = path / '.health_check_test'
                try:
                    test_file.touch()
                    test_file.unlink()
                except Exception as e:
                    self.logger.error(f"No write permission for {path}: {e}")
                    return False

            # Check server jar
            jar_path = Path(self.config.server_path) / self.config.server_jar
            if not jar_path.exists():
                self.logger.error(f"Server jar not found: {jar_path}")
                return False

            # Test network connectivity
            try:
                import aiohttp
                async with aiohttp.ClientSession() as session:
                    timeout = aiohttp.ClientTimeout(total=10)
                    url = 'https://api.modrinth.com/v2/'
                    async with session.get(url, timeout=timeout) as response:
                        if response.status != 200:
                            self.logger.error("Cannot connect to Modrinth API")
                            return False
            except Exception as e:
                self.logger.error(f"Network connectivity check failed: {e}")
                return False

            self.logger.info("All health checks passed")
            return True

        except Exception as e:
            self.logger.error(f"Health check failed: {str(e)}")
            return False

    def run_cleanup(self) -> None:
        """Run system cleanup operations."""
        try:
            self.logger.info("Starting system cleanup...")

            cleaned_items = 0

            # Cleanup old backups
            backup_dir = Path(self.config.backups_path)
            if backup_dir.exists():
                backups = list(backup_dir.glob("*.tar.gz"))
                backups.sort(key=lambda x: x.stat().st_mtime, reverse=True)

                for backup in backups[self.config.max_backups:]:
                    backup.unlink()
                    cleaned_items += 1
                    self.logger.info(f"Removed old backup: {backup.name}")

            # Cleanup temporary files
            temp_patterns = ["*.tmp", "*.temp", "*~", ".DS_Store", "Thumbs.db"]
            directories_to_clean = [
                Path(self.config.server_path),
                Path(self.config.mods_path),
                Path(self.config.backups_path),
            ]

            for directory in directories_to_clean:
                if not directory.exists():
                    continue

                for pattern in temp_patterns:
                    for temp_file in directory.glob(pattern):
                        if temp_file.is_file():
                            temp_file.unlink()
                            cleaned_items += 1

            self.logger.info(f"System cleanup completed. Removed {cleaned_items} items.")

        except Exception as e:
            self.logger.error(f"Cleanup failed: {str(e)}")


def parse_args() -> argparse.Namespace:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="Minecraft Server Mod Manager",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    %(prog)s --config custom_config.toml --auto-update
    %(prog)s --maintenance
    %(prog)s --health-check
    %(prog)s --cleanup
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
    group.add_argument(
        "--health-check",
        action="store_true",
        help="Run system health checks",
    )
    group.add_argument(
        "--cleanup",
        action="store_true",
        help="Run system cleanup operations",
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
        elif args.health_check:
            healthy = await manager.run_health_check()
            if not healthy:
                sys.exit(1)
        elif args.cleanup:
            manager.run_cleanup()

    except Exception as e:
        logging.error(f"Error: {str(e)}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
