# MinecraftModManager.py
# A comprehensive Minecraft server management tool that handles mod updates, server maintenance,
# backups, and notifications. Supports automated updates and server control via command line.

# Standard library imports
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Optional
import argparse
import asyncio
import json
import logging
import os
import sys
import tarfile
import time
import subprocess
import re

# Third party imports
import aiohttp
import requests
from tqdm import tqdm

# Support classes for configuration and logging management
class ConfigManager:
    """Handles configuration loading and validation."""
    # Configuration management code
    
class LoggingManager:
    """Handles logging configuration and setup."""
    # Logging setup code

def parse_jsonc(jsonc_str: str) -> dict:
    """
    Parse JSONC (JSON with Comments) into a Python dictionary.
    Handles both single-line (//) and multi-line (/* */) comments.
    """
    # Remove multi-line comments
    jsonc_str = re.sub(r'/\*.*?\*/', '', jsonc_str, flags=re.DOTALL)
    
    # Remove single-line comments
    jsonc_str = re.sub(r'//.*$', '', jsonc_str, flags=re.MULTILINE)
    
    # Remove empty lines and leading/trailing whitespace
    jsonc_str = '\n'.join(line.strip() for line in jsonc_str.splitlines() if line.strip())
    
    try:
        return json.loads(jsonc_str)
    except json.JSONDecodeError as e:
        raise ValueError(f"Invalid JSONC format: {str(e)}")

