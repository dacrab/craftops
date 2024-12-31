"""Server controller module."""

import logging
import subprocess
import time
from pathlib import Path
from typing import Optional, Protocol

from ..config.config import Config


class ServerControllerProtocol(Protocol):
    """Protocol defining the interface for ServerController."""
    def start(self) -> bool: ...
    def stop(self) -> bool: ...
    def verify_status(self) -> bool: ...


class ServerController:
    """Controls the Minecraft server process."""

    def __init__(self, config: 'Config', logger: Optional[logging.Logger] = None) -> None:
        """
        Initialize the server controller.

        Args:
            config: Configuration object
            logger: Optional logger instance
        """
        self.config = config
        self.logger = logger or logging.getLogger(__name__)
        self.process: Optional[subprocess.Popen] = None

    def start(self) -> bool:
        """
        Start the Minecraft server.

        Returns:
            bool: True if server started successfully, False otherwise
        """
        try:
            if self.verify_status():
                self.logger.warning("Server is already running")
                return True

            self.logger.info("Starting server...")
            self.process = subprocess.Popen(
                self.config.server.start_command.split(),
                cwd=self.config.paths.server,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )

            # Wait for server to start
            time.sleep(5)
            if self.verify_status():
                self.logger.info("Server started successfully")
                return True

            self.logger.error("Server failed to start")
            return False

        except Exception as e:
            self.logger.error(f"Error starting server: {str(e)}")
            return False

    def stop(self) -> bool:
        """
        Stop the Minecraft server.

        Returns:
            bool: True if server stopped successfully, False otherwise
        """
        try:
            if not self.verify_status():
                self.logger.warning("Server is not running")
                return True

            self.logger.info("Stopping server...")
            subprocess.run(
                f'screen -S minecraft -X stuff "{self.config.server.stop_command}^M"',
                shell=True,
                check=True
            )

            # Wait for server to stop
            start_time = time.time()
            while self.verify_status():
                if time.time() - start_time > self.config.server.max_stop_wait:
                    self.logger.error("Server failed to stop within timeout")
                    return False
                time.sleep(1)

            self.logger.info("Server stopped successfully")
            return True

        except Exception as e:
            self.logger.error(f"Error stopping server: {str(e)}")
            return False

    def verify_status(self) -> bool:
        """
        Check if the server is running.

        Returns:
            bool: True if server is running, False otherwise
        """
        try:
            if self.process and self.process.poll() is None:
                return True

            # Check for running screen session
            result = subprocess.run(
                "screen -ls | grep minecraft",
                shell=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            return result.returncode == 0

        except Exception as e:
            self.logger.error(f"Error checking server status: {str(e)}")
            return False