"""Mod updates and version checking via Modrinth API."""

import asyncio
import logging
import re
import subprocess
import time
from collections.abc import AsyncIterator
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Final, List, NotRequired, TypedDict, cast
from urllib.parse import urlparse
import shutil
from datetime import datetime

import aiohttp
from tqdm import tqdm

from ..config.config import Config
from ..managers.notification import NotificationManager
from ..utils.constants import DEFAULT_TIMEOUT


@dataclass
class ModSource:
    """Internal representation of a mod source."""
    type: str
    url: str


# Type definitions
class ModFile(TypedDict):
    """Type definition for mod file data."""
    url: str
    filename: str
    size: NotRequired[int]


class ModVersion(TypedDict):
    """Type definition for mod version data."""
    id: str
    version_number: str
    game_versions: List[str]
    loaders: List[str]
    files: List[ModFile]


class ModInfo(TypedDict):
    """Type definition for mod info data."""
    version_id: str
    version_number: str
    download_url: str
    filename: str
    project_name: str


# Constants
API_BASE_URL: Final[str] = "https://api.modrinth.com/v2"
RATE_LIMIT_STATUS: Final[int] = 429


class ModManager:
    """Handles mod updates and version checking via Modrinth API."""
    
    def __init__(self, config: Config, logger: logging.Logger) -> None:
        """Initialize the mod manager."""
        self.config = config
        self.logger = logger
        self.session: aiohttp.ClientSession | None = None
        self.mods_dir = Path(config.paths.mods)
        self.notification = NotificationManager(config, logger)
        
        # Ensure mods directory exists
        self.mods_dir.mkdir(parents=True, exist_ok=True)
    
    async def __aenter__(self) -> 'ModManager':
        """Setup async context with connection pooling."""
        connector = aiohttp.TCPConnector(limit=10)
        timeout = aiohttp.ClientTimeout(total=DEFAULT_TIMEOUT)
        self.session = aiohttp.ClientSession(
            headers={"Accept": "application/json"},
            connector=connector,
            timeout=timeout
        )
        return self
    
    async def __aexit__(self, *exc_info: Any) -> None:
        """Cleanup async context."""
        if self.session:
            await self.session.close()
            self.session = None
    
    async def make_request(self, endpoint: str, retry_count: int = 0) -> Dict[str, Any]:
        """Make API request with retry logic."""
        if not self.session:
            raise RuntimeError("Session not initialized")
            
        try:
            await asyncio.sleep(1)  # Rate limiting delay
            async with self.session.get(endpoint) as response:
                if response.status == RATE_LIMIT_STATUS:
                    if retry_count < self.config.mods.max_retries:
                        retry_delay = self.config.mods.base_delay * (2 ** retry_count)
                        self.logger.warning(f"Rate limited, waiting {retry_delay} seconds...")
                        await asyncio.sleep(retry_delay)
                        return await self.make_request(endpoint, retry_count + 1)
                    raise RuntimeError(f"Rate limit exceeded after {self.config.mods.max_retries} retries")
                
                if response.status != 200:
                    raise RuntimeError(f"API returned status {response.status}")
                
                return await response.json()
                
        except Exception as e:
            raise RuntimeError(f"Request failed: {str(e)}") from e
    
    def _get_mod_sources(self) -> List[ModSource]:
        """Get all mod sources from config."""
        sources = []
        for url in self.config.mods.sources.modrinth:
            sources.append(ModSource(type="modrinth", url=url))
        for url in self.config.mods.sources.curseforge:
            sources.append(ModSource(type="curseforge", url=url))
        return sources

    async def fetch_latest_versions(self) -> Dict[str, ModInfo]:
        """Query Modrinth API for latest versions of all configured mods."""
        if not self.session:
            raise RuntimeError("ModManager must be used as async context manager")
        
        mod_info: Dict[str, ModInfo] = {}
        failed_mods: List[str] = []
        
        # Get all mod sources
        sources = self._get_mod_sources()
        
        progress_bar = tqdm(
            total=len(sources),
            desc="Checking mod versions"
        )
        
        try:
            # Process mods in chunks to avoid rate limits
            chunk_size = self.config.mods.chunk_size
            for i in range(0, len(sources), chunk_size):
                chunk = sources[i:i + chunk_size]
                tasks = [
                    self._fetch_mod_info(source, mod_info, failed_mods)
                    for source in chunk
                ]
                
                await asyncio.gather(*tasks)
                progress_bar.update(len(chunk))
                
                # Rate limiting delay between chunks
                if i + chunk_size < len(sources):
                    await asyncio.sleep(2)
            
            return mod_info
            
        finally:
            progress_bar.close()
            if failed_mods:
                self._notify_failures(failed_mods)
    
    def _extract_project_id(self, source: ModSource) -> str:
        """Extract project ID from mod URL."""
        parsed_url = urlparse(source.url)
        
        if source.type == "modrinth":
            # Extract from modrinth.com/mod/project-id
            match = re.search(r'/mod/([^/]+)', parsed_url.path)
            if not match:
                raise ValueError(f"Invalid Modrinth URL format: {source.url}")
            return match.group(1)
            
        elif source.type == "curseforge":
            # Extract from curseforge.com/minecraft/mc-mods/project-id
            match = re.search(r'/mc-mods/([^/]+)', parsed_url.path)
            if not match:
                raise ValueError(f"Invalid CurseForge URL format: {source.url}")
            return match.group(1)
            
        else:
            raise ValueError(f"Unsupported mod source type: {source.type}")

    async def _fetch_mod_info(
        self,
        source: ModSource,
        mod_info: Dict[str, ModInfo],
        failed_mods: List[str]
    ) -> None:
        """Fetch version info for a single mod."""
        try:
            project_id = self._extract_project_id(source)
            
            # Get project details
            project_data = await self.make_request(f"{API_BASE_URL}/project/{project_id}")
            mod_name = project_data.get('title', project_id)
            
            # Get version list
            versions_data = await self.make_request(f"{API_BASE_URL}/project/{project_id}/version")
            versions = cast(List[ModVersion], versions_data)
            
            # Filter for compatible versions
            compatible_versions = [
                v for v in versions
                if (self.config.minecraft.version in v['game_versions'] and
                    any(loader.lower() == self.config.minecraft.modloader.lower()
                        for loader in v['loaders']))
            ]

            if compatible_versions:
                # Store latest compatible version info
                latest = compatible_versions[0]
                mod_info[project_id] = {
                    'version_id': latest['id'],
                    'version_number': latest['version_number'],
                    'download_url': latest['files'][0]['url'],
                    'filename': latest['files'][0]['filename'],
                    'project_name': mod_name
                }
            else:
                self._log_compatibility_warning(mod_name, versions)
                failed_mods.append(
                    f"{mod_name} (no compatible version for "
                    f"MC {self.config.minecraft.version} "
                    f"with {self.config.minecraft.modloader})"
                )
                
        except Exception as e:
            mod_name = source.url.split('/')[-1]
            failed_mods.append(f"{mod_name} ({str(e)})")
            self.logger.error(f"Error processing mod {source.url}: {str(e)}")
    
    async def update_mods(self) -> None:
        """Update all mods to their latest compatible versions."""
        try:
            mod_info = await self.fetch_latest_versions()
            if not mod_info:
                self.notification.send_discord_notification(
                    "Mod Updates", 
                    "✅ All mods are up to date!"
                )
                return
            
            # Track update status
            updated_mods: List[str] = []
            skipped_mods: List[str] = []
            failed_mods: List[str] = []
            newly_added_mods: List[str] = []
            
            progress_bar = tqdm(total=len(mod_info), desc="Downloading mods")
            
            try:
                # Download mods in parallel
                async with aiohttp.ClientSession() as session:
                    tasks = [
                        self._update_single_mod(
                            session, info,
                            updated_mods, skipped_mods, failed_mods, 
                            newly_added_mods
                        )
                        for info in mod_info.values()
                    ]
                    await asyncio.gather(*tasks)
                    progress_bar.update(len(tasks))
            
            finally:
                progress_bar.close()
            
            self._send_update_summary(updated_mods, skipped_mods, failed_mods, len(mod_info))
            
        except Exception as e:
            self.logger.error(f"Error updating mods: {str(e)}")
            self.notification.send_discord_notification("Mod Update Error", str(e), True)
            raise
    
    def _get_mod_backup_dir(self, mod_name: str) -> Path:
        """Get the backup directory for a specific mod."""
        backup_base = Path(self.config.paths.backups) / self.config.backup.mod_backup_dir
        mod_backup_dir = backup_base / mod_name
        mod_backup_dir.mkdir(parents=True, exist_ok=True)
        return mod_backup_dir

    def _backup_mod(self, mod_path: Path, mod_name: str) -> Path | None:
        """Backup a mod file before updating.
        
        Returns:
            Path to the backup file, or None if the source file doesn't exist.
        """
        if not mod_path.exists():
            return None

        # Create backup directory
        backup_dir = self._get_mod_backup_dir(mod_name)
        
        # Generate backup filename with timestamp
        timestamp = datetime.now().strftime(self.config.backup.name_format)
        backup_path = backup_dir / f"{mod_path.stem}_{timestamp}{mod_path.suffix}"
        
        # Copy the mod file to backup
        shutil.copy2(mod_path, backup_path)
        
        # Clean up old backups
        self._cleanup_old_backups(backup_dir)
        
        return backup_path

    def _cleanup_old_backups(self, backup_dir: Path) -> None:
        """Remove old backups exceeding the max_mod_backups limit."""
        backups = sorted(backup_dir.glob("*"), key=lambda x: x.stat().st_mtime, reverse=True)
        for old_backup in backups[self.config.backup.max_mod_backups:]:
            old_backup.unlink()

    def _rollback_mod(self, mod_name: str) -> bool:
        """Rollback a mod to its previous version.
        
        Returns True if rollback was successful.
        """
        backup_dir = self._get_mod_backup_dir(mod_name)
        backups = sorted(backup_dir.glob("*"), key=lambda x: x.stat().st_mtime, reverse=True)
        
        if not backups:
            self.logger.warning(f"No backup found for mod: {mod_name}")
            return False
            
        # Get latest backup
        latest_backup = backups[0]
        mod_path = self.mods_dir / latest_backup.name.split("_")[0]
        
        # Remove current version and restore backup
        if mod_path.exists():
            mod_path.unlink()
        shutil.copy2(latest_backup, mod_path)
        
        return True

    async def _verify_server_startup(self) -> bool:
        """Attempt to start the server and verify it works.
        
        Returns True if server starts successfully, False otherwise.
        """
        try:
            # Start the server
            start_cmd = self.config.server.start_command
            process = subprocess.Popen(
                start_cmd.split(),
                cwd=Path(self.config.paths.server),
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            # Wait for server to start (max 60 seconds)
            start_time = time.time()
            while time.time() - start_time < 60:
                if process.poll() is not None:
                    # Server crashed or failed to start
                    stdout, stderr = process.communicate()
                    self.logger.error(f"Server failed to start:\n{stderr}")
                    return False
                
                # Check if server is responding (you might want to implement a more robust check)
                if "Done" in (process.stdout.readline() if process.stdout else ""):
                    break
                    
                await asyncio.sleep(1)
            
            # Stop the server gracefully
            process.terminate()
            process.wait(timeout=30)
            return True
            
        except Exception as e:
            self.logger.error(f"Error verifying server startup: {str(e)}")
            return False
        finally:
            # Make sure process is killed if it's still running
            if process and process.poll() is None:
                process.kill()

    async def _update_single_mod(
        self,
        session: aiohttp.ClientSession,
        info: ModInfo,
        updated_mods: List[str],
        skipped_mods: List[str],
        failed_mods: List[str],
        newly_added_mods: List[str]
    ) -> None:
        """Update a single mod."""
        try:
            mod_path = self.mods_dir / info['filename']
            
            # Step 1: Backup existing mod if it exists
            if mod_path.exists():
                backup_path = self._backup_mod(mod_path, info['project_name'])
                if not backup_path:
                    failed_mods.append(f"{info['project_name']} (backup failed)")
                    return
            
            # Step 2: Download and save the new version
            try:
                async with session.get(info['download_url']) as response:
                    if response.status != 200:
                        raise RuntimeError(f"Download failed with status {response.status}")
                    
                    content = await response.read()
                    mod_path.write_bytes(content)
                    
                # Step 3: Verify server startup
                if not await self._verify_server_startup():
                    self.logger.warning(f"Server failed to start with new mod version: {info['project_name']}")
                    
                    # Rollback to previous version
                    if self._rollback_mod(info['project_name']):
                        failed_mods.append(
                            f"{info['project_name']} (incompatible - rolled back to previous version)"
                        )
                    else:
                        failed_mods.append(
                            f"{info['project_name']} (incompatible - rollback failed)"
                        )
                    return
                    
                # Update successful
                if not mod_path.exists():
                    newly_added_mods.append(info['project_name'])
                else:
                    updated_mods.append(info['project_name'])
                        
            except Exception as e:
                # If download or save fails, attempt rollback
                if self._rollback_mod(info['project_name']):
                    failed_mods.append(f"{info['project_name']} (update failed, rolled back)")
                else:
                    failed_mods.append(f"{info['project_name']} (update and rollback failed)")
                raise
                
        except Exception as e:
            self.logger.error(f"Error updating {info['project_name']}: {str(e)}")
            if info['project_name'] not in failed_mods:
                failed_mods.append(f"{info['project_name']} ({str(e)})")

    def _notify_failures(self, failed_mods: List[str]) -> None:
        """Send notification about failed mod updates."""
        message = "Failed to process the following mods:\n" + "\n".join(failed_mods)
        self.notification.send_discord_notification("Mod Update Issues", message, True)
    
    def _log_compatibility_warning(self, mod_name: str, versions: List[ModVersion]) -> None:
        """Log warning about mod version compatibility."""
        version_list = ', '.join(v['version_number'] for v in versions[:5])
        self.logger.warning(
            f"No compatible version found for {mod_name}. "
            f"Available versions: {version_list}"
        )
    
    def _send_update_summary(
        self,
        updated_mods: List[str],
        skipped_mods: List[str],
        failed_mods: List[str],
        total_mods: int
    ) -> None:
        """Send update summary notification."""
        if not any([updated_mods, skipped_mods, failed_mods]):
            return
            
        summary = ["📦 Mod Update Summary"]
        
        if updated_mods:
            summary.append("\n✅ Updated:")
            summary.extend(f"  • {mod}" for mod in updated_mods)
            
        if skipped_mods:
            summary.append("\n⏭️ Skipped:")
            summary.extend(f"  • {mod}" for mod in skipped_mods)
            
        if failed_mods:
            summary.append("\n❌ Failed:")
            summary.extend(f"  • {mod}" for mod in failed_mods)
        
        self.notification.send_discord_notification(
            "Mod Updates",
            "\n".join(summary)
        )