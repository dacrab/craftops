"""Configuration handling module."""

from dataclasses import dataclass
from typing import Any, Dict, List

from ..utils import toml_utils


@dataclass
class PathsConfig:
    """Paths configuration."""
    server: str
    mods: str
    backups: str
    logs: str


@dataclass
class ServerConfig:
    """Server configuration."""
    jar: str
    java_flags: List[str]
    stop_command: str
    status_check_interval: int
    max_stop_wait: int

    @property
    def start_command(self) -> str:
        """Generate the server start command from Java flags."""
        return f"java {' '.join(self.java_flags)} -jar {self.jar} nogui"


@dataclass
class BackupConfig:
    """Backup configuration."""
    max_mod_backups: int
    mod_backup_dir: str
    name_format: str


@dataclass
class NotificationsConfig:
    """Notifications configuration."""
    discord_webhook: str
    warning_template: str
    warning_intervals: List[int]


@dataclass
class MinecraftConfig:
    """Minecraft configuration."""
    version: str
    modloader: str


@dataclass
class ModSourcesConfig:
    """Mod sources configuration."""
    modrinth: List[str]
    curseforge: List[str]


@dataclass
class ModsConfig:
    """Mods configuration."""
    auto_update: bool
    backup_before_update: bool
    notify_before_update: bool
    update_check_interval: int
    chunk_size: int
    max_retries: int
    base_delay: int
    sources: ModSourcesConfig


@dataclass
class Config:
    """Main configuration class."""
    paths: PathsConfig
    server: ServerConfig
    backup: BackupConfig
    notifications: NotificationsConfig
    minecraft: MinecraftConfig
    mods: ModsConfig

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Config':
        """Create Config instance from dictionary."""
        return cls(
            paths=PathsConfig(**data['paths']),
            server=ServerConfig(**data['server']),
            backup=BackupConfig(**data['backup']),
            notifications=NotificationsConfig(**data['notifications']),
            minecraft=MinecraftConfig(**data['minecraft']),
            mods=ModsConfig(
                **{k: v for k, v in data['mods'].items() if k != 'sources'},
                sources=ModSourcesConfig(**data['mods']['sources'])
            )
        )


def load_config(config_path: str) -> Config:
    """Load configuration from file."""
    data = toml_utils.load_toml(config_path)
    return Config.from_dict(data)