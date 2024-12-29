"""
Minecraft Mod Manager - A comprehensive Minecraft server management tool.

This package provides functionality for:
- Automated mod updates via Modrinth API
- Server process control (start/stop/restart)
- Backup creation and rotation
- Discord notifications
- Player warnings for maintenance
"""

from .config.config import Config
from .controllers.server import ServerController
from .managers.backup import BackupManager
from .managers.mod import ModManager
from .managers.notification import NotificationManager
from .minecraft_mod_manager import MinecraftModManager

__version__ = "1.0.0"
__author__ = "dacrab"
__license__ = "MIT"

__all__ = [
    "Config",
    "ServerController",
    "BackupManager",
    "ModManager",
    "NotificationManager",
    "MinecraftModManager",
] 