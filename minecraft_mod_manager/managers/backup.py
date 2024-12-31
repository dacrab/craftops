"""Server backup creation and cleanup."""

import logging
import tarfile
from datetime import datetime
from pathlib import Path
from typing import List, Sequence, Tuple

from ..config.config import Config
from ..managers.notification import NotificationManager


class BackupManager:
    """Handles server backup creation and cleanup."""
    
    def __init__(self, config: Config, logger: logging.Logger) -> None:
        """Initialize backup manager."""
        self.config = config
        self.logger = logger
        self.backup_dir = Path(config.paths.backups)
        self.server_dir = Path(config.paths.server)
        self.notification = NotificationManager(config, logger)
        
        # Ensure backup directory exists
        self.backup_dir.mkdir(parents=True, exist_ok=True)
    
    def create_backup(self) -> bool:
        """Create a new backup of the server."""
        try:
            # Generate timestamp
            timestamp = datetime.now().strftime(
                self.config.backup.name_format
            )
            
            # Create backup archive
            backup_path = self.backup_dir / f"{timestamp}.tar.gz"
            
            with tarfile.open(backup_path, "w:gz") as tar:
                tar.add(self.server_dir, arcname=".")
            
            self._send_success_notification(timestamp)
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to create backup: {str(e)}")
            self.notification.send_discord_notification(
                "Backup Failed",
                f"❌ Error creating backup: {str(e)}",
                True
            )
            return False
    
    def cleanup_old_backups(self) -> None:
        """Remove old backups based on count limit."""
        try:
            # Get config values
            max_backups = self.config.backup.max_mod_backups
            
            # Get all backups with modification times
            backups = self._get_sorted_backups()
            
            # Keep only the newest max_backups
            for backup_path in backups[max_backups:]:
                self._delete_backup(backup_path, "exceeded max backups")
                        
        except Exception as e:
            self.logger.error(f"Failed to cleanup old backups: {str(e)}")
    
    def _get_sorted_backups(self) -> Sequence[Path]:
        """Get list of backups sorted by modification time (newest first)."""
        backups = list(self.backup_dir.glob("*.tar.gz"))
        return sorted(
            backups,
            key=lambda p: p.stat().st_mtime,
            reverse=True
        )
    
    def _delete_backup(self, backup_path: Path, reason: str) -> None:
        """Delete a backup file and log the action."""
        try:
            backup_path.unlink()
            self.logger.info(f"Removed old backup: {backup_path.name} (reason: {reason})")
        except PermissionError:
            self.logger.warning(f"Permission denied deleting backup: {backup_path.name}")
        except Exception as e:
            self.logger.error(f"Error deleting backup {backup_path.name}: {str(e)}")
    
    def _send_success_notification(self, timestamp: str) -> None:
        """Send backup success notification."""
        self.notification.send_discord_notification(
            "Server Backup", 
            f"✅ Created backup: {timestamp}.tar.gz"
        )