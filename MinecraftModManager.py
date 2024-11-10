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
        self.logger: Optional[logging.Logger] = None
        self.setup_logging()
        
        # Config paths
        self.server_jar_path: Optional[Path] = None
        self.local_mods_path: Optional[Path] = None 
        self.backup_path: Optional[Path] = None
        self.minecraft_path: Optional[Path] = None
        
        # Config settings
        self.minecraft_version: str = "1.21.1"
        self.modloader: str = "fabric"
        self.modrinth_urls: List[str] = []
        self.discord_webhook: Optional[str] = None
        
        # Load config
        config_path = Path(config_path)
        if not config_path.is_absolute():
            config_path = Path(__file__).parent / config_path
            
        try:
            self.load_config(config_path)
            self._validate_server_jar()
        except Exception as e:
            self.logger.error(f"Initialization failed: {str(e)}")
            raise

    #region Configuration
    def load_config(self, config_path: Path) -> None:
        """Load configuration from JSON file."""
        try:
            config = json.loads(config_path.read_text())
            
            # Load paths with server-specific defaults
            paths_config = config.get('paths', {})
            self.local_mods_path = Path(paths_config.get('local_mods', '/home/Minecraft/mods')).expanduser()
            self.backup_path = Path(paths_config.get('backups', '/home/Minecraft/backups')).expanduser()
            self.minecraft_path = Path(paths_config.get('minecraft', '/home/Minecraft')).expanduser()
            self.server_jar_path = Path(paths_config.get('server_jar', 
                                             self.minecraft_path / "server.jar")).expanduser()
            
            # Load other settings
            self.minecraft_version = config.get('minecraft_version', self.minecraft_version)
            self.modloader = config.get('modloader', self.modloader).lower()
            self.modrinth_urls = config.get('modrinth_urls', [])
            self.discord_webhook = config.get('notifications', {}).get('discord_webhook')

            self._validate_config()
            self._ensure_directories()

        except Exception as e:
            if self.logger:
                self.logger.error(f"Error loading config: {str(e)}")
            raise

    def _validate_config(self) -> None:
        """Validate required configuration fields."""
        required_config = {
            'local_mods_path': 'Local mods directory',
            'minecraft_path': 'Minecraft directory',
            'modloader': 'Modloader type'
        }
        
        missing = [name for field, name in required_config.items() 
                  if not getattr(self, field, None)]
        
        if missing:
            raise ValueError(f"Missing required config: {', '.join(missing)}")
            
        # Validate modloader value
        valid_modloaders = ['fabric', 'forge', 'quilt']
        if self.modloader.lower() not in valid_modloaders:
            raise ValueError(f"Invalid modloader: {self.modloader}. Must be one of: {', '.join(valid_modloaders)}")

    def _validate_server_jar(self) -> None:
        """Validate server.jar exists in correct location."""
        if not self.server_jar_path.exists():
            raise FileNotFoundError(f"server.jar not found at {self.server_jar_path}")

    def setup_logging(self) -> None:
        """Configure logging with file and console handlers."""
        self.logger = logging.getLogger('ModManager')
        self.logger.setLevel(logging.DEBUG)
        self.logger.handlers = []
        
        # Use minecraft directory for logs
        log_path = self.minecraft_path / "logs" / "mod_manager.log" if hasattr(self, 'minecraft_path') else Path("/home/Minecraft/logs/mod_manager.log")
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
            log_file = self.minecraft_path / "logs/latest.log"
            if log_file.exists():
                lines = log_file.read_text(encoding='utf-8').splitlines()[-50:]
                for line in reversed(lines):
                    if "There are" in line and "players online" in line:
                        try:
                            return int(line.split()[2])
                        except (IndexError, ValueError):
                            continue
            return 0
            
        except Exception as e:
            self.logger.error(f"Failed to get player count: {str(e)}")
            return 0

    def verify_server_status(self) -> bool:
        """Check if server is running and responding."""
        try:
            # Check process
            result = os.popen("ps aux | grep '[j]ava.*server.jar'").read().strip()
            if not result:
                self.logger.debug("No Java process found for Minecraft server")
                return False
            
            # Get PID
            try:
                pid = result.split()[1]
                self.logger.debug(f"Found Minecraft server process with PID: {pid}")
            except IndexError:
                self.logger.debug("Could not parse PID from process list")
                return False

            # Check server log
            log_file = self.minecraft_path / "logs" / "latest.log"
            if not log_file.exists():
                self.logger.debug("Server log file not found")
                return False
                
            try:
                # Read last few lines
                with open(log_file, 'r', encoding='utf-8') as f:
                    lines = f.readlines()[-100:]
                    log_content = ''.join(lines)
                    
                    # Check startup messages
                    startup_indicators = [
                        "Done (",  # Server done loading
                        "For help, type",  # Command help message
                        "Starting minecraft server"  # Initial startup
                    ]
                    
                    server_started = any(msg in log_content for msg in startup_indicators)
                    
                    if server_started:
                        self.logger.debug("Server startup messages found in log")
                        return True
                    else:
                        self.logger.debug("No startup messages found in log")
                        return False
                    
            except Exception as e:
                self.logger.error(f"Error reading server log: {str(e)}")
                return True
                
        except Exception as e:
            self.logger.error(f"Server verification error: {str(e)}")
            return False

    def stop_server(self) -> bool:
        """Stop the Minecraft server gracefully."""
        try:
            # Get PID
            process_info = os.popen("ps aux | grep '[j]ava.*server.jar'").read().strip()
            if not process_info:
                self.logger.info("Server is already stopped")
                return True
            
            try:
                pid = process_info.split()[1]
                self.logger.info(f"Found server process with PID: {pid}")
            except IndexError:
                self.logger.warning("Could not find server PID")
                return False
            
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
            self.logger.error(f"Error stopping server: {str(e)}")
            self.send_discord_notification("Server Status", f"❌ Server shutdown failed: {str(e)}", True)
            return False

    def _build_java_command(self) -> str:
        """Build Java command from config."""
        java_config = self.config.get('server', {}).get('java', {})
        
        # Get memory settings
        min_mem = java_config.get('min_memory', '4G')
        max_mem = java_config.get('max_memory', '6G')
        
        # Get additional flags
        flags = java_config.get('flags', [])
        
        # Build command
        cmd = [
            f"-Xms{min_mem}",
            f"-Xmx{max_mem}"
        ]
        cmd.extend(flags)
        
        return ' '.join(cmd)

    def restart_server(self) -> bool:
        """Restart the Minecraft server."""
        try:
            if not self.server_jar_path.exists():
                raise FileNotFoundError(f"server.jar not found at {self.server_jar_path}")
            
            # Stop if running
            if self.verify_server_status():
                self.logger.info("Stopping existing server process...")
                if not self.stop_server():
                    raise Exception("Failed to stop existing server")
            
            # Setup directories and logs
            os.makedirs(self.minecraft_path / "logs", exist_ok=True)
            
            log_file = self.minecraft_path / "logs" / "latest.log"
            if log_file.exists():
                os.remove(log_file)
            
            # Get Java flags from config
            java_flags = self._build_java_command()
            
            # Start server
            start_cmd = (
                f"cd {self.minecraft_path} && "
                f"nohup java {java_flags} -jar {self.server_jar_path} nogui "
                f"> logs/latest.log 2>&1 &"
            )
            
            self.logger.info("Starting server...")
            if os.system(start_cmd) != 0:
                raise Exception("Failed to execute start command")
            
            # Wait for startup
            time.sleep(5)
            
            startup_timeout = 120  # 2 minutes
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
                    try:
                        log_file = self.minecraft_path / "logs" / "latest.log"
                        if log_file.exists():
                            with open(log_file, 'r') as f:
                                last_lines = f.readlines()[-5:]
                            self.logger.debug(f"Recent log entries:\n{''.join(last_lines)}")
                    except Exception as e:
                        self.logger.error(f"Could not read log file: {str(e)}")
            
            raise Exception(f"Server failed to start after {startup_timeout} seconds")
            
        except Exception as e:
            self.logger.error(f"Error restarting server: {str(e)}")
            self.send_discord_notification("Server Status", f"❌ Server startup failed: {str(e)}", True)
            return False
    #endregion

    #region Mod Management
    async def fetch_latest_mod_versions(self) -> Dict[str, dict]:
        """Fetch latest compatible mod versions from Modrinth."""
        headers = {
            "User-Agent": "MinecraftModManager/1.0",
            "Accept": "application/json"
        }
        mod_info = {}
        failed_mods = []

        # Connection pooling
        connector = aiohttp.TCPConnector(limit=10)
        timeout = aiohttp.ClientTimeout(total=300)
        
        progress_bar = tqdm(total=len(self.modrinth_urls), desc="Checking mod versions")
        
        async with aiohttp.ClientSession(headers=headers, 
                                       connector=connector, 
                                       timeout=timeout) as session:
            # Process in chunks to avoid rate limits
            chunk_size = 10
            for i in range(0, len(self.modrinth_urls), chunk_size):
                chunk = self.modrinth_urls[i:i + chunk_size]
                tasks = []
                
                for url in chunk:
                    tasks.append(self._fetch_mod_info(session, url, mod_info, failed_mods, progress_bar))
                
                await asyncio.gather(*tasks)
                
                if i + chunk_size < len(self.modrinth_urls):
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
        max_retries = 5
        base_delay = 3
        
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
                if self.minecraft_version in v.get('game_versions', []) and
                self.modloader in [loader.lower() for loader in v.get('loaders', [])]
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
                    f"Required MC version: {self.minecraft_version}, "
                    f"Required modloader: {self.modloader}, "
                    f"Available versions: {[v.get('game_versions', []) for v in versions[:3]]}"
                )
                failed_mods.append(
                    f"{mod_name} (no compatible version for MC {self.minecraft_version} with {self.modloader})"
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
            backup_dir = self.local_mods_path / "mods_backup"
            backup_dir.mkdir(exist_ok=True)
            
            mod_info = await self.fetch_latest_mod_versions()
            if not mod_info:
                self.send_discord_notification("Mod Updates", "✅ All mods are up to date!")
                return
            
            updated_mods = []
            skipped_mods = []
            failed_mods = []
            
            progress_bar = tqdm(total=len(mod_info), desc="Downloading mods")
            
            async with aiohttp.ClientSession() as session:
                tasks = []
                for url, info in mod_info.items():
                    tasks.append(self._update_single_mod(session, info, backup_dir, 
                                                       updated_mods, skipped_mods, failed_mods, progress_bar))
                await asyncio.gather(*tasks)
            
            progress_bar.close()
            
            self._send_update_summary(updated_mods, skipped_mods, failed_mods, len(mod_info))
            
        except Exception as e:
            self.logger.error(f"Error updating mods: {str(e)}")
            self.send_discord_notification("Mod Update Error", str(e), True)
            raise

    async def _update_single_mod(self, session: aiohttp.ClientSession, info: Dict, 
                               backup_dir: Path, updated_mods: List, skipped_mods: List, 
                               failed_mods: List, progress_bar: tqdm) -> None:
        """Update a single mod and track its status."""
        try:
            mod_name = info['project_name']
            current_mod_path = self.local_mods_path / info['filename']
            
            if current_mod_path.exists():
                current_size = current_mod_path.stat().st_size
                async with session.head(info['download_url']) as response:
                    new_size = int(response.headers.get('content-length', 0))
                
                if current_size == new_size:
                    skipped_mods.append(f"• {mod_name} ({info['version_number']})")
                    progress_bar.update(1)
                    return
            
            if current_mod_path.exists():
                backup_name = f"{mod_name}-{info['version_number']}.jar.backup"
                current_mod_path.rename(backup_dir / backup_name)
            
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
                    "Update Skipped", 
                    "⚠️ Players online - will try again later"
                )
                return
            
            self.send_discord_notification("Update Started", "🔄 Beginning update process...")
            self.run_maintenance()
            
        except Exception as e:
            self.logger.error(f"Automated update failed: {str(e)}")
            self.send_discord_notification("Update Failed", f"❌ {str(e)}", True)
    #endregion

    #region Utilities
    def send_discord_notification(self, title: str, message: str, is_error: bool = False) -> None:
        """Send Discord webhook notification."""
        if not self.discord_webhook:
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
                self.discord_webhook,
                json=payload,
                headers={'Content-Type': 'application/json'}
            )
            
            if response.status_code not in (200, 204):
                self.logger.error(f"Discord API error: {response.status_code}")
                
        except Exception as e:
            self.logger.error(f"Failed to send Discord notification: {str(e)}")

    def _ensure_directories(self) -> None:
        """Ensure all required directories exist."""
        for directory in [self.local_mods_path, self.backup_path, self.minecraft_path]:
            if directory and not directory.exists():
                os.system(f"sudo mkdir -p {directory}")
                os.system(f"sudo chown -R $USER:$USER {directory}")
                self.logger.info(f"Created directory: {directory}")

    def create_backup(self) -> bool:
        """Create server backup of mods and config files."""
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            backup_path = self.backup_path / f"server_backup_{timestamp}.tar.gz"
            
            # Setup backup directory
            os.system(f"sudo mkdir -p {self.backup_path}")
            os.system(f"sudo chown -R $USER:$USER {self.backup_path}")
            os.system(f"sudo chmod -R a+r {self.minecraft_path}")
            
            # Create temp directory
            temp_backup_dir = self.backup_path / "temp_backup"
            os.system(f"sudo rm -rf {temp_backup_dir}")
            os.system(f"sudo mkdir -p {temp_backup_dir}")
            os.system(f"sudo chown -R $USER:$USER {temp_backup_dir}")
            
            # Copy files
            os.system(f"sudo cp -r {self.local_mods_path} {temp_backup_dir}/mods")
            os.system(f"sudo cp -r {self.minecraft_path}/config {temp_backup_dir}/config")
            os.system(f"sudo cp -r {self.minecraft_path}/world {temp_backup_dir}/world")
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
                f"✅ Created backup: server_backup_{timestamp}.tar.gz"
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

    def cleanup_old_backups(self, keep_days: int = 7) -> None:
        """Remove backups older than specified days."""
        try:
            now = time.time()
            for backup in self.backup_path.glob('server_backup_*.tar.gz'):
                if backup.stat().st_mtime < now - (keep_days * 86400):
                    os.system(f"sudo rm {backup}")
                    self.logger.info(f"Removed old backup: {backup.name}")
                        
        except Exception as e:
            self.logger.error(f"Failed to cleanup old backups: {str(e)}")

    def _warn_players(self) -> None:
        """Send countdown warnings to players."""
        warnings = [
            (15, "minutes"),
            (10, "minutes"),
            (5, "minutes"),
            (1, "minute"),
            (30, "seconds"),
            (10, "seconds"),
            (5, "seconds")
        ]
        
        try:
            for time_val, unit in warnings:
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
        notification = []
        
        if updated_mods:
            notification.append("📦 **Updated Mods:**")
            notification.extend(updated_mods)
            notification.append("")
            
        if failed_mods:
            notification.append("❌ **Failed Updates:**")
            notification.extend(failed_mods)
        
        if updated_mods or failed_mods:
            self.send_discord_notification(
                f"Mod Updates Summary ({len(updated_mods)} updated, {len(failed_mods)} failed)",
                "\n".join(notification)
            )
        else:
            self.send_discord_notification(
                "Mod Updates",
                f"✅ All {total_mods} mods are up to date!"
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
