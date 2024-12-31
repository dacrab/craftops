"""Basic tests for minecraft-mod-manager."""

import pytest
from pathlib import Path

from minecraft_mod_manager.managers.mod import ModManager
from minecraft_mod_manager.utils import toml_utils

def test_load_toml(test_data_dir):
    """Test TOML loading functionality."""
    test_data = {
        'minecraft': {
            'version': '1.20.1',
            'modloader': 'fabric'
        },
        'paths': {
            'server': str(test_data_dir / 'server'),
            'mods': str(test_data_dir / 'mods'),
            'backups': str(test_data_dir / 'backups'),
            'logs': str(test_data_dir / 'logs/mod-manager.log')
        }
    }
    
    # Create test file
    test_file = test_data_dir / 'test_config.toml'
    toml_utils.save_toml(str(test_file), test_data)
    
    # Load and verify
    loaded_data = toml_utils.load_toml(str(test_file))
    assert loaded_data['minecraft']['version'] == '1.20.1'
    assert loaded_data['minecraft']['modloader'] == 'fabric'
    assert loaded_data['paths']['server'] == str(test_data_dir / 'server')
    
    # Cleanup
    test_file.unlink()

@pytest.mark.asyncio
async def test_mod_manager_context(test_config, logger):
    """Test ModManager async context management."""
    manager = ModManager(test_config, logger)
    
    async with manager:
        assert manager.session is not None
        assert manager.mods_dir.exists()
    
    assert manager.session is None

def test_config_validation(test_config, test_data_dir):
    """Test configuration validation."""
    assert test_config.minecraft.version == "1.20.1"
    assert test_config.minecraft.modloader == "fabric"
    assert Path(test_config.paths.server) == test_data_dir / "server"
    assert test_config.mods.auto_update is True
    assert len(test_config.mods.sources.modrinth) == 0
    assert len(test_config.mods.sources.curseforge) == 0

def test_server_command(test_config):
    """Test server command generation."""
    assert test_config.server.start_command == "java -Xmx2G -jar server.jar nogui" 