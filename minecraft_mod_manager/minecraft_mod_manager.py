"""Main class for Minecraft Mod Manager."""

import logging
from pathlib import Path
from typing import Optional, Union

from .config.config import Config
from .controllers.server import ServerController
from .managers.backup import BackupManager
from .managers.mod import ModManager
from .managers.notification import NotificationManager


class MinecraftModManager:
    """Main class that orchestrates server management and mod updates."""
    
    def __init__(self, config_path: Union[str, Path] = "config.jsonc"):
        """Initialize manager with configuration."""
        # Setup logging
        self.logger = logging.getLogger("MinecraftModManager")
        self._setup_logging()
        
        # Load configuration
        self.config = Config(config_path)
        
        # Initialize components
        self.server = ServerController(self.config, self.logger)
        self.backup = BackupManager(self.config, self.logger)
        self.notification = NotificationManager(self.config, self.logger)
        self.mod_manager: Optional[ModManager] = None
    
    def _setup_logging(self) -> None:
        """Configure logging with file and console handlers."""
        self.logger.setLevel(logging.INFO)
        
        # Console handler
        console = logging.StreamHandler()
        console.setLevel(logging.INFO)
        console.setFormatter(logging.Formatter(
            '%(asctime)s - %(levelname)s - %(message)s'
        ))
        self.logger.addHandler(console)
        
        # File handler (if configured)
        if hasattr(self, 'config'):
            log_path = Path(self.config['paths']['logs'])
            log_path.parent.mkdir(parents=True, exist_ok=True)
            
            file_handler = logging.FileHandler(log_path)
            file_handler.setLevel(logging.DEBUG)
            file_handler.setFormatter(logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            ))
            self.logger.addHandler(file_handler)
    
    def verify_server_status(self) -> bool:
        """Check if server is running."""
        return self.server.verify_status()
    
    def get_player_count(self) -> int:
        """Get number of online players."""
        return self.server.get_player_count()
    
    def control_server(self, action: str) -> bool:
        """Control server process (start/stop/restart)."""
        return self.server.control(action)
    
    async def run_automated_update(self) -> None:
        """Run automated update process with player warnings."""
        try:
            # Check server status
            if not self.verify_server_status():
                self.logger.error("Server must be running for automated updates")
                return
            
            # Initialize mod manager
            self.mod_manager = ModManager(self.config, self.logger)
            
            # Send initial notification
            self.notification.send_discord_notification(
                "Update Started",
                "Starting automated mod update process..."
            )
            
            # Warn players
            self.notification.warn_players()
            
            # Stop server
            if not self.control_server('stop'):
                raise RuntimeError("Failed to stop server")
            
            # Create backup
            if not self.backup.create_backup():
                raise RuntimeError("Failed to create backup")
            
            # Update mods
            async with self.mod_manager as mm:
                await mm.update_mods()
            
            # Start server
            if not self.control_server('start'):
                raise RuntimeError("Failed to start server")
            
            # Send completion notification
            self.notification.send_discord_notification(
                "Update Complete",
                "✅ Server updated and restarted successfully!"
            )
            
        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            self.notification.send_discord_notification(
                "Update Failed",
                f"❌ Error during update process: {str(e)}",
                True
            )
            raise
    
    async def run_maintenance(self) -> None:
        """Run manual maintenance process."""
        try:
            # Initialize mod manager
            self.mod_manager = ModManager(self.config, self.logger)
            
            # Send initial notification
            self.notification.send_discord_notification(
                "Maintenance Started",
                "Starting manual maintenance process..."
            )
            
            # Create backup
            if not self.backup.create_backup():
                raise RuntimeError("Failed to create backup")
            
            # Update mods
            async with self.mod_manager as mm:
                await mm.update_mods()
            
            # Cleanup old backups
            self.backup.cleanup_old_backups()
            
            # Send completion notification
            self.notification.send_discord_notification(
                "Maintenance Complete",
                "✅ Server maintenance completed successfully!"
            )
            
        except Exception as e:
            self.logger.error(f"Maintenance failed: {str(e)}")
            self.notification.send_discord_notification(
                "Maintenance Failed",
                f"❌ Error during maintenance: {str(e)}",
                True
            )
            raise 