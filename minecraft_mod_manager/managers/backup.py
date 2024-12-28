"""Server backup creation and cleanup."""

from datetime import datetime
import logging
from pathlib import Path
import tarfile
import time
from typing import List, Optional, Tuple

from ..config.config import Config
from ..managers.notification import NotificationManager

class BackupManager:
    """Handles server backup creation and cleanup."""
    
    def __init__(self, config: Config, logger: logging.Logger):
        """Initialize backup manager."""
        self.config = config
        self.logger = logger
        self.backup_dir = Path(config['paths']['backups'])
        self.minecraft_dir = Path(config['paths']['minecraft'])
        self.notification = NotificationManager(config, logger)
        
        # Ensure backup directory exists
        self.backup_dir.mkdir(parents=True, exist_ok=True)
    
    def create_backup(self) -> bool:
        """Create a new backup of the server."""
        try:
            # Generate timestamp
            timestamp = datetime.now().strftime(
                self.config['maintenance']['backup_name_format']
            )
            
            # Create backup archive
            backup_path = self.backup_dir / f"{timestamp}.tar.gz"
            
            with tarfile.open(backup_path, "w:gz") as tar:
                tar.add(self.minecraft_dir, arcname=".")
            
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
        """Remove old backups based on age and count limits."""
        try:
            # Get config values
            keep_days = self.config['maintenance']['backup_retention_days']
            max_backups = self.config['maintenance']['max_backups']
            now = time.time()
            
            # Match backup filename pattern
            backup_pattern = self._get_backup_pattern()
            
            # Get all backups with modification times
            backups = self._get_sorted_backups(backup_pattern)
            
            # Process each backup
            for idx, (backup, mtime) in enumerate(backups):
                should_delete = False
                delete_reason = None
                
                # Check age
                if mtime < now - (keep_days * 86400):
                    should_delete = True
                    delete_reason = "age"
                
                # Check count
                if idx >= max_backups:
                    should_delete = True
                    delete_reason = "count"
                
                if should_delete:
                    self._delete_backup(backup, delete_reason)
                        
        except Exception as e:
            self.logger.error(f"Failed to cleanup old backups: {str(e)}")
    
    def _get_backup_pattern(self) -> str:
        """Get glob pattern for matching backup files."""
        format_str = self.config['maintenance']['backup_name_format']
        replacements = {
            '%Y': '*', '%m': '*', '%d': '*',
            '%H': '*', '%M': '*'
        }
        for old, new in replacements.items():
            format_str = format_str.replace(old, new)
        return f"*{format_str}.tar.gz"
    
    def _get_sorted_backups(self, pattern: str) -> List[Tuple[Path, float]]:
        """Get list of backups sorted by modification time (newest first)."""
        backups = [
            (backup, backup.stat().st_mtime)
            for backup in self.backup_dir.glob(pattern)
        ]
        return sorted(backups, key=lambda x: x[1], reverse=True)
    
    def _delete_backup(self, backup: Path, reason: Optional[str]) -> None:
        """Delete a backup file and log the action."""
        try:
            backup.unlink()
            self.logger.info(f"Removed old backup: {backup.name} (reason: {reason or 'unknown'})")
        except PermissionError:
            self.logger.warning(f"Permission denied deleting backup: {backup.name}")
    
    def _send_success_notification(self, timestamp: str) -> None:
        """Send backup success notification."""
        self.notification.send_discord_notification(
            "Server Backup",
            f"✅ Created backup: {timestamp}.tar.gz"
        ) 