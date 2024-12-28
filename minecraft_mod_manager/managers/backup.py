"""Server backup creation and cleanup."""

import logging
import os
import tarfile
import time
from datetime import datetime
from pathlib import Path
from typing import List, Optional, Tuple, Union

from ..config.config import Config
from ..managers.notification import NotificationManager

class BackupManager:
    """Handles server backup creation and cleanup."""
    
    def __init__(self, config: Config, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.minecraft_dir = Path(config['paths']['minecraft'])
        self.backup_dir = Path(config['paths']['backups'])
        self.mods_dir = Path(config['paths']['local_mods'])
        self.notification = NotificationManager(config, logger)
    
    def create_backup(self) -> bool:
        """Create server backup with mods, config, and world data."""
        try:
            # Generate backup name with timestamp
            timestamp = datetime.now().strftime(self.config['maintenance']['backup_name_format'])
            backup_path = self.backup_dir / f"{timestamp}.tar.gz"
            
            # Setup backup directory
            self._ensure_backup_permissions()
            
            # Create temp directory
            temp_backup_dir = self.backup_dir / "temp_backup"
            self._setup_temp_dir(temp_backup_dir)
            
            try:
                # Copy files to backup
                self._copy_backup_files(temp_backup_dir)
                
                # Create archive
                self._create_archive(backup_path, temp_backup_dir)
                
                self.logger.info(f"Created backup at {backup_path}")
                self._send_success_notification(timestamp)
                return True
                
            finally:
                # Cleanup temp files
                self._cleanup_temp_dir(temp_backup_dir)
            
        except Exception as e:
            self.logger.error(f"Backup error: {str(e)}")
            self._send_error_notification(str(e))
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
    
    def _ensure_backup_permissions(self) -> None:
        """Setup backup directory with correct permissions."""
        self.backup_dir.mkdir(parents=True, exist_ok=True)
        os.system(f"sudo chown -R $USER:$USER {self.backup_dir}")
        os.system(f"sudo chmod -R a+r {self.minecraft_dir}")
    
    def _setup_temp_dir(self, temp_dir: Path) -> None:
        """Create and setup temporary backup directory."""
        os.system(f"sudo rm -rf {temp_dir}")
        os.system(f"sudo mkdir -p {temp_dir}")
        os.system(f"sudo chown -R $USER:$USER {temp_dir}")
    
    def _copy_backup_files(self, temp_dir: Path) -> None:
        """Copy server files to temporary backup directory."""
        os.system(f"sudo cp -r {self.mods_dir} {temp_dir}/mods")
        os.system(f"sudo cp -r {self.minecraft_dir}/config {temp_dir}/config")
        os.system(f"sudo cp -r {self.minecraft_dir}/world {temp_dir}/world")
        os.system(f"sudo chown -R $USER:$USER {temp_dir}")
    
    def _create_archive(self, backup_path: Path, temp_dir: Path) -> None:
        """Create compressed archive of backup files."""
        with tarfile.open(backup_path, 'w:gz') as tar:
            tar.add(temp_dir / 'mods', arcname='mods')
            tar.add(temp_dir / 'config', arcname='config')
            tar.add(temp_dir / 'world', arcname='world')
    
    def _cleanup_temp_dir(self, temp_dir: Path) -> None:
        """Remove temporary backup directory."""
        os.system(f"sudo rm -rf {temp_dir}")
    
    def _get_backup_pattern(self) -> str:
        """Get glob pattern for matching backup files."""
        format_str = self.config['maintenance']['backup_name_format']
        return f"*{format_str.replace('%Y','*').replace('%m','*').replace('%d','*').replace('%H','*').replace('%M','*')}.tar.gz"
    
    def _get_sorted_backups(self, pattern: str) -> List[Tuple[Path, float]]:
        """Get list of backups sorted by modification time (newest first)."""
        backups = []
        for backup in self.backup_dir.glob(pattern):
            mtime = backup.stat().st_mtime
            backups.append((backup, mtime))
        backups.sort(key=lambda x: x[1], reverse=True)
        return backups
    
    def _delete_backup(self, backup: Path, reason: Optional[str]) -> None:
        """Delete a backup file and log the action."""
        if reason is None:
            reason = "unknown"
        os.system(f"sudo rm {backup}")
        self.logger.info(f"Removed old backup: {backup.name} (reason: {reason})")
    
    def _send_success_notification(self, timestamp: str) -> None:
        """Send backup success notification."""
        self.notification.send_discord_notification(
            "Server Backup",
            f"✅ Created backup: {timestamp}.tar.gz"
        )
    
    def _send_error_notification(self, error: str) -> None:
        """Send backup error notification."""
        self.notification.send_discord_notification(
            "Backup Failed",
            f"❌ Backup failed: {error}",
            True
        ) 