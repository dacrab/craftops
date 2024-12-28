"""Constants used throughout the application."""

from pathlib import Path
from typing import Final

# Default paths
DEFAULT_CONFIG_PATH: Final[str] = str(Path.home() / ".config" / "minecraft-mod-manager" / "config.jsonc")

# API timeouts
DEFAULT_TIMEOUT: Final[int] = 30  # seconds

# Discord message settings
DISCORD_MAX_LENGTH: Final[int] = 2000
DISCORD_SUCCESS_COLOR: Final[int] = 0x00FF00  # Green
DISCORD_ERROR_COLOR: Final[int] = 0xFF0000    # Red
DISCORD_FOOTER_TEXT: Final[str] = "Minecraft Mod Manager"

# Warning intervals
WARNING_SLEEP_MINUTES: Final[int] = 60  # Sleep time between minute warnings
WARNING_SLEEP_SECONDS: Final[int] = 10  # Sleep time between second warnings

# Server settings
DEFAULT_MEMORY: Final[str] = "2G"
DEFAULT_MAX_WAIT: Final[int] = 120
DEFAULT_CHUNK_SIZE: Final[int] = 10
DEFAULT_MAX_RETRIES: Final[int] = 3
DEFAULT_BASE_DELAY: Final[int] = 2

# Server process settings
SHUTDOWN_TIMEOUT: Final[int] = 20  # seconds to wait for graceful shutdown
STARTUP_CHECK_INTERVAL: Final[int] = 2  # seconds between startup checks
LOG_CHECK_LINES: Final[int] = 100  # number of log lines to check for startup
PLAYER_CHECK_LINES: Final[int] = 50  # number of log lines to check for player count 