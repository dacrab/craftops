"""Configuration handling module."""

import logging
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    import tomllib  # type: ignore[import-not-found]
except ImportError:
    import tomli as tomllib  # type: ignore[import-not-found]

logger = logging.getLogger(__name__)


@dataclass
class Config:
    """Main configuration class with simplified structure."""

    # Paths
    server_path: str
    mods_path: str
    backups_path: str
    logs_path: str

    # Server settings
    server_jar: str
    java_flags: List[str]
    stop_command: str = "stop"
    max_stop_wait: int = 300

    # Minecraft settings
    minecraft_version: str = "1.20.1"
    modloader: str = "fabric"

    # Mod settings
    chunk_size: int = 5
    max_retries: int = 3
    base_delay: int = 2
    modrinth_mods: Optional[List[str]] = None
    curseforge_mods: Optional[List[str]] = None

    # Backup settings
    max_backups: int = 3
    backup_name_format: str = "%Y%m%d_%H%M%S"

    # Notification settings
    discord_webhook: str = ""
    warning_template: str = "Server will restart in {minutes} minutes for updates"
    warning_intervals: Optional[List[int]] = None

    def __post_init__(self) -> None:
        """Initialize default values for mutable fields."""
        if self.modrinth_mods is None:
            self.modrinth_mods = []
        if self.curseforge_mods is None:
            self.curseforge_mods = []
        if self.warning_intervals is None:
            self.warning_intervals = [15, 10, 5, 1]

    @property
    def start_command(self) -> str:
        """Generate the server start command from Java flags."""
        return f"java {' '.join(self.java_flags)} -jar {self.server_jar} nogui"

    # Legacy compatibility properties
    @property
    def paths(self) -> Any:
        """Legacy paths access."""
        return type('Paths', (), {
            'server': self.server_path,
            'mods': self.mods_path,
            'backups': self.backups_path,
            'logs': self.logs_path
        })()

    @property
    def server(self) -> Any:
        """Legacy server access."""
        return type('Server', (), {
            'jar': self.server_jar,
            'java_flags': self.java_flags,
            'stop_command': self.stop_command,
            'max_stop_wait': self.max_stop_wait,
            'start_command': self.start_command
        })()

    @property
    def minecraft(self) -> Any:
        """Legacy minecraft access."""
        return type('Minecraft', (), {
            'version': self.minecraft_version,
            'modloader': self.modloader
        })()

    @property
    def mods(self) -> Any:
        """Legacy mods access."""
        sources = type('Sources', (), {
            'modrinth': self.modrinth_mods,
            'curseforge': self.curseforge_mods
        })()

        return type('Mods', (), {
            'chunk_size': self.chunk_size,
            'max_retries': self.max_retries,
            'base_delay': self.base_delay,
            'sources': sources
        })()

    @property
    def backup(self) -> Any:
        """Legacy backup access."""
        return type('Backup', (), {
            'max_mod_backups': self.max_backups,
            'name_format': self.backup_name_format
        })()

    @property
    def notifications(self) -> Any:
        """Legacy notifications access."""
        return type('Notifications', (), {
            'discord_webhook': self.discord_webhook,
            'warning_template': self.warning_template,
            'warning_intervals': self.warning_intervals
        })()

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Config':
        """Create Config instance from dictionary."""
        # Extract nested values with defaults
        paths = data.get('paths', {})
        server = data.get('server', {})
        minecraft = data.get('minecraft', {})
        mods = data.get('mods', {})
        sources = mods.get('sources', {})
        backup = data.get('backup', {})
        notifications = data.get('notifications', {})

        return cls(
            # Paths
            server_path=paths.get('server', '/path/to/minecraft/server'),
            mods_path=paths.get('mods', '/path/to/minecraft/server/mods'),
            backups_path=paths.get('backups', '/path/to/minecraft/backups'),
            logs_path=paths.get('logs', '/path/to/minecraft/logs/mod-manager.log'),

            # Server
            server_jar=server.get('jar', 'server.jar'),
            java_flags=server.get('java_flags', ['-Xms4G', '-Xmx4G']),
            stop_command=server.get('stop_command', 'stop'),
            max_stop_wait=server.get('max_stop_wait', 300),

            # Minecraft
            minecraft_version=minecraft.get('version', '1.20.1'),
            modloader=minecraft.get('modloader', 'fabric'),

            # Mods
            chunk_size=mods.get('chunk_size', 5),
            max_retries=mods.get('max_retries', 3),
            base_delay=mods.get('base_delay', 2),
            modrinth_mods=sources.get('modrinth', []),
            curseforge_mods=sources.get('curseforge', []),

            # Backup
            max_backups=backup.get('max_mod_backups', 3),
            backup_name_format=backup.get('name_format', '%Y%m%d_%H%M%S'),

            # Notifications
            discord_webhook=notifications.get('discord_webhook', ''),
            warning_template=notifications.get(
                'warning_template', 'Server will restart in {minutes} minutes for updates'
            ),
            warning_intervals=notifications.get('warning_intervals', [15, 10, 5, 1])
        )

    def validate(self) -> bool:
        """Validate configuration and create necessary directories."""
        errors = []

        # Validate paths
        for path_name, path_value in [
            ('server', self.server_path),
            ('mods', self.mods_path),
            ('backups', self.backups_path)
        ]:
            path = Path(path_value)
            if path_name == 'server' and not path.exists():
                errors.append(f"Server directory does not exist: {path}")
            else:
                # Create directory if it doesn't exist
                path.mkdir(parents=True, exist_ok=True)

        # Validate server jar
        jar_path = Path(self.server_path) / self.server_jar
        if not jar_path.exists() and Path(self.server_path).exists():
            errors.append(f"Server jar file not found: {jar_path}")

        # Validate mod sources
        total_mods = len(self.modrinth_mods or []) + len(self.curseforge_mods or [])
        if total_mods == 0:
            logger.warning("No mod sources configured")

        # Log errors
        if errors:
            for error in errors:
                logger.error(f"Config error: {error}")
            return False

        logger.info("Configuration validation passed")
        return True


def setup_logging(log_file: Optional[str] = None, level: int = logging.INFO) -> None:
    """Setup logging configuration."""
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')

    # Console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setFormatter(formatter)

    # Root logger
    root_logger = logging.getLogger()
    root_logger.setLevel(level)
    root_logger.addHandler(console_handler)

    # File handler if specified
    if log_file:
        log_path = Path(log_file)
        log_path.parent.mkdir(parents=True, exist_ok=True)

        file_handler = logging.FileHandler(log_file)
        file_handler.setFormatter(formatter)
        root_logger.addHandler(file_handler)


def load_config(config_path: str) -> Config:
    """Load configuration from file."""
    try:
        with open(config_path, 'rb') as f:
            data = tomllib.load(f)

        config = Config.from_dict(data)

        # Setup logging
        setup_logging(config.logs_path)

        # Validate configuration
        if not config.validate():
            raise ValueError("Configuration validation failed")

        return config

    except FileNotFoundError:
        logger.error(f"Configuration file not found: {config_path}")
        raise
    except Exception as e:
        logger.error(f"Error loading configuration: {str(e)}")
        raise
