"""TOML file handling utilities."""

from typing import Any, Dict, List, TypeVar, Union, cast

try:
    import tomllib  # type: ignore[import-not-found]  # Python 3.11+
except ImportError:
    import tomli as tomllib  # Python <3.11

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
        # Convert the data to TOML format
        toml_str = _dict_to_toml(data)

        # Write to file
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(toml_str)
    except Exception as err:
        raise RuntimeError(f"Error saving configuration file: {str(err)}") from err


def _format_value(value: TomlValue) -> str:
    """Format a value for TOML output."""
    if isinstance(value, str):
        return f'"{value}"'
    elif isinstance(value, bool):
        return str(value).lower()
    elif isinstance(value, (list, tuple)):
        if not value:
            return "[]"
        if isinstance(value[0], dict):
            return ""  # Array of tables are handled separately
        items = [_format_value(cast(TomlValue, item)) for item in value]
        return f"[{', '.join(items)}]"
    else:
        return str(value)


def _dict_to_toml(data: Dict[str, Any], indent: int = 0) -> str:
    """Convert a dictionary to TOML format."""
    lines = []
    arrays_of_tables: Dict[str, List[Dict[str, Any]]] = {}

    # Sort keys to ensure consistent output
    for key, value in sorted(data.items()):
        if isinstance(value, dict):
            # Handle nested dictionaries (tables)
            lines.append(f"\n[{key}]")
            lines.append(_dict_to_toml(value, indent + 2))
        elif isinstance(value, (list, tuple)) and value and isinstance(value[0], dict):
            # Collect arrays of tables for later
            arrays_of_tables[key] = cast(List[Dict[str, Any]], value)
        else:
            # Handle basic types and simple arrays
            formatted_value = _format_value(cast(TomlValue, value))
            if formatted_value:  # Skip empty array of tables
                lines.append(f"{' ' * indent}{key} = {formatted_value}")

    # Add arrays of tables at the end
    for key, array in arrays_of_tables.items():
        for item in array:
            lines.append(f"\n[[{key}]]")
            lines.append(_dict_to_toml(item, indent + 2))

    return "\n".join(lines)
