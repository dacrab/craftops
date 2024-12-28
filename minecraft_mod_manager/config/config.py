"""Configuration management for Minecraft Mod Manager."""

import json
import logging
from pathlib import Path
from typing import Any, Dict, List, TypeVar, Union, overload

T = TypeVar('T')

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
    
    @overload
    def __getitem__(self, key: str) -> Dict[str, Any]: ...
    
    @overload
    def __getitem__(self, key: str) -> Any: ...
    
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