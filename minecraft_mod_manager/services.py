"""Core services for Minecraft Mod Manager."""

import asyncio
import logging
import re
import shutil
import tarfile
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, Final, List, Optional, TypedDict, cast
from urllib.parse import urlparse

import aiohttp
from tqdm import tqdm
from typing_extensions import NotRequired

# API Constants
MODRINTH_API_URL: Final[str] = "https://api.modrinth.com/v2"
CURSEFORGE_API_URL: Final[str] = "https://api.curseforge.com/v1"
CURSEFORGE_API_KEY: Final[str] = "$2a$10$bL4bIL5pUWqfcO7KQtnMReakwtfHbNKh6v1uTpKlzhwoueEJQnPnm"  # Public key
RATE_LIMIT_STATUS: Final[int] = 429
DEFAULT_TIMEOUT: Final[int] = 30

# Discord Constants
DISCORD_MAX_LENGTH: Final[int] = 2000
DISCORD_SUCCESS_COLOR: Final[int] = 0x00FF00  # Green
DISCORD_ERROR_COLOR: Final[int] = 0xFF0000    # Red
DISCORD_FOOTER_TEXT: Final[str] = "Minecraft Mod Manager"

# Minecraft Constants
MINECRAFT_GAME_ID: Final[int] = 432  # CurseForge game ID for Minecraft

logger = logging.getLogger(__name__)


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


@dataclass
class ModSource:
    """Internal representation of a mod source."""
    type: str
    url: str


class MinecraftModManagerError(Exception):
    """Base exception for all Minecraft Mod Manager errors."""
    pass


class ModUpdateError(MinecraftModManagerError):
    """Raised when there's an issue with mod updates."""
    pass


class BackupError(MinecraftModManagerError):
    """Raised when there's an issue with backup operations."""
    pass


class ServerError(MinecraftModManagerError):
    """Raised when there's an issue with server operations."""
    pass


class NetworkError(MinecraftModManagerError):
    """Raised when there are network connectivity issues."""
    pass


class RateLimitError(ModUpdateError):
    """Raised when API rate limits are exceeded."""
    pass


async def retry_with_backoff(
    operation: Any,
    max_attempts: int = 3,
    initial_delay: float = 1.0,
    exception_types: tuple = (Exception,),
    *args: Any,
    **kwargs: Any
) -> Any:
    """Execute an async operation with exponential backoff retry logic."""
    last_error = None

    for attempt in range(max_attempts):
        try:
            return await operation(*args, **kwargs)
        except exception_types as error:
            last_error = error

            if attempt == max_attempts - 1:
                logger.error(f"Operation {operation.__name__} failed after {max_attempts} attempts")
                break

            wait_time = initial_delay * (2 ** attempt)
            logger.warning(f"Attempt {attempt + 1}/{max_attempts} failed: {str(error)}. Retrying in {wait_time:.1f}s")
            await asyncio.sleep(wait_time)

    if last_error:
        raise last_error
    raise MinecraftModManagerError("Retry operation failed without exception")


