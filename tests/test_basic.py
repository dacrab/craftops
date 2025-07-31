"""Basic tests for minecraft-mod-manager."""

from pathlib import Path

import pytest

from minecraft_mod_manager.core import ModManager
from minecraft_mod_manager.config.config import Config


def test_config_creation():
    """Test configuration creation and validation."""
    config_data = {
        'minecraft': {'version': '1.20.1', 'modloader': 'fabric'},
        'paths': {
            'server': '/tmp/test/server',
            'mods': '/tmp/test/mods',
            'backups': '/tmp/test/backups',
            'logs': '/tmp/test/logs/mod-manager.log'
        },
        'server': {
            'jar': 'server.jar',
            'java_flags': ['-Xmx2G'],
            'stop_command': 'stop',
            'max_stop_wait': 300
        },
        'backup': {
            'max_mod_backups': 3,
            'name_format': '%Y%m%d_%H%M%S'
        },
        'notifications': {
            'discord_webhook': '',
            'warning_template': 'Server will restart in {minutes} minutes',
            'warning_intervals': [15, 10, 5, 1]
        },
        'mods': {
            'chunk_size': 5,
            'max_retries': 3,
            'base_delay': 2,
            'sources': {
                'modrinth': [],
                'curseforge': []
            }
        }
    }
    
    config = Config.from_dict(config_data)
    
    # Test basic properties
    assert config.minecraft_version == "1.20.1"
    assert config.modloader == "fabric"
    assert config.server_path == "/tmp/test/server"
    assert config.mods_path == "/tmp/test/mods"
    
    # Test legacy compatibility
    assert config.minecraft.version == "1.20.1"
    assert config.server.jar == "server.jar"
    assert config.server.start_command == "java -Xmx2G -jar server.jar nogui"


@pytest.mark.asyncio
async def test_mod_manager_context(test_config, logger):
    """Test ModManager async context management."""
    manager = ModManager(test_config, logger)

    async with manager:
        assert manager.session is not None
        assert manager.mods_dir.exists()

    assert manager.session is None


def test_mod_manager_initialization(test_config, logger):
    """Test ModManager initialization."""
    manager = ModManager(test_config, logger)
    
    assert manager.config == test_config
    assert manager.logger == logger
    assert manager.session is None
    assert manager.mods_dir.exists()


def test_config_legacy_compatibility(test_config):
    """Test that legacy config access still works."""
    # Test paths access
    assert hasattr(test_config, 'paths')
    assert test_config.paths.server == test_config.server_path
    assert test_config.paths.mods == test_config.mods_path
    
    # Test server access
    assert hasattr(test_config, 'server')
    assert test_config.server.jar == test_config.server_jar
    
    # Test minecraft access
    assert hasattr(test_config, 'minecraft')
    assert test_config.minecraft.version == test_config.minecraft_version
