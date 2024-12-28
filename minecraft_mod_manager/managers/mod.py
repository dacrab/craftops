"""Mod updates and version checking via Modrinth API."""

import asyncio
import logging
from pathlib import Path
from typing import Any, Dict, List, NotRequired, Optional, TypedDict, cast

import aiohttp
from tqdm import tqdm

from ..config.config import Config
from ..managers.notification import NotificationManager
from ..utils.constants import DEFAULT_TIMEOUT


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

class ModManager:
    """Handles mod updates and version checking via Modrinth API."""
    
    def __init__(self, config: Config, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.session: Optional[aiohttp.ClientSession] = None
        self.mods_dir = Path(config['paths']['local_mods'])
        self.notification = NotificationManager(config, logger)
    
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
    
    async def __aexit__(self, *exc_info) -> None:
        """Cleanup async context."""
        if self.session:
            await self.session.close()
    
    async def make_request(self, endpoint: str, retry_count: int = 0) -> Dict[str, Any]:
        """Make API request with retry logic."""
        if not self.session:
            raise RuntimeError("Session not initialized")
            
        try:
            await asyncio.sleep(1)  # Rate limiting delay
            async with self.session.get(endpoint) as response:
                if response.status == 429:  # Rate limited
                    if retry_count < self.config['api']['max_retries']:
                        retry_delay = self.config['api']['base_delay'] * (2 ** retry_count)
                        self.logger.warning(f"Rate limited, waiting {retry_delay} seconds...")
                        await asyncio.sleep(retry_delay)
                        return await self.make_request(endpoint, retry_count + 1)
                    max_retries = self.config['api']['max_retries']
                    raise RuntimeError(f"Rate limit exceeded after {max_retries} retries") from None
                
                if response.status != 200:
                    raise RuntimeError(f"API returned status {response.status}") from None
                
                return await response.json()
                
        except Exception as e:
            raise RuntimeError(f"Request failed: {str(e)}") from e
    
    async def fetch_latest_versions(self) -> Dict[str, ModInfo]:
        """Query Modrinth API for latest versions of all configured mods."""
        if not self.session:
            raise RuntimeError("ModManager must be used as async context manager")
        
        mod_info: Dict[str, ModInfo] = {}
        failed_mods: List[str] = []
        
        modrinth_urls = self.config.get_list('modrinth_urls')
        progress_bar = tqdm(total=len(modrinth_urls), 
                          desc="Checking mod versions")
        
        try:
            # Process mods in chunks to avoid rate limits
            chunk_size = self.config['api']['chunk_size']
            for i in range(0, len(modrinth_urls), chunk_size):
                chunk = modrinth_urls[i:i + chunk_size]
                tasks = [
                    self._fetch_mod_info(url, mod_info, failed_mods)
                    for url in chunk
                ]
                
                await asyncio.gather(*tasks)
                
                # Rate limiting delay between chunks
                if i + chunk_size < len(modrinth_urls):
                    await asyncio.sleep(2)
            
            return mod_info
            
        finally:
            progress_bar.close()
            if failed_mods:
                self._notify_failures(failed_mods)
    
    async def _fetch_mod_info(self, url: str, mod_info: Dict[str, ModInfo], 
                            failed_mods: List[str]) -> None:
        """Fetch version info for a single mod."""
        try:
            # Extract project ID from URL
            project_id = url.split('/')[-1]
            
            # Get project details
            project_data = await self.make_request(f"https://api.modrinth.com/v2/project/{project_id}")
            mod_name = project_data.get('title', project_id)
            
            # Get version list
            versions_data = await self.make_request(f"https://api.modrinth.com/v2/project/{project_id}/version")
            versions = cast(List[ModVersion], versions_data)
            
            # Filter for compatible versions
            compatible_versions = [
                v for v in versions
                if (self.config['minecraft']['version'] in v['game_versions'] and
                    any(loader.lower() == self.config['minecraft']['modloader'].lower()
                        for loader in v['loaders']))
            ]

            if compatible_versions:
                # Store latest compatible version info
                latest = compatible_versions[0]
                mod_info[url] = {
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
                    f"MC {self.config['minecraft']['version']} "
                    f"with {self.config['minecraft']['modloader']})"
                )
                
        except Exception as e:
            mod_name = project_id if 'mod_name' not in locals() else mod_name
            failed_mods.append(f"{mod_name} ({str(e)})")
            self.logger.error(f"Error processing mod {url}: {str(e)}")
    
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
                        for url, info in mod_info.items()
                    ]
                    await asyncio.gather(*tasks)
            
            finally:
                progress_bar.close()
            
            self._send_update_summary(updated_mods, skipped_mods, failed_mods, len(mod_info))
            
        except Exception as e:
            self.logger.error(f"Error updating mods: {str(e)}")
            self.notification.send_discord_notification("Mod Update Error", str(e), True)
            raise
    
    async def _update_single_mod(self, session: aiohttp.ClientSession, 
                               info: ModInfo, updated_mods: List[str], 
                               skipped_mods: List[str], failed_mods: List[str],
                               newly_added_mods: List[str]) -> None:
        """Update a single mod file."""
        try:
            mod_name = info['project_name']
            current_mod_path = self.mods_dir / info['filename']
            
            # Check if update needed
            if current_mod_path.exists():
                current_size = current_mod_path.stat().st_size
                async with session.head(info['download_url']) as response:
                    new_size = int(response.headers.get('content-length', 0))
                
                if current_size == new_size:
                    skipped_mods.append(f"• {mod_name} ({info['version_number']})")
                    return
                
                # Remove old version
                current_mod_path.unlink()
            else:
                newly_added_mods.append(f"• {mod_name} ({info['version_number']})")
            
            # Download and save new version
            async with session.get(info['download_url']) as response:
                content = await response.read()
                current_mod_path.write_bytes(content)
            
            updated_mods.append(f"• {mod_name} → {info['version_number']}")
            
        except Exception as e:
            failed_mods.append(f"• {mod_name}: {str(e)}")
    
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
    
    def _check_version_compatibility(self, version: ModVersion) -> bool:
        """Check if a mod version is compatible with current server."""
        mc_version = self.config['minecraft']['version']
        modloader = self.config['minecraft']['modloader'].lower()
        return (
            mc_version in version['game_versions'] and
            modloader in [loader.lower() for loader in version['loaders']]
        )
    
    def _format_mod_failure(self, mod_name: str) -> str:
        """Format mod failure message."""
        mc_version = self.config['minecraft']['version']
        modloader = self.config['minecraft']['modloader']
        return (
            f"{mod_name} (no compatible version for "
            f"MC {mc_version} with {modloader})"
        )
    
    def _send_update_notification(self) -> None:
        """Send notification about up-to-date mods."""
        self.notification.send_discord_notification(
            "Mod Updates",
            "✅ All mods are up to date!"
        )
    
    def _send_update_summary(self, updated_mods: List[str], skipped_mods: List[str],
                           failed_mods: List[str], total_mods: int) -> None:
        """Send summary of mod update results."""
        message = f"📊 Mod Update Summary ({total_mods} total):\n\n"
        
        if updated_mods:
            message += "✅ Updated:\n" + "\n".join(updated_mods) + "\n\n"
        if skipped_mods:
            message += "⏭️ Skipped (up to date):\n" + "\n".join(skipped_mods) + "\n\n"
        if failed_mods:
            message += "❌ Failed:\n" + "\n".join(failed_mods)
        
        self.notification.send_discord_notification("Mod Update Complete", message) 