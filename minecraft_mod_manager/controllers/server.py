"""Server process control and status monitoring."""

import logging
import re
import subprocess
import time
from pathlib import Path
from typing import Optional

from ..config.config import Config


class ServerController:
    """Handles server process control and status monitoring."""
    
    def __init__(self, config: Config, logger: logging.Logger):
        self.config = config
        self.logger = logger
        self.minecraft_dir = Path(config['paths']['minecraft'])
        self.server_jar = Path(config['paths']['server_jar'])
        self.process: Optional[subprocess.Popen] = None
    
    def verify_status(self) -> bool:
        """Check if server process is running."""
        if not self.process:
            return False
            
        return self.process.poll() is None
    
    def get_player_count(self) -> int:
        """Get number of online players."""
        # TODO: Implement RCON support for accurate player count
        return 0
    
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
        
        self.logger.info("Server process started")
        
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
    
    def _get_server_flags(self) -> list[str]:
        """Get Java flags for server process."""
        if self.config['server']['flags_source'] == 'custom':
            # Use custom flags
            flags_str = self.config['server']['custom_flags']
        else:
            # Use default flags
            memory = self.config['server'].get('memory', '4G')
            flags_str = (
                f"java -Xms{memory} -Xmx{memory} "
                "-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 "
                "-XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC "
                "-XX:+AlwaysPreTouch -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40 "
                "-XX:G1HeapRegionSize=8M -XX:G1ReservePercent=20 -XX:G1HeapWastePercent=5 "
                "-XX:G1MixedGCCountTarget=4 -XX:InitiatingHeapOccupancyPercent=15 "
                "-XX:G1MixedGCLiveThresholdPercent=90 -XX:G1RSetUpdatingPauseTimePercent=5 "
                "-XX:SurvivorRatio=32 -XX:+PerfDisableSharedMem -XX:MaxTenuringThreshold=1"
            )
        
        # Split into list and append jar
        flags = flags_str.split()
        flags.extend(['-jar', str(self.server_jar), 'nogui'])
        
        return flags
    
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