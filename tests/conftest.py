"""Pytest configuration for minecraft-mod-manager tests."""

import logging
from pathlib import Path

import pytest


@pytest.fixture
def logger():
    """Create a test logger."""
    return logging.getLogger('test')

@pytest.fixture
def test_data_dir():
    """Get the test data directory."""
    return Path('tests/test_data')

@pytest.fixture
def test_config(test_data_dir):
    """Create a test configuration."""
    from minecraft_mod_manager.config.config import Config

    # Create test directories
    (test_data_dir / 'server').mkdir(parents=True, exist_ok=True)
    (test_data_dir / 'mods').mkdir(parents=True, exist_ok=True)
    (test_data_dir / 'backups').mkdir(parents=True, exist_ok=True)
    (test_data_dir / 'logs').mkdir(parents=True, exist_ok=True)

    return Config.from_dict({
        'minecraft': {'version': '1.20.1', 'modloader': 'fabric'},
        'paths': {
            'server': str(test_data_dir / 'server'),
            'mods': str(test_data_dir / 'mods'),
            'backups': str(test_data_dir / 'backups'),
            'logs': str(test_data_dir / 'logs/mod-manager.log')
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
    })