class ModManager:
    """Manages mod updates and version checking operations."""

    def __init__(self, configuration: Any, log_handler: logging.Logger) -> None:
        """Initialize the mod manager with configuration and logging."""
        self.configuration = configuration
        self.log_handler = log_handler
        self.http_session: Optional[aiohttp.ClientSession] = None
        self.mods_directory = Path(configuration.paths.mods)
        self.mods_directory.mkdir(parents=True, exist_ok=True)

    async def __aenter__(self) -> 'ModManager':
        """Initialize async context with HTTP session and connection pooling."""
        tcp_connector = aiohttp.TCPConnector(limit=10)
        request_timeout = aiohttp.ClientTimeout(total=DEFAULT_TIMEOUT)
        self.http_session = aiohttp.ClientSession(
            headers={"Accept": "application/json"},
            connector=tcp_connector,
            timeout=request_timeout
        )
        return self

    async def __aexit__(self, *exception_info: Any) -> None:
        """Clean up async context and close HTTP session."""
        if self.http_session:
            await self.http_session.close()
            self.http_session = None

    async def execute_api_request(self, api_endpoint: str, headers: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
        """Execute API request with comprehensive error handling."""
        if not self.http_session:
            raise ModUpdateError("HTTP session not initialized")

        async def _execute_single_request() -> Dict[str, Any]:
            if not self.http_session:
                raise ModUpdateError("HTTP session not initialized")

            await asyncio.sleep(1)  # Rate limiting delay
            request_headers = headers or {}
            
            async with self.http_session.get(api_endpoint, headers=request_headers) as api_response:
                if api_response.status == RATE_LIMIT_STATUS:
                    raise RateLimitError(f"API rate limit exceeded: {api_endpoint}")

                if api_response.status == 404:
                    raise ModUpdateError(f"API resource not found: {api_endpoint}")

                if api_response.status >= 500:
                    raise NetworkError(f"Server error {api_response.status}: {api_endpoint}")

                if api_response.status != 200:
                    raise ModUpdateError(f"API request failed with status {api_response.status}: {api_endpoint}")

                try:
                    response_data: Dict[str, Any] = await api_response.json()
                    return response_data
                except Exception as parse_error:
                    raise ModUpdateError(f"Failed to parse API response: {str(parse_error)}") from parse_error

        result = await retry_with_backoff(
            _execute_single_request,
            max_attempts=self.configuration.mods.max_retries,
            initial_delay=self.configuration.mods.base_delay,
            exception_types=(RateLimitError, NetworkError)
        )
        return cast(Dict[str, Any], result)

    def _collect_mod_sources(self) -> List[ModSource]:
        """Collect all configured mod sources from configuration."""
        mod_sources = []
        for modrinth_url in self.configuration.mods.sources.modrinth:
            mod_sources.append(ModSource(type="modrinth", url=modrinth_url))
        for curseforge_url in self.configuration.mods.sources.curseforge:
            mod_sources.append(ModSource(type="curseforge", url=curseforge_url))
        return mod_sources

    def _parse_project_id(self, mod_source: ModSource) -> str:
        """Parse project ID from mod source URL."""
        url_parts = urlparse(mod_source.url)

        if mod_source.type == "modrinth":
            id_match = re.search(r'/mod/([^/]+)', url_parts.path)
            if not id_match:
                raise ValueError(f"Invalid Modrinth URL format: {mod_source.url}")
            return id_match.group(1)
        elif mod_source.type == "curseforge":
            id_match = re.search(r'/mc-mods/([^/]+)', url_parts.path)
            if not id_match:
                raise ValueError(f"Invalid CurseForge URL format: {mod_source.url}")
            return id_match.group(1)
        else:
            raise ValueError(f"Unsupported mod source type: {mod_source.type}")

    async def retrieve_latest_versions(self) -> Dict[str, ModInfo]:
        """Retrieve latest compatible versions for all configured mods."""
        if not self.http_session:
            raise ModUpdateError("ModManager must be used as async context manager")

        version_info: Dict[str, ModInfo] = {}
        processing_failures: List[str] = []
        mod_sources = self._collect_mod_sources()

        if not mod_sources:
            self.log_handler.info("No mod sources configured")
            return version_info

        progress_tracker = tqdm(total=len(mod_sources), desc="Checking mod versions")

        try:
            # Process mods in batches to respect rate limits
            batch_size = self.configuration.mods.chunk_size
            for batch_start in range(0, len(mod_sources), batch_size):
                current_batch = mod_sources[batch_start:batch_start + batch_size]
                batch_tasks = [
                    self._retrieve_mod_info_safely(source, version_info, processing_failures)
                    for source in current_batch
                ]
                await asyncio.gather(*batch_tasks, return_exceptions=True)
                progress_tracker.update(len(current_batch))

                # Rate limiting delay between batches
                if batch_start + batch_size < len(mod_sources):
                    await asyncio.sleep(2)

            self.log_handler.info(f"Successfully processed {len(version_info)} mods")
            return version_info

        finally:
            progress_tracker.close()
            if processing_failures:
                self.log_handler.warning("Failed mods: " + ", ".join(processing_failures))

    async def _retrieve_mod_info_safely(
        self,
        mod_source: ModSource,
        version_info: Dict[str, ModInfo],
        processing_failures: List[str]
    ) -> None:
        """Safely retrieve mod information with comprehensive error handling."""
        try:
            project_identifier = self._parse_project_id(mod_source)

            if mod_source.type == "modrinth":
                await self._fetch_modrinth_mod_info(project_identifier, version_info, processing_failures)
            elif mod_source.type == "curseforge":
                await self._fetch_curseforge_mod_info(project_identifier, version_info, processing_failures)
            else:
                processing_failures.append(f"{project_identifier} (unsupported platform: {mod_source.type})")

        except Exception as processing_error:
            project_name = mod_source.url.split('/')[-1]
            processing_failures.append(f"{project_name} ({str(processing_error)})")

    async def _fetch_modrinth_mod_info(
        self,
        project_id: str,
        version_info: Dict[str, ModInfo],
        processing_failures: List[str]
    ) -> None:
        """Fetch mod information from Modrinth API."""
        # Get project metadata
        project_metadata = await self.execute_api_request(f"{MODRINTH_API_URL}/project/{project_id}")
        project_name = project_metadata.get('title', project_id)

        # Get version information
        version_params = f"?game_versions=[\"{self.configuration.minecraft.version}\"]&loaders=[\"{self.configuration.minecraft.modloader}\"]"
        version_data = await self.execute_api_request(f"{MODRINTH_API_URL}/project/{project_id}/version{version_params}")
        available_versions = cast(List[ModVersion], version_data)

        if available_versions:
            latest_version = available_versions[0]
            version_info[project_id] = {
                'version_id': latest_version['id'],
                'version_number': latest_version['version_number'],
                'download_url': latest_version['files'][0]['url'],
                'filename': latest_version['files'][0]['filename'],
                'project_name': project_name
            }
        else:
            processing_failures.append(f"{project_name} (no compatible version)")

    async def _fetch_curseforge_mod_info(
        self,
        project_slug: str,
        version_info: Dict[str, ModInfo],
        processing_failures: List[str]
    ) -> None:
        """Fetch mod information from CurseForge API."""
        headers = {"x-api-key": CURSEFORGE_API_KEY}
        
        # Search for project by slug
        search_url = f"{CURSEFORGE_API_URL}/mods/search?gameId={MINECRAFT_GAME_ID}&slug={project_slug}"
        search_response = await self.execute_api_request(search_url, headers)
        
        if not search_response.get('data'):
            processing_failures.append(f"{project_slug} (not found on CurseForge)")
            return
            
        project_data = search_response['data'][0]
        project_id = project_data['id']
        project_name = project_data['name']
        
        # Get mod files
        files_url = f"{CURSEFORGE_API_URL}/mods/{project_id}/files"
        files_response = await self.execute_api_request(files_url, headers)
        
        if not files_response.get('data'):
            processing_failures.append(f"{project_name} (no files available)")
            return
            
        # Filter for compatible versions
        compatible_files = []
        for file_data in files_response['data']:
            game_versions = [gv['versionString'] for gv in file_data.get('gameVersions', [])]
            if self.configuration.minecraft.version in game_versions:
                # Check mod loader compatibility
                mod_loaders = [ml['id'] for ml in file_data.get('sortableGameVersions', []) 
                              if ml['gameVersionTypeId'] == 68]  # Mod loader type ID
                if any(self.configuration.minecraft.modloader.lower() in str(ml).lower() for ml in mod_loaders):
                    compatible_files.append(file_data)
        
        if compatible_files:
            # Sort by file date (newest first)
            latest_file = sorted(compatible_files, key=lambda f: f['fileDate'], reverse=True)[0]
            version_info[project_slug] = {
                'version_id': str(latest_file['id']),
                'version_number': latest_file['displayName'],
                'download_url': latest_file['downloadUrl'],
                'filename': latest_file['fileName'],
                'project_name': project_name
            }
        else:
            processing_failures.append(f"{project_name} (no compatible version)")
            self.log_handler.error(f"Error processing mod {mod_source.url}: {str(processing_error)}")

    async def execute_mod_updates(self) -> None:
        """Execute updates for all mods to their latest compatible versions."""
        version_info = await self.retrieve_latest_versions()
        if not version_info:
            self.log_handler.info("All mods are up to date!")
            return

        successful_updates: List[str] = []
        update_failures: List[str] = []

        update_progress = tqdm(total=len(version_info), desc="Downloading mods")

        try:
            async with aiohttp.ClientSession() as download_session:
                update_tasks = [
                    self._execute_single_mod_update(download_session, mod_info, successful_updates, update_failures)
                    for mod_info in version_info.values()
                ]
                await asyncio.gather(*update_tasks, return_exceptions=True)
                update_progress.update(len(update_tasks))

        finally:
            update_progress.close()

        # Log update summary
        if successful_updates:
            self.log_handler.info(f"Updated mods: {', '.join(successful_updates)}")
        if update_failures:
            self.log_handler.warning(f"Failed mods: {', '.join(update_failures)}")

    async def _execute_single_mod_update(
        self,
        download_session: aiohttp.ClientSession,
        mod_info: ModInfo,
        successful_updates: List[str],
        update_failures: List[str]
    ) -> None:
        """Execute update for a single mod with error handling."""
        try:
            target_path = self.mods_directory / mod_info['filename']

            # Create backup of existing mod
            if target_path.exists():
                backup_timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
                backup_location = target_path.with_suffix(f".backup.{backup_timestamp}")
                shutil.copy2(target_path, backup_location)

            # Download updated version
            async with download_session.get(mod_info['download_url']) as download_response:
                if download_response.status != 200:
                    raise RuntimeError(f"Download failed with status {download_response.status}")

                mod_content = await download_response.read()
                target_path.write_bytes(mod_content)

            successful_updates.append(mod_info['project_name'])

        except Exception as update_error:
            self.log_handler.error(f"Error updating {mod_info['project_name']}: {str(update_error)}")
            update_failures.append(f"{mod_info['project_name']} ({str(update_error)})")


class BackupManager:
    """Manages server backup creation and maintenance operations."""

    def __init__(self, configuration: Any, log_handler: logging.Logger) -> None:
        """Initialize backup manager with configuration and logging."""
        self.configuration = configuration
        self.log_handler = log_handler
        self.backup_directory = Path(configuration.paths.backups)
        self.server_directory = Path(configuration.paths.server)
        self.backup_directory.mkdir(parents=True, exist_ok=True)

    async def generate_backup(self) -> bool:
        """Generate a new compressed backup of the server directory."""
        try:
            if not self.server_directory.exists():
                raise BackupError(f"Server directory does not exist: {self.server_directory}")

            backup_timestamp = datetime.now().strftime(self.configuration.backup.name_format)
            backup_file_path = self.backup_directory / f"{backup_timestamp}.tar.gz"

            with tarfile.open(backup_file_path, "w:gz") as backup_archive:
                backup_archive.add(self.server_directory, arcname=".")

            if not backup_file_path.exists() or backup_file_path.stat().st_size == 0:
                raise BackupError("Backup file was not created or is empty")

            backup_size_mb = backup_file_path.stat().st_size / (1024 * 1024)
            self.log_handler.info(f"Created backup: {backup_timestamp}.tar.gz ({backup_size_mb:.1f} MB)")
            return True

        except Exception as backup_error:
            self.log_handler.error(f"Failed to create backup: {str(backup_error)}")
            return False

    def maintain_backup_retention(self) -> None:
        """Remove old backups based on configured retention policy."""
        try:
            existing_backups = list(self.backup_directory.glob("*.tar.gz"))
            existing_backups.sort(key=lambda backup: backup.stat().st_mtime, reverse=True)

            retention_limit = self.configuration.backup.max_mod_backups
            for old_backup in existing_backups[retention_limit:]:
                old_backup.unlink()
                self.log_handler.info(f"Removed old backup: {old_backup.name}")

        except Exception as cleanup_error:
            self.log_handler.error(f"Failed to cleanup old backups: {str(cleanup_error)}")


class NotificationManager:
    """Manages Discord notifications and player communication."""

    def __init__(self, configuration: Any, log_handler: logging.Logger) -> None:
        """Initialize notification manager with configuration and logging."""
        self.configuration = configuration
        self.log_handler = log_handler
        self.discord_webhook = configuration.notifications.discord_webhook

    async def dispatch_discord_message(
        self, notification_title: str, notification_content: str, is_error_message: bool = False
    ) -> None:
        """Dispatch formatted notification message to Discord webhook."""
        if not self.discord_webhook:
            self.log_handler.debug("Discord notifications disabled - no webhook URL configured")
            return

        try:
            if len(notification_content) > DISCORD_MAX_LENGTH:
                notification_content = notification_content[:DISCORD_MAX_LENGTH - 3] + "..."

            discord_payload: Dict[str, Any] = {
                "embeds": [{
                    "title": notification_title,
                    "description": notification_content,
                    "color": DISCORD_ERROR_COLOR if is_error_message else DISCORD_SUCCESS_COLOR,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "footer": {"text": DISCORD_FOOTER_TEXT}
                }]
            }

            async with aiohttp.ClientSession() as notification_session:
                async with notification_session.post(self.discord_webhook, json=discord_payload) as webhook_response:
                    if webhook_response.status not in (200, 204):
                        raise RuntimeError(f"Discord API returned status {webhook_response.status}")

        except Exception as notification_error:
            self.log_handler.error(f"Failed to send Discord notification: {str(notification_error)}")

    async def broadcast_player_warning(self) -> None:
        """Broadcast warning messages to players about upcoming server maintenance."""
        try:
            warning_message = self.configuration.notifications.warning_template.format(
                minutes=self.configuration.notifications.warning_intervals[0]
            )
            self.log_handler.info(f"Server restart warning: {warning_message}")
            await self.dispatch_discord_message("Server Warning", warning_message)
        except Exception as warning_error:
            self.log_handler.error(f"Failed to send player warning: {str(warning_error)}")


class ServerController:
    """Controls Minecraft server process lifecycle and status monitoring."""

    def __init__(self, configuration: Any, log_handler: logging.Logger) -> None:
        """Initialize server controller with configuration and logging."""
        self.configuration = configuration
        self.log_handler = log_handler
        self.server_process: Optional[asyncio.subprocess.Process] = None

    async def initialize_server(self) -> bool:
        """Initialize and start the Minecraft server process."""
        try:
            if await self.check_server_status():
                self.log_handler.warning("Server is already running")
                return True

            server_directory = Path(self.configuration.paths.server)
            server_jar = server_directory / self.configuration.server.jar

            if not server_directory.exists():
                raise ServerError(f"Server directory does not exist: {server_directory}")
            if not server_jar.exists():
                raise ServerError(f"Server jar file does not exist: {server_jar}")

            self.log_handler.info("Starting server...")
            self.server_process = await asyncio.create_subprocess_shell(
                self.configuration.server.start_command,
                cwd=str(server_directory),
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # Monitor server startup with timeout
            startup_timeout = 60
            elapsed_startup_time = 0

            while elapsed_startup_time < startup_timeout:
                await asyncio.sleep(2)
                elapsed_startup_time += 2

                if await self.check_server_status():
                    self.log_handler.info(f"Server started successfully after {elapsed_startup_time}s")
                    return True

                if self.server_process and self.server_process.returncode is not None:
                    raise ServerError("Server process terminated unexpectedly")

            raise ServerError(f"Server failed to start within {startup_timeout}s")

        except Exception as startup_error:
            self.log_handler.error(f"Error starting server: {str(startup_error)}")
            return False

    async def terminate_server(self) -> bool:
        """Terminate the Minecraft server process gracefully."""
        try:
            if not await self.check_server_status():
                self.log_handler.warning("Server is not running")
                return True

            self.log_handler.info("Stopping server...")

            # Send graceful shutdown command via screen session
            await asyncio.create_subprocess_shell(
                f'screen -S minecraft -X stuff "{self.configuration.server.stop_command}^M"'
            )

            # Monitor server shutdown
            shutdown_start = asyncio.get_event_loop().time()
            while await self.check_server_status():
                if asyncio.get_event_loop().time() - shutdown_start > self.configuration.server.max_stop_wait:
                    self.log_handler.error("Server failed to stop within timeout")
                    return False
                await asyncio.sleep(1)

            self.log_handler.info("Server stopped successfully")
            return True

        except Exception as shutdown_error:
            self.log_handler.error(f"Error stopping server: {str(shutdown_error)}")
            return False

    async def check_server_status(self) -> bool:
        """Check current status of the Minecraft server process."""
        try:
            if self.server_process and self.server_process.returncode is None:
                return True

            # Verify screen session existence
            status_check = await asyncio.create_subprocess_shell(
                "screen -ls | grep minecraft",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            await status_check.communicate()
            return status_check.returncode == 0

        except Exception as status_error:
            self.log_handler.error(f"Error checking server status: {str(status_error)}")
            return False
