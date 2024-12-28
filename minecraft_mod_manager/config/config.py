"""Configuration management for Minecraft Mod Manager."""

import json
import logging
from pathlib import Path
from typing import Any, Dict, List, TypeVar, Union

T = TypeVar('T')

SUPPORTED_MODLOADERS = ['fabric', 'forge', 'quilt']

class Config:
    """Handles configuration loading and validation."""
    
    def __init__(self, config_path: Union[str, Path]):
        """Initialize configuration from file."""
        self.config_path = Path(config_path)
        self.config: Dict[str, Any] = {}
        self.logger = logging.getLogger("Config")
        self._load_config()
        self._validate_config()
    
    def _load_config(self) -> None:
        """Load and parse configuration file."""
        try:
            with open(self.config_path, 'r', encoding='utf-8') as f:
                content = ""
                for line in f:
                    # Skip comments
                    if not line.lstrip().startswith('//'):
                        content += line
                self.config = json.loads(content)
                
        except FileNotFoundError as err:
            raise RuntimeError(f"Configuration file not found: {self.config_path}") from err
        except json.JSONDecodeError as err:
            raise RuntimeError(f"Invalid JSON in configuration: {str(err)}") from err
        except Exception as err:
            raise RuntimeError(f"Failed to load configuration: {str(err)}") from err
    
    def _validate_config(self) -> None:
        """Validate required configuration sections and fields."""
        # Basic structure validation
        required_sections = {
            'paths': ['minecraft', 'backups', 'local_mods', 'logs'],
            'minecraft': ['version', 'modloader'],
            'maintenance': ['warning_intervals', 'backup_name_format'],
            'api': ['chunk_size', 'max_retries', 'base_delay'],
            'notifications': ['discord_webhook']
        }
        
        for section, fields in required_sections.items():
            if section not in self.config:
                raise RuntimeError(f"Missing required section: {section}")
            
            for field in fields:
                if field not in self.config[section]:
                    raise RuntimeError(f"Missing required field: {section}.{field}")
        
        # Validate modloader
        modloader = self.config['minecraft']['modloader'].lower()
        if modloader not in SUPPORTED_MODLOADERS:
            raise RuntimeError(
                f"Unsupported modloader: {modloader}. "
                f"Supported modloaders: {', '.join(SUPPORTED_MODLOADERS)}"
            )
        
        # Validate version format (basic check)
        version = self.config['minecraft']['version']
        if not isinstance(version, str) or not version.count('.') >= 1:
            raise RuntimeError(
                f"Invalid Minecraft version format: {version}. "
                "Expected format: X.Y.Z or X.Y"
            )
        
        # Validate paths exist or can be created
        for path_key in ['minecraft', 'backups', 'local_mods']:
            path = Path(self.config['paths'][path_key])
            try:
                path.mkdir(parents=True, exist_ok=True)
            except Exception as e:
                raise RuntimeError(f"Failed to create directory {path}: {str(e)}")
        
        # Validate server configuration if present
        if 'server' in self.config:
            if 'memory' in self.config['server']:
                memory = self.config['server']['memory']
                if not isinstance(memory, dict) or 'min' not in memory or 'max' not in memory:
                    raise RuntimeError("Server memory configuration must include 'min' and 'max' values")
    
    def __getitem__(self, key: str) -> Any:
        """Access configuration values."""
        return self.config[key]
    
    def get(self, key: str, default: T = None) -> Union[Dict[str, Any], T]:
        """Get configuration value with default."""
        return self.config.get(key, default)
    
    def get_list(self, key: str, default: Union[List[T], None] = None) -> List[T]:
        """Get list configuration value with default."""
        value = self.config.get(key, default if default is not None else [])
        if not isinstance(value, list):
            raise TypeError(f"Configuration value for {key} must be a list")
        return value