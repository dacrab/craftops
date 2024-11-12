#region Imports
from datetime import datetime, timezone
from logging.handlers import RotatingFileHandler
from pathlib import Path
from typing import Dict, List, Optional
import argparse
import asyncio
import aiohttp
import json
import logging
import os
import requests
import sys
import tarfile
import time
from tqdm import tqdm
#endregion

class MinecraftModManager:
    """Manages Minecraft mods, server maintenance, and backups."""

    def __init__(self, config_path: str = "config.json"):
        """Initialize the mod manager with config."""
        # Initialize logger first
        self.logger: Optional[logging.Logger] = None
        self.setup_logging()
        
        # Load and validate config
        config_path = Path(config_path)
        if not config_path.is_absolute():
            config_path = Path(__file__).parent / config_path
            
        try:
            # Load config file
            self.config = json.loads(config_path.read_text())
            self._validate_config()
            
            # Create required directories in parallel
            required_dirs = [
                Path(self.config['paths']['local_mods']),
                Path(self.config['paths']['backups']),
                Path(self.config['paths']['minecraft'])
            ]
            
            # Create directories concurrently
            for directory in required_dirs:
                if directory and not directory.exists():
                    self._ensure_directory(directory)
                
            self._validate_server_jar()
            
        except Exception as e:
            self._handle_operation_error("Initialization", e)
            raise

    #region Configuration
    def _validate_config(self) -> None:
        """Validate required configuration fields."""
        # Define required config structure
        required_config = {
            'paths': {
                'local_mods': 'Local mods directory',
                'minecraft': 'Minecraft directory', 
                'server_jar': 'Server JAR file path',
                'backups': 'Backup directory'
            },
            'minecraft': {
                'modloader': 'Modloader type',
                'version': 'Minecraft version'
            },
            'server': {
                'memory': 'Server memory settings',
                'java_flags': 'Java flags configuration'
            },
            'maintenance': {
                'warning_intervals': 'Warning intervals',
                'backup_name_format': 'Backup name format',
                'backup_retention_days': 'Backup retention days'
            },
            'api': {
                'startup_timeout': 'Server startup timeout',
                'user_agent': 'API user agent'
            }
        }
        
        # Track missing fields
        missing = []
        
        # Validate all required fields
        for section, fields in required_config.items():
            if section not in self.config:
                missing.extend(fields.values())
            else:
                missing.extend(desc for key, desc in fields.items() 
                             if key not in self.config[section])
        
        if missing:
            raise ValueError(f"Missing required config: {', '.join(missing)}")
            
        # Validate modloader value
        valid_modloaders = ['fabric', 'forge', 'quilt']
        if self.config['minecraft']['modloader'].lower() not in valid_modloaders:
            raise ValueError(f"Invalid modloader: {self.config['minecraft']['modloader']}. Must be one of: {', '.join(valid_modloaders)}")

    def _validate_server_jar(self) -> None:
        """Validate server.jar exists in correct location."""
        if not Path(self.config['paths']['server_jar']).exists():
            raise FileNotFoundError(f"server.jar not found at {self.config['paths']['server_jar']}")

    def setup_logging(self) -> None:
        """Configure logging with file and console handlers."""
        self.logger = logging.getLogger('ModManager')
        self.logger.setLevel(logging.DEBUG)
        self.logger.handlers = []
        
        # Use configured log path from config
        log_path = Path(self.config['paths']['logs']) if hasattr(self, 'config') else Path("/home/Minecraft/logs/mod_manager.log")
        os.makedirs(log_path.parent, exist_ok=True)
        
        # File handler
        file_handler = RotatingFileHandler(
            log_path, 
            maxBytes=5*1024*1024,
            backupCount=5
        )
        file_handler.setFormatter(
            logging.Formatter(
                '%(asctime)s [%(levelname)s] '
                '%(filename)s:%(lineno)d - %(funcName)s: %(message)s'
            )
        )
        file_handler.setLevel(logging.DEBUG)
        
        # Console handler
        console_handler = logging.StreamHandler()
        console_handler.setFormatter(
            logging.Formatter('%(levelname)s: %(message)s')
        )
        console_handler.setLevel(logging.INFO)
        
        self.logger.addHandler(file_handler)
        self.logger.addHandler(console_handler)
        
        # Log startup info
        self.logger.info("="*50)
        self.logger.info(f"Mod Manager Started at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        self.logger.info(f"Python Version: {sys.version}")
        self.logger.info(f"Running as user: {os.getenv('USER')}")
        self.logger.debug(f"Working Directory: {os.getcwd()}")
        self.logger.debug(f"Script Location: {Path(__file__).absolute()}")
        self.logger.info("="*50)
    #endregion

    #region Server Management
    def get_player_count(self) -> int:
        """Get current number of online players."""
        try:
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

    def verify_server_status(self) -> bool:
        """Check if server is running and responding."""
        try:
            # Check process
            pid = self._get_server_pid()
            if not pid:
                self.logger.debug("No Java process found for Minecraft server")
                return False

            # Check server log
            lines = self._read_server_log(self.config['logging']['max_lines']['server_check'])
            if not lines:
                return False
                
            # Check startup messages
            startup_indicators = [
                "Done (",  # Server done loading
                "For help, type",  # Command help message
                "Starting minecraft server"  # Initial startup
            ]
            
            return any(msg in ''.join(lines) for msg in startup_indicators)
                
        except Exception as e:
            self._handle_operation_error("Server Status Check", e, notify=False)
            return False

    def stop_server(self) -> bool:
        """Stop the Minecraft server gracefully."""
        try:
            pid = self._get_server_pid()
            if not pid:
                self.logger.info("Server is already stopped")
                return True
            
            self.logger.info(f"Found server process with PID: {pid}")
            self.logger.info("Initiating server shutdown...")
            
            # Try graceful shutdown
            os.system(f"sudo kill {pid}")
            
            # Wait for shutdown
            for i in range(30):  # 30 second timeout
                if i % 5 == 0:  # Log every 5 seconds
                    self.logger.info(f"Waiting for server to stop... ({i}s)")
                if not self.verify_server_status():
                    self.logger.info("Server stopped gracefully")
                    return True
                time.sleep(1)
            
            # Force kill if needed
            self.logger.warning("Server did not stop gracefully, forcing shutdown...")
            os.system(f"sudo kill -9 {pid}")
            
            # Final check
            time.sleep(2)
            if self.verify_server_status():
                raise Exception("Failed to stop server even after force kill")
            
            self.logger.info("Server stopped successfully")
            return True
            
        except Exception as e:
            self._handle_operation_error("Server Stop", e)
            return False

    def restart_server(self) -> bool:
        """Restart the Minecraft server."""
        try:
            if not Path(self.config['paths']['server_jar']).exists():
                raise FileNotFoundError(f"server.jar not found at {self.config['paths']['server_jar']}")
            
            # Stop if running
            if self.verify_server_status():
                self.logger.info("Stopping existing server process...")
                if not self.stop_server():
                    raise Exception("Failed to stop existing server")
            
            # Setup directories and logs
            log_dir = Path(self.config['paths']['minecraft']) / "logs"
            self._ensure_directory(log_dir)
            
            log_file = log_dir / "latest.log"
            if log_file.exists():
                os.remove(log_file)
            
            # Get memory and Java flags
            memory_flags = (self.config['server']['memory']['flags_sh'] 
                          if self.config['server']['memory']['source'] == 'flags_sh'
                          else f"-Xms{self.config['server']['memory']['min']} -Xmx{self.config['server']['memory']['max']}")

            java_flags = (self.config['server']['java_flags']['flags_sh']
                         if self.config['server']['java_flags']['source'] == 'flags_sh'
                         else " ".join(self.config['server']['java_flags']['custom']))
            
            # Start server
            start_cmd = (
                f"cd {self.config['paths']['minecraft']} && "
                f"nohup java {memory_flags} {java_flags} -jar {self.config['paths']['server_jar']} nogui "
                f"> logs/latest.log 2>&1 &"
            )
            
            self.logger.info("Starting server...")
            self.logger.debug(f"Start command: {start_cmd}")
            
            if os.system(start_cmd) != 0:
                raise Exception("Failed to execute start command")
            
            # Wait for startup
            startup_timeout = self.config['api']['startup_timeout']
            self.logger.info(f"Waiting up to {startup_timeout} seconds for server to start...")
            
            for i in range(startup_timeout // 5):
                if self.verify_server_status():
                    # Double check stability
                    time.sleep(2)
                    if self.verify_server_status():
                        self.logger.info("Server started successfully")
                        self.send_discord_notification("Server Status", "🟢 Server is now online!")
                        return True
                
                time.sleep(5)
                if i % 6 == 0:  # Log every 30 seconds
                    self.logger.info(f"Still waiting for server to start... ({i*5} seconds)")
                    # Log recent activity
                    lines = self._read_server_log(5)
                    if lines:
                        self.logger.debug(f"Recent log entries:\n{''.join(lines)}")
            
            raise Exception(f"Server failed to start after {startup_timeout} seconds")
            
        except Exception as e:
            self._handle_operation_error("Server Restart", e)
            return False

    #region Utility Methods
    def _get_server_pid(self) -> Optional[str]:
        """Get server process ID if running."""
        try:
            result = os.popen("ps aux | grep '[j]ava.*server.jar'").read().strip()
            if not result:
                return None
            return result.split()[1]
        except (IndexError, Exception) as e:
            self.logger.debug(f"Could not get server PID: {str(e)}")
            return None

    def _read_server_log(self, max_lines: int = 50) -> List[str]:
        """Read last N lines from server log."""
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
        """Create directory with proper permissions if it doesn't exist."""
        try:
            if not path.exists():
                os.system(f"sudo mkdir -p {path}")
                os.system(f"sudo chown -R $USER:$USER {path}")
                self.logger.info(f"Created directory: {path}")
        except Exception as e:
            self.logger.error(f"Failed to create directory {path}: {str(e)}")
            raise

    def _handle_operation_error(self, operation_name: str, error: Exception, notify: bool = True) -> None:
        """Centralized error handling with optional notification."""
        self.logger.error(f"Error in {operation_name}: {str(error)}")
        if notify:
            self.send_discord_notification(
                f"{operation_name} Failed",
                f"❌ {str(error)}",
                True
            )
    #endregion

    #region Mod Management
    async def fetch_latest_mod_versions(self) -> Dict[str, dict]:
        """Fetch latest compatible mod versions from Modrinth."""
        headers = {
            "Accept": "application/json"
        }
        mod_info = {}
        failed_mods = []

        # Connection pooling
        connector = aiohttp.TCPConnector(limit=10)
        timeout = aiohttp.ClientTimeout(total=300)
        
        progress_bar = tqdm(total=len(self.config['modrinth_urls']), desc="Checking mod versions")
        
        async with aiohttp.ClientSession(headers=headers, 
                                       connector=connector, 
                                       timeout=timeout) as session:
            # Process in chunks to avoid rate limits
            chunk_size = self.config['api']['chunk_size']
            for i in range(0, len(self.config['modrinth_urls']), chunk_size):
                chunk = self.config['modrinth_urls'][i:i + chunk_size]
                tasks = []
                
                for url in chunk:
                    tasks.append(self._fetch_mod_info(session, url, mod_info, failed_mods, progress_bar))
                
                await asyncio.gather(*tasks)
                
                if i + chunk_size < len(self.config['modrinth_urls']):
                    await asyncio.sleep(2)  # Rate limiting delay
        
        progress_bar.close()

        if failed_mods:
            self.send_discord_notification(
                "Mod Update Issues",
                "Failed to process:\n" + "\n".join(failed_mods),
                True
            )

        return mod_info

    async def _fetch_mod_info(self, session: aiohttp.ClientSession, url: str, 
                            mod_info: Dict, failed_mods: List, progress_bar: tqdm) -> None:
        """Fetch info for a single mod."""
        max_retries = self.config['api']['max_retries']
        base_delay = self.config['api']['base_delay']
        
        async def make_request(endpoint: str, retry_count: int = 0):
            try:
                await asyncio.sleep(2)
                async with session.get(endpoint) as response:
                    if response.status == 429:
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
            project_id = url.split('/')[-1]
            
            project_data = await make_request(f"https://api.modrinth.com/v2/project/{project_id}")
            if isinstance(project_data, str):
                raise Exception(f"Invalid project data: {project_data}")
            mod_name = project_data.get('title', project_id)
            
            await asyncio.sleep(2)
            
            versions = await make_request(f"https://api.modrinth.com/v2/project/{project_id}/version")
            if isinstance(versions, str):
                raise Exception(f"Invalid version data: {versions}")
            
            compatible_versions = [
                v for v in versions
                if self.config['minecraft']['version'] in v.get('game_versions', []) and
                self.config['minecraft']['modloader'].lower() in [loader.lower() for loader in v.get('loaders', [])]
            ]

            if compatible_versions:
                latest = compatible_versions[0]
                mod_info[url] = {
                    'version_id': latest['id'],
                    'version_number': latest['version_number'],
                    'download_url': latest['files'][0]['url'],
                    'filename': latest['files'][0]['filename'],
                    'project_name': mod_name
                }
            else:
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
        """Update all mods to latest versions."""
        try:
            mod_info = await self.fetch_latest_mod_versions()
            if not mod_info:
                self.send_discord_notification("Mod Updates", "✅ All mods are up to date!")
                return
            
            updated_mods = []
            skipped_mods = []
            failed_mods = []
            newly_added_mods = []
            
            progress_bar = tqdm(total=len(mod_info), desc="Downloading mods")
            
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
            
            self._send_update_summary(updated_mods, skipped_mods, failed_mods, len(mod_info))
            
        except Exception as e:
            self.logger.error(f"Error updating mods: {str(e)}")
            self.send_discord_notification("Mod Update Error", str(e), True)
            raise

    async def _update_single_mod(self, session: aiohttp.ClientSession, info: Dict,
                               updated_mods: List, skipped_mods: List, 
                               failed_mods: List, newly_added_mods: List, 
                               progress_bar: tqdm) -> None:
        """Update a single mod and track its status."""
        try:
            mod_name = info['project_name']
            current_mod_path = Path(self.config['paths']['local_mods']) / info['filename']
            
            if current_mod_path.exists():
                current_size = current_mod_path.stat().st_size
                async with session.head(info['download_url']) as response:
                    new_size = int(response.headers.get('content-length', 0))
                
                if current_size == new_size:
                    skipped_mods.append(f"• {mod_name} ({info['version_number']})")
                    progress_bar.update(1)
                    return
                
                # Remove old version
                os.system(f"sudo rm {current_mod_path}")
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
    #endregion

    #region Maintenance
    def run_maintenance(self) -> None:
        """Run complete maintenance cycle in correct order."""
        try:
            # 1. Check for players
            if self.get_player_count() > 0:
                self._warn_players()
            
            # 2. Stop server
            if self.verify_server_status():
                self.send_discord_notification("Maintenance", "🔄 Stopping server...")
                if not self.stop_server():
                    raise Exception("Failed to stop server")
            
            # 3. Backup
            if not self.create_backup():
                raise Exception("Backup failed")

            # 4. Update mods
            asyncio.run(self.update_mods())
            
            # 5. Cleanup old backups
            self.cleanup_old_backups()

            # 6. Start server
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
        """Run automated update process with safety checks."""
        try:
            if not self.verify_server_status():
                self.logger.info("Server is not running, starting it...")
                if not self.restart_server():
                    raise Exception("Failed to start server")
            
            if self.get_player_count() > 0:
                self.send_discord_notification(
                    "Server Update", 
                    "🔄 Players online - initiating shutdown countdown"
                )
                self._warn_players()
            
            self.send_discord_notification("Update Started", "🔄 Beginning update process...")
            self.run_maintenance()
            
        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            self.send_discord_notification("Update Failed", f"❌ {str(e)}", True)
    #endregion

    #region Utilities
    def send_discord_notification(self, title: str, message: str, is_error: bool = False) -> None:
        """Send Discord webhook notification."""
        if not self.config['notifications']['discord_webhook']:
            return
            
        try:
            payload = {
                "embeds": [{
                    "title": title,
                    "description": message[:1997] + "..." if len(message) > 2000 else message,
                    "color": 0xFF0000 if is_error else 0x00FF00,
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "footer": {"text": "Minecraft Mod Manager"}
                }]
            }
            
            response = requests.post(
                self.config['notifications']['discord_webhook'],
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            if response.status_code not in (200, 204):
                self.logger.error(f"Discord API error: {response.status_code}")
                
        except Exception as e:
            self.logger.error(f"Failed to send Discord notification: {str(e)}")

    def create_backup(self) -> bool:
        """Create server backup of mods and config files."""
        try:
            # Use backup name format from config
            timestamp = datetime.now().strftime(self.config['maintenance']['backup_name_format'])
            backup_path = Path(self.config['paths']['backups']) / f"{timestamp}.tar.gz"
            
            # Setup backup directory
            os.system(f"sudo mkdir -p {self.config['paths']['backups']}")
            os.system(f"sudo chown -R $USER:$USER {self.config['paths']['backups']}")
            os.system(f"sudo chmod -R a+r {self.config['paths']['minecraft']}")
            
            # Create temp directory
            temp_backup_dir = Path(self.config['paths']['backups']) / "temp_backup"
            os.system(f"sudo rm -rf {temp_backup_dir}")
            os.system(f"sudo mkdir -p {temp_backup_dir}")
            os.system(f"sudo chown -R $USER:$USER {temp_backup_dir}")
            
            # Copy files
            os.system(f"sudo cp -r {self.config['paths']['local_mods']} {temp_backup_dir}/mods")
            os.system(f"sudo cp -r {self.config['paths']['minecraft']}/config {temp_backup_dir}/config")
            os.system(f"sudo cp -r {self.config['paths']['minecraft']}/world {temp_backup_dir}/world")
            os.system(f"sudo chown -R $USER:$USER {temp_backup_dir}")
            
            # Create archive
            with tarfile.open(backup_path, 'w:gz') as tar:
                tar.add(temp_backup_dir / 'mods', arcname='mods')
                tar.add(temp_backup_dir / 'config', arcname='config')
                tar.add(temp_backup_dir / 'world', arcname='world')
            
            # Cleanup
            os.system(f"sudo rm -rf {temp_backup_dir}")
            
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
        """Remove backups older than specified days."""
        try:
            # Get retention days from config
            keep_days = self.config['maintenance']['backup_retention_days']
            now = time.time()
            
            # Update glob pattern to match backup name format
            backup_pattern = f"*{self.config['maintenance']['backup_name_format'].replace('%Y','*').replace('%m','*').replace('%d','*').replace('%H','*').replace('%M','*')}.tar.gz"
            
            for backup in Path(self.config['paths']['backups']).glob(backup_pattern):
                if backup.stat().st_mtime < now - (keep_days * 86400):
                    os.system(f"sudo rm {backup}")
                    self.logger.info(f"Removed old backup: {backup.name}")
                        
        except Exception as e:
            self.logger.error(f"Failed to cleanup old backups: {str(e)}")

    def _warn_players(self) -> None:
        """Send countdown warnings to players."""
        warnings = self.config['maintenance']['warning_intervals']
        
        try:
            for warning in warnings:
                # Update to match config format
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

    def _send_update_summary(self, updated_mods: List[str], 
                            skipped_mods: List[str], 
                            failed_mods: List[str],
                            total_mods: int) -> None:
        """Send a summary of mod updates to Discord."""
        if not (updated_mods or failed_mods):
            self.send_discord_notification(
                "Mod Updates",
                f"✅ All {total_mods} mods are up to date!"
            )
            return
        
        message = []
        if updated_mods:
            message.extend(["📦 **Updated Mods:**", *updated_mods, ""])
        if failed_mods:
            message.extend(["❌ **Failed Updates:**", *failed_mods])
        
        self.send_discord_notification(
            f"Mod Updates Summary ({len(updated_mods)} updated, {len(failed_mods)} failed)",
            "\n".join(message)
        )
    #endregion

def main():
    """Main entry point for the application."""
    parser = argparse.ArgumentParser(description='Minecraft Mod Manager')
    parser.add_argument('--auto-update', action='store_true', 
                       help='Run automated update process')
    args = parser.parse_args()

    try:
        manager = MinecraftModManager()
        
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
