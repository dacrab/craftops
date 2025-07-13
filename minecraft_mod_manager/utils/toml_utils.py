"""TOML file handling utilities."""

from typing import Any, Dict

try:
    import tomllib
except ImportError:
    import tomli as tomllib

import toml

T = TypeVar('T')
TomlValue = Union[str, int, float, bool, List[Any], Dict[str, Any]]
TomlDict = Dict[str, TomlValue]


def load_toml(file_path: str) -> Dict[str, Any]:
    """Load a TOML file and return a dictionary.

    Args:
        file_path: Path to the TOML file to load

    Returns:
        Dictionary containing the parsed TOML data

    Raises:
        FileNotFoundError: If the file doesn't exist
        tomllib.TOMLDecodeError: If the TOML is invalid
        Exception: For other errors during loading
    """
    try:
        with open(file_path, 'rb') as f:
            data: Dict[str, Any] = tomllib.load(f)
            return data

    except FileNotFoundError as err:
        raise FileNotFoundError(f"Configuration file not found: {file_path}") from err
    except tomllib.TOMLDecodeError as err:
        raise tomllib.TOMLDecodeError(f"Invalid TOML format: {str(err)}") from err
    except Exception as err:
        raise RuntimeError(f"Error loading configuration file: {str(err)}") from err


def save_toml(file_path: str, data: Dict[str, Any]) -> None:
    """Save a dictionary to a TOML file.

    Args:
        file_path: Path to save the TOML file
        data: Dictionary containing the data to save

    Raises:
        Exception: For errors during saving
    """
    try:
        with open(file_path, 'w', encoding='utf-8') as f:
            toml.dump(data, f)
    except Exception as err:
        raise RuntimeError(f"Error saving configuration file: {str(err)}") from err
