"""Server process control and status monitoring."""

import logging
import re
import subprocess
import time
from pathlib import Path
from typing import List, Optional, Dict

from ..config.config import Config

class ServerController:
    """Handles server process control and status monitoring."""
    
    MODLOADER_JARS = {
        'fabric': 'fabric-server-launch.jar',
        'forge': 'forge-server.jar',
        'quilt': 'quilt-server-launch.jar'
    }
    
    def __init__(self, config: Config, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.minecraft_dir = Path(config['paths']['minecraft'])
        self.server_jar = Path(config['paths']['server_jar'])
        self.process: Optional[subprocess.Popen] = None
        self.modloader = config['minecraft']['modloader'].lower()
        
        # Validate modloader-specific requirements
        self._validate_modloader_setup()
    
    def _validate_modloader_setup(self) -> None:
        """Validate modloader-specific requirements."""
        if not self.server_jar.exists():
            raise RuntimeError(f"Server JAR not found: {self.server_jar}")
            
        modloader_jar = self.minecraft_dir / self.MODLOADER_JARS.get(self.modloader, '')
        if self.modloader != 'vanilla' and not modloader_jar.exists():
            raise RuntimeError(
                f"Modloader JAR not found: {modloader_jar}\n"
                f"Please install {self.modloader} for Minecraft {self.config['minecraft']['version']}"
            )
    
    def verify_status(self) -> bool:
        """Check if server process is running."""
        if not self.process:
            return False
            
        return self.process.poll() is None
    
    def control(self, action: str) -> bool:
        """Control server process (start/stop/restart)."""
        action = action.lower()
        
        if action not in ('start', 'stop', 'restart'):
            self.logger.error(f"Invalid server control action: {action}")
            return False
        
        try:
            if action in ('stop', 'restart'):
                self._stop_server()
            
            if action in ('start', 'restart'):
                self._start_server()
            
            return True
            
        except Exception as e:
            self.logger.error(f"Server control error ({action}): {str(e)}")
            return False
    
    def _get_server_flags(self) -> List[str]:
        """Get server startup flags based on configuration."""
        flags = ['java']
        
        # Add memory settings
        if 'memory' in self.config['server']:
            memory = self.config['server']['memory']
            flags.extend([
                f"-Xms{memory['min']}",
                f"-Xmx{memory['max']}"
            ])
        
        # Add custom flags if configured
        if 'flags_source' in self.config['server']:
            if self.config['server']['flags_source'] == 'custom':
                custom_flags = self.config['server']['custom_flags'].split()
                # Remove 'java' if it's the first flag
                if custom_flags and custom_flags[0].lower() == 'java':
                    custom_flags = custom_flags[1:]
                flags.extend(custom_flags)
            else:
                # Add default JVM flags
                flags.extend(self.config['server'].get('java_flags', []))
        
        # Add server jar
        flags.extend(['-jar', str(self.server_jar), 'nogui'])
        
        return flags
    
    def _start_server(self) -> None:
        """Start server process."""
        if self.verify_status():
            self.logger.warning("Server is already running")
            return
        
        # Get server flags
        flags = self._get_server_flags()
        
        # Start process
        self.process = subprocess.Popen(
            flags,
            cwd=self.minecraft_dir,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            universal_newlines=True
        )
        
        self.logger.info(f"Starting {self.modloader} server for Minecraft {self.config['minecraft']['version']}")
        
        # Wait for server to start
        self._wait_for_startup()
    
    def _stop_server(self) -> None:
        """Stop server process."""
        if not self.verify_status() or not self.process:
            self.logger.warning("Server is not running")
            return
        
        # Send SIGTERM
        self.process.terminate()
        
        try:
            # Wait for graceful shutdown
            self.process.wait(timeout=30)
        except subprocess.TimeoutExpired:
            # Force kill if timeout
            self.process.kill()
            self.process.wait()
        
        self.logger.info("Server process stopped")
        self.process = None
    
    def _wait_for_startup(self, timeout: int = 300) -> None:
        """Wait for server to complete startup."""
        if not self.process or not self.process.stdout:
            return
            
        start_time = time.time()
        done_pattern = re.compile(r'Done \([0-9.]+s\)!')
        
        while True:
            line = self.process.stdout.readline()
            
            if not line:
                # Check if process died
                if self.process.poll() is not None:
                    raise RuntimeError("Server process died during startup")
                continue
            
            # Check for completion
            if done_pattern.search(line):
                self.logger.info("Server startup complete")
                return
            
            # Check timeout
            if time.time() - start_time > timeout:
                raise TimeoutError("Server startup timed out")
    
    def get_player_count(self) -> int:
        """Get number of currently online players."""
        # TODO: Implement actual player count checking via server query
        # For now just return 0 if server is not running
        if not self.verify_status():
            return 0
        return 0