class MinecraftModManager:
    """
    Main class that manages all aspects of a Minecraft server including:
    - Mod updates and version checking via Modrinth API
    - Server start/stop/restart
    - Automated backups and cleanup
    - Discord notifications
    - Player warnings for maintenance
    """
    
    def __init__(self, config_path: str = "config.jsonc"):
        """
        Initialize the mod manager by:
        1. Loading and validating config
        2. Setting up logging
        3. Creating required directories
        4. Validating server jar exists
        """
        # Resolve config path relative to script location
        config_path = Path(config_path)
        if not config_path.is_absolute():
            config_path = Path(__file__).parent / config_path
            
        try:
            # Load and parse JSONC config
            config_text = config_path.read_text(encoding='utf-8')
            self.config = parse_jsonc(config_text)
            
            # Initialize logging
            self.logger: Optional[logging.Logger] = None
            self.setup_logging()
            
            self._validate_config()
            
            # Create all required directories in parallel
            required_dirs = [
                Path(self.config['paths']['local_mods']),
                Path(self.config['paths']['backups']), 
                Path(self.config['paths']['minecraft'])
            ]
            
            for directory in required_dirs:
                self._ensure_directory(directory)
                
            self._validate_server_jar()
            
        except Exception as e:
            # Log initialization errors if logger is ready, otherwise print
            if hasattr(self, 'logger') and self.logger:
                self._handle_operation_error("Initialization", e)
            else:
                print(f"Initialization error: {str(e)}")
            raise

    def verify_server_status(self) -> bool:
        """Check if server process is currently running by looking for java process"""
        return bool(self._get_server_pid())

    def get_player_count(self) -> int:
        """Parse server log to get current online player count"""
        if not self.verify_server_status():
            return 0
        try:
            # Check last 50 lines for player count message
            lines = self._read_server_log(50)
            for line in reversed(lines):
                if "There are" in line and "players online" in line:
                    try:
                        return int(line.split()[2])
                    except (IndexError, ValueError):
                        continue
            return 0
            
        except Exception as e:
            self._handle_operation_error("Player Count Check", e, notify=False)
            return 0

    async def fetch_latest_mod_versions(self) -> Dict[str, dict]:
        """
        Query Modrinth API to get latest versions of all configured mods.
        Uses connection pooling and rate limiting to avoid API issues.
        Returns dict mapping mod URLs to version info.
        """
        headers = {
            "Accept": "application/json"
        }
        mod_info = {}
        failed_mods = []

        # Setup connection pool with limits
        connector = aiohttp.TCPConnector(limit=10)
        timeout = aiohttp.ClientTimeout(total=300)
        
        progress_bar = tqdm(total=len(self.config['modrinth_urls']), desc="Checking mod versions")
        
        async with aiohttp.ClientSession(headers=headers, 
                                       connector=connector, 
                                       timeout=timeout) as session:
            # Process mods in chunks to avoid rate limits
            chunk_size = self.config['api']['chunk_size']
            for i in range(0, len(self.config['modrinth_urls']), chunk_size):
                chunk = self.config['modrinth_urls'][i:i + chunk_size]
                tasks = []
                
                for url in chunk:
                    tasks.append(self._fetch_mod_info(session, url, mod_info, failed_mods, progress_bar))
                
                await asyncio.gather(*tasks)
                
                # Rate limiting delay between chunks
                if i + chunk_size < len(self.config['modrinth_urls']):
                    await asyncio.sleep(2)
        
        progress_bar.close()

        # Notify about any failed mods
        if failed_mods:
            self.send_discord_notification(
                "Mod Update Issues",
                "Failed to process:\n" + "\n".join(failed_mods),
                True
            )

        return mod_info

    async def _fetch_mod_info(self, session: aiohttp.ClientSession, url: str, 
                            mod_info: Dict, failed_mods: List, progress_bar: tqdm) -> None:
        """
        Fetch version info for a single mod from Modrinth API.
        Handles retries and rate limiting.
        Updates mod_info dict with latest compatible version.
        """
        max_retries = self.config['api']['max_retries']
        base_delay = self.config['api']['base_delay']
        
        async def make_request(endpoint: str, retry_count: int = 0):
            """Helper to make API request with retry logic"""
            try:
                await asyncio.sleep(2)  # Rate limiting delay
                async with session.get(endpoint) as response:
                    if response.status == 429:  # Rate limited
                        if retry_count < max_retries:
                            retry_delay = base_delay * (2 ** retry_count)
                            self.logger.warning(f"Rate limited, waiting {retry_delay} seconds...")
                            await asyncio.sleep(retry_delay)
                            return await make_request(endpoint, retry_count + 1)
                        else:
                            raise Exception(f"Rate limit exceeded after {max_retries} retries")
                    
                    if response.status != 200:
                        raise Exception(f"API returned status {response.status}")
                    
                    return await response.json()
                    
            except Exception as e:
                raise Exception(f"Request failed: {str(e)}")

        try:
            # Extract project ID from URL
            project_id = url.split('/')[-1]
            
            # Get project details
            project_data = await make_request(f"https://api.modrinth.com/v2/project/{project_id}")
            if isinstance(project_data, str):
                raise Exception(f"Invalid project data: {project_data}")
            mod_name = project_data.get('title', project_id)
            
            await asyncio.sleep(2)  # Rate limiting delay
            
            # Get version list
            versions = await make_request(f"https://api.modrinth.com/v2/project/{project_id}/version")
            if isinstance(versions, str):
                raise Exception(f"Invalid version data: {versions}")
            
            # Filter for compatible versions
            compatible_versions = [
                v for v in versions
                if self.config['minecraft']['version'] in v.get('game_versions', []) and
                self.config['minecraft']['modloader'].lower() in [loader.lower() for loader in v.get('loaders', [])]
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
                # Log warning if no compatible version found
                self.logger.warning(
                    f"No compatible version for {mod_name}. "
                    f"Required MC version: {self.config['minecraft']['version']}, "
                    f"Required modloader: {self.config['minecraft']['modloader']}, "
                    f"Available versions: {[v.get('game_versions', []) for v in versions[:3]]}"
                )
                failed_mods.append(
                    f"{mod_name} (no compatible version for MC {self.config['minecraft']['version']} with {self.config['minecraft']['modloader']})"
                )

        except Exception as e:
            mod_name = project_id if 'mod_name' not in locals() else mod_name
            failed_mods.append(f"{mod_name} ({str(e)})")
            self.logger.error(f"Error processing mod {url}: {str(e)}")
        finally:
            progress_bar.update(1)

    async def update_mods(self) -> None:
        """
        Update all mods to their latest compatible versions.
        Downloads new versions and tracks update status.
        Sends Discord notification with results.
        """
        try:
            # Get latest version info for all mods
            mod_info = await self.fetch_latest_mod_versions()
            if not mod_info:
                self.send_discord_notification("Mod Updates", "✅ All mods are up to date!")
                return
            
            # Track status of each mod
            updated_mods = []
            skipped_mods = []
            failed_mods = []
            newly_added_mods = []
            
            progress_bar = tqdm(total=len(mod_info), desc="Downloading mods")
            
            # Download mods in parallel
            async with aiohttp.ClientSession() as session:
                tasks = []
                for url, info in mod_info.items():
                    tasks.append(self._update_single_mod(
                        session, info,
                        updated_mods, skipped_mods, failed_mods, 
                        newly_added_mods, 
                        progress_bar
                    ))
                await asyncio.gather(*tasks)
            
            progress_bar.close()
            
            # Send summary notification
            self._send_update_summary(updated_mods, skipped_mods, failed_mods, len(mod_info))
            
        except Exception as e:
            self.logger.error(f"Error updating mods: {str(e)}")
            self.send_discord_notification("Mod Update Error", str(e), True)
            raise

    async def _update_single_mod(self, session: aiohttp.ClientSession, info: Dict,
                               updated_mods: List, skipped_mods: List, 
                               failed_mods: List, newly_added_mods: List, 
                               progress_bar: tqdm) -> None:
        """
        Update a single mod file:
        1. Check if update needed by comparing file sizes
        2. Remove old version if exists
        3. Download and save new version
        4. Track status in appropriate list
        """
        try:
            mod_name = info['project_name']
            current_mod_path = Path(self.config['paths']['local_mods']) / info['filename']
            
            # Check if update needed
            if current_mod_path.exists():
                current_size = current_mod_path.stat().st_size
                async with session.head(info['download_url']) as response:
                    new_size = int(response.headers.get('content-length', 0))
                
                if current_size == new_size:
                    skipped_mods.append(f"• {mod_name} ({info['version_number']})")
                    progress_bar.update(1)
                    return
                
                # Remove old version
                self._run_command(f"sudo rm {current_mod_path}")
            else:
                newly_added_mods.append(f"• {mod_name} ({info['version_number']})")
            
            # Download and save new version
            async with session.get(info['download_url']) as response:
                content = await response.read()
                current_mod_path.write_bytes(content)
            
            updated_mods.append(f"• {mod_name} → {info['version_number']}")
            
        except Exception as e:
            failed_mods.append(f"• {mod_name}: {str(e)}")
        finally:
            progress_bar.update(1)

    def run_maintenance(self) -> None:
        """
        Run complete server maintenance cycle:
        1. Check for and warn players
        2. Stop server
        3. Create backup
        4. Update mods
        5. Clean old backups
        6. Restart server
        """
        try:
            # Check for players and warn if needed
            if self.get_player_count() > 0:
                self._warn_players()
            
            # Stop server if running
            if self.verify_server_status():
                self.send_discord_notification("Maintenance", "🔄 Stopping server...")
                if not self.stop_server():
                    raise Exception("Failed to stop server")
            
            # Create backup
            if not self.create_backup():
                raise Exception("Backup failed")

            # Update mods
            asyncio.run(self.update_mods())
            
            # Clean old backups
            self.cleanup_old_backups()

            # Restart server
            self.send_discord_notification("Maintenance", "🔄 Starting server...")
            if not self.restart_server():
                raise Exception("Server restart failed")

            self.send_discord_notification(
                "Maintenance Complete",
                "✅ Server maintenance completed successfully"
            )
            
        except Exception as e:
            self.logger.error(f"Maintenance failed: {str(e)}")
            self.send_discord_notification("Maintenance Failed", f"❌ {str(e)}", True)

    def run_automated_update(self) -> None:
        """
        Run automated update process with safety checks:
        1. Start server if not running
        2. Check for players and warn if needed
        3. Run maintenance cycle
        """
        try:
            # Start server if needed
            if not self.verify_server_status():
                self.logger.info("Server is not running, starting it...")
                if not self.restart_server():
                    raise Exception("Failed to start server")
            
            # Check for players
            if self.get_player_count() > 0:
                self.send_discord_notification(
                    "Server Update", 
                    "🔄 Players online - initiating shutdown countdown"
                )
                self._warn_players()
            
            # Run maintenance
            self.send_discord_notification("Update Started", "🔄 Beginning update process...")
            self.run_maintenance()
            
        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            self.send_discord_notification("Update Failed", f"❌ {str(e)}", True)

    def _get_server_pid(self) -> Optional[str]:
        """Get PID of running Minecraft server process"""
        try:
            result = os.popen("ps aux | grep '[j]ava.*server.jar'").read().strip()
            if not result:
                return None
            return result.split()[1]
        except (IndexError, Exception) as e:
            self.logger.debug(f"Could not get server PID: {str(e)}")
            return None

    def _read_server_log(self, max_lines: int = 50) -> List[str]:
        """Read last N lines from server log file"""
        try:
            log_file = Path(self.config['paths']['minecraft']) / "logs/latest.log"
            if not log_file.exists():
                return []
            with open(log_file, 'r', encoding='utf-8') as f:
                return f.readlines()[-max_lines:]
        except Exception as e:
            self.logger.error(f"Error reading log file: {str(e)}")
            return []

    def _ensure_directory(self, path: Path) -> None:
        """Create directory if it doesn't exist"""
        if not path.exists():
            os.makedirs(path, exist_ok=True)

    def _handle_operation_error(self, operation: str, error: Exception) -> None:
        """Centralized error handling with logging and notifications"""
        error_msg = f"{operation} failed: {str(error)}"
        self.logger.error(error_msg)
        self.send_discord_notification(
            f"{operation} Error",
            f"❌ {error_msg}",
            is_error=True
        )

    def _run_command(self, command: str) -> str:
        """Execute system command and return output"""
        try:
            result = subprocess.run(
                command,
                shell=True,
                check=True,
                capture_output=True,
                text=True
            )
            return result.stdout
        except subprocess.CalledProcessError as e:
            raise RuntimeError(f"Command failed: {e.stderr}")

    def create_backup(self) -> bool:
        """
        Create server backup:
        1. Create temp directory
        2. Copy mods, config, and world
        3. Create tar archive
        4. Clean up temp files
        """
        try:
            # Generate backup name with timestamp
            timestamp = datetime.now().strftime(self.config['maintenance']['backup_name_format'])
            backup_path = Path(self.config['paths']['backups']) / f"{timestamp}.tar.gz"
            
            # Setup backup directory
            self._run_command(f"sudo mkdir -p {self.config['paths']['backups']}")
            self._run_command(f"sudo chown -R $USER:$USER {self.config['paths']['backups']}")
            self._run_command(f"sudo chmod -R a+r {self.config['paths']['minecraft']}")
            
            # Create temp directory
            temp_backup_dir = Path(self.config['paths']['backups']) / "temp_backup"
            self._run_command(f"sudo rm -rf {temp_backup_dir}")
            self._run_command(f"sudo mkdir -p {temp_backup_dir}")
            self._run_command(f"sudo chown -R $USER:$USER {temp_backup_dir}")
            
            # Copy files to backup
            self._run_command(f"sudo cp -r {self.config['paths']['local_mods']} {temp_backup_dir}/mods")
            self._run_command(f"sudo cp -r {self.config['paths']['minecraft']}/config {temp_backup_dir}/config")
            self._run_command(f"sudo cp -r {self.config['paths']['minecraft']}/world {temp_backup_dir}/world")
            self._run_command(f"sudo chown -R $USER:$USER {temp_backup_dir}")
            
            # Create archive
            with tarfile.open(backup_path, 'w:gz') as tar:
                tar.add(temp_backup_dir / 'mods', arcname='mods')
                tar.add(temp_backup_dir / 'config', arcname='config')
                tar.add(temp_backup_dir / 'world', arcname='world')
            
            # Cleanup temp files
            self._run_command(f"sudo rm -rf {temp_backup_dir}")
            
            self.logger.info(f"Created backup at {backup_path}")
            self.send_discord_notification(
                "Server Backup",
                f"✅ Created backup: {timestamp}.tar.gz"
            )
            return True
            
        except Exception as e:
            self.logger.error(f"Backup error: {str(e)}")
            self.send_discord_notification(
                "Backup Failed",
                f"❌ Backup failed: {str(e)}",
                True
            )
            return False

    def cleanup_old_backups(self) -> None:
        """
        Remove old backups based on:
        - Age (older than retention days)
        - Count (keep only max_backups)
        """
        try:
            # Get config values
            keep_days = self.config['maintenance']['backup_retention_days']
            max_backups = self.config['maintenance']['max_backups']
            now = time.time()
            
            # Match backup filename pattern
            backup_pattern = f"*{self.config['maintenance']['backup_name_format'].replace('%Y','*').replace('%m','*').replace('%d','*').replace('%H','*').replace('%M','*')}.tar.gz"
            
            # Get all backups with modification times
            backups = []
            for backup in Path(self.config['paths']['backups']).glob(backup_pattern):
                mtime = backup.stat().st_mtime
                backups.append((backup, mtime))
            
            # Sort by newest first
            backups.sort(key=lambda x: x[1], reverse=True)
            
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
                    self._run_command(f"sudo rm {backup}")
                    self.logger.info(f"Removed old backup: {backup.name} (reason: {delete_reason})")
                        
        except Exception as e:
            self.logger.error(f"Failed to cleanup old backups: {str(e)}")

    def send_discord_notification(self, title: str, message: str, is_error: bool = False) -> None:
        """Send formatted notification to Discord webhook"""
        if not self.config['notifications']['discord_webhook']:
            return
            
        try:
            # Format Discord embed
            payload = {
                "embeds": [{
                    "title": title,
                    "description": message[:1997] + "..." if len(message) > 2000 else message,
                    "color": 0xFF0000 if is_error else 0x00FF00,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "footer": {"text": "Minecraft Mod Manager"}
                }]
            }
            
            # Send webhook request
            response = requests.post(
                self.config['notifications']['discord_webhook'],
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            if response.status_code not in (200, 204):
                self.logger.error(f"Discord API error: {response.status_code}")
                
        except Exception as e:
            self.logger.error(f"Failed to send Discord notification: {str(e)}")

    def _warn_players(self) -> None:
        """Send countdown warnings to players before maintenance"""
        warnings = self.config['maintenance']['warning_intervals']
        
        try:
            for warning in warnings:
                time_val = warning['time']
                unit = warning['unit']
                
                self.send_server_message(
                    f"§c[WARNING] Server maintenance in {time_val} {unit}!"
                )
                sleep_time = 60 if unit == "minutes" else 5
                time.sleep(sleep_time)
                
            self.send_server_message("§c[WARNING] Starting maintenance now!")
            
        except Exception as e:
            self.logger.error(f"Failed to send warnings: {str(e)}")

    def setup_logging(self) -> None:
        """
        Configure logging with:
        - File output
        - Console output
        - Formatting
        """
        try:
            # Create logs directory
            log_dir = Path(self.config['paths']['logs']).parent
            if not log_dir.exists():
                os.makedirs(log_dir, exist_ok=True)
            
            # Setup logger
            self.logger = logging.getLogger('MinecraftModManager')
            self.logger.setLevel(logging.INFO)
            
            # File handler
            file_handler = logging.FileHandler(
                self.config['paths']['logs'],
                mode='a'
            )
            
            # Console handler
            console_handler = logging.StreamHandler()
            
            # Formatter
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )
            
            # Apply formatter
            file_handler.setFormatter(formatter)
            console_handler.setFormatter(formatter)
            
            # Add handlers
            self.logger.addHandler(file_handler)
            self.logger.addHandler(console_handler)
            
            self.logger.info("Logging initialized")
            
        except Exception as e:
            print(f"Failed to setup logging: {str(e)}")
            raise

    def _validate_config(self) -> None:
        """Validate all required config settings are present"""
        required_paths = ['local_mods', 'backups', 'minecraft', 'server_jar', 'logs']
        for path in required_paths:
            if path not in self.config['paths']:
                raise ValueError(f"Missing required path: {path}")
            
        if 'minecraft' not in self.config:
            raise ValueError("Missing minecraft configuration section")
        if 'version' not in self.config['minecraft']:
            raise ValueError("Missing minecraft version")
        if 'modloader' not in self.config['minecraft']:
            raise ValueError("Missing minecraft modloader")

    def _validate_server_jar(self) -> None:
        """Verify server jar file exists"""
        server_jar = Path(self.config['paths']['server_jar'])
        if not server_jar.exists():
            raise FileNotFoundError(f"Server jar not found at {server_jar}")

    def control_server(self, action: str) -> bool:
        """
        Unified server control method for start/stop/restart operations.
        Args:
            action: One of 'start', 'stop', or 'restart'
        """
        try:
            server_running = self.verify_server_status()
            
            if action == 'stop':
                if not server_running:
                    return True
                return self._stop_server_process()
            
            elif action == 'start':
                if server_running:
                    return True
                return self._start_server_process()
            
            elif action == 'restart':
                if server_running and not self._stop_server_process():
                    return False
                return self._start_server_process()
            
            else:
                raise ValueError(f"Invalid server action: {action}")
            
        except Exception as e:
            self._handle_operation_error(f"Server {action}", e)
            return False

    def _stop_server_process(self) -> bool:
        """Internal method to stop the server process"""
        try:
            pid = self._get_server_pid()
            if not pid:
                return True
            
            # Try graceful shutdown
            self._run_command(f"sudo kill -TERM {pid}")
            
            # Wait for shutdown
            for _ in range(20):
                if not self._get_server_pid():
                    return True
                time.sleep(1)
            
            # Force kill if needed
            if self._get_server_pid():
                self._run_command(f"sudo kill -9 {pid}")
                time.sleep(2)
            
            return not bool(self._get_server_pid())
            
        except Exception as e:
            raise RuntimeError(f"Failed to stop server: {str(e)}")

    def _start_server_process(self) -> bool:
        """Internal method to start the server process"""
        try:
            os.chdir(self.config['paths']['minecraft'])
            
            command = (self.config['server'].get('custom_flags') 
                      if self.config['server'].get('flags_source') == 'custom'
                      else self._build_default_command())
            command = f"{command} -jar {self.config['paths']['server_jar']} nogui"
            
            self.logger.info(f"Starting server with command: {command}")
            self._run_command(f"nohup {command} > /dev/null 2>&1 &")
            
            return self._wait_for_server_ready()
            
        except Exception as e:
            raise RuntimeError(f"Failed to start server: {str(e)}")

    def _wait_for_server_ready(self, max_wait: int = 120) -> bool:
        """Wait for server to be fully initialized"""
        start_time = time.time()
        
        while time.time() - start_time < max_wait:
            if self.verify_server_status():
                log_lines = self._read_server_log(100)
                for line in reversed(log_lines):
                    if "Done" in line and "For help, type" in line:
                        self.logger.info("Server is now online and ready!")
                        self.send_discord_notification(
                            "Server Status",
                            "✅ Server is now online and ready!"
                        )
                        return True
            time.sleep(2)
        
        self.logger.error("Server did not start within the expected time")
        return False

    def send_server_message(self, message: str) -> None:
        """Send in-game message (requires RCON support)"""
        self.logger.warning("Server messages not implemented - no RCON support")

    def _send_update_summary(self, updated_mods: List[str], skipped_mods: List[str], 
                            failed_mods: List[str], total_mods: int) -> None:
        """Send formatted update summary to Discord"""
        summary = []
        
        if updated_mods:
            # Show updates and failures
            summary.append("📦 **Updated Mods:**\n" + "\n".join(updated_mods))
            if failed_mods:
                summary.append("❌ **Failed Mods:**\n" + "\n".join(failed_mods))
        else:
            # Show up-to-date count
            summary.append(f"✅ All mods are up to date! ({len(skipped_mods)}/{total_mods})")
        
        # Add statistics
        stats = (
            f"📊 **Statistics:**\n"
            f"Total Mods: {total_mods}\n"
            f"Updated: {len(updated_mods)}\n"
            f"Up to date: {len(skipped_mods)}\n"
            f"Failed: {len(failed_mods)}"
        )
        
        summary.append(stats)
        
        self.send_discord_notification(
            "Mod Update Summary",
            "\n\n".join(summary)
        )

def main():
    """
    Command-line interface entry point.
    Supports:
    - Auto-update
    - Server start/stop/restart
    - Status check
    - Manual maintenance
    """
    parser = argparse.ArgumentParser(description='Minecraft Mod Manager')
    
    # Create a mutually exclusive group for server control commands
    server_group = parser.add_mutually_exclusive_group()
    server_group.add_argument('--start', action='store_true',
                            help='Start the Minecraft server')
    server_group.add_argument('--stop', action='store_true',
                            help='Stop the Minecraft server')
    server_group.add_argument('--restart', action='store_true',
                            help='Restart the Minecraft server')
    server_group.add_argument('--status', action='store_true',
                            help='Check server status')
    
    parser.add_argument('--auto-update', action='store_true',
                       help='Run automated update process')
    
    args = parser.parse_args()

    try:
        manager = MinecraftModManager()
        
        # Handle server control commands
        if args.status:
            status = "running" if manager.verify_server_status() else "stopped"
            players = manager.get_player_count() if status == "running" else 0
            print(f"Server is {status} with {players} players online")
            sys.exit(0)
            
        for action in ['start', 'stop', 'restart']:
            if getattr(args, action):
                print(f"{action.capitalize()}ing server...")
                if manager.control_server(action):
                    print(f"Server {action}ed successfully")
                    sys.exit(0)
                else:
                    print(f"Failed to {action} server")
                    sys.exit(1)
                
        if args.auto_update:
            manager.run_automated_update()
            
        else:
            print("Warning: This will update all mods and restart the server.")
            print("Players will be warned with a countdown if any are online.")
            input("Press Enter to continue or Ctrl+C to cancel...")
            manager.run_maintenance()
            
    except KeyboardInterrupt:
        print("\nOperation cancelled by user")
        sys.exit(0)
    except Exception as e:
        print(f"Error: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main()