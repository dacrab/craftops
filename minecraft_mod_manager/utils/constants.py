"""Constants used throughout the Minecraft Mod Manager."""

# Configuration
DEFAULT_CONFIG_PATH = "config.jsonc"

# Server settings
DEFAULT_MEMORY = "2G"
DEFAULT_MAX_WAIT = 120
DEFAULT_CHUNK_SIZE = 10
DEFAULT_TIMEOUT = 300
DEFAULT_MAX_RETRIES = 3
DEFAULT_BASE_DELAY = 2

# Discord webhook settings
DISCORD_SUCCESS_COLOR = 0x00FF00
DISCORD_ERROR_COLOR = 0xFF0000
DISCORD_MAX_LENGTH = 2000
DISCORD_FOOTER_TEXT = "Minecraft Mod Manager"

# Warning intervals
WARNING_SLEEP_MINUTES = 60
WARNING_SLEEP_SECONDS = 5

# Server process settings
SHUTDOWN_TIMEOUT = 20  # seconds to wait for graceful shutdown
STARTUP_CHECK_INTERVAL = 2  # seconds between startup checks
LOG_CHECK_LINES = 100  # number of log lines to check for startup
PLAYER_CHECK_LINES = 50  # number of log lines to check for player count 