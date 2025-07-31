"""Minecraft Mod Manager - A tool to manage Minecraft mods across multiple servers."""

from typing import Final

from .core import (
    BackupError,
    BackupManager,
    MinecraftModManagerError,
    ModManager,
    ModUpdateError,
    NotificationManager,
    ServerController,
    ServerError,
)
from .minecraft_mod_manager import MinecraftModManager

__version__: Final[str] = "1.0.0"
__author__: Final[str] = "dacrab"
__license__: Final[str] = "MIT"

__all__ = [
    "__version__",
    "__author__",
    "__license__",
    "MinecraftModManager",
    "BackupManager",
    "ModManager",
    "NotificationManager",
    "ServerController",
    "MinecraftModManagerError",
    "ModUpdateError",
    "BackupError",
    "ServerError",
]
