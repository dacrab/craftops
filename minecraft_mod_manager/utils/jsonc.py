"""JSONC (JSON with Comments) file handling utilities."""

import json
import re
from typing import Dict, Any

def load_jsonc(file_path: str) -> Dict[str, Any]:
    """
    Load a JSONC file and return a dictionary.
    
    Strips out comments (both single-line and multi-line) before parsing the configuration.
    Single-line comments start with '//' and multi-line comments are wrapped in '/* */'.
    
    Args:
        file_path: Path to the JSONC file to load
        
    Returns:
        Dictionary containing the parsed JSON data
        
    Raises:
        FileNotFoundError: If the file doesn't exist
        json.JSONDecodeError: If the JSON is invalid
        Exception: For other errors during loading
    """
    try:
        with open(file_path, 'r') as f:
            content = f.read()
        
        # Remove single-line comments
        content = re.sub(r'//.*$', '', content, flags=re.MULTILINE)
        # Remove multi-line comments
        content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)
        
        return json.loads(content)
        
    except FileNotFoundError:
        raise FileNotFoundError(f"Configuration file not found: {file_path}")
    except json.JSONDecodeError as e:
        raise json.JSONDecodeError(
            f"Invalid JSON in configuration file: {str(e)}",
            e.doc,
            e.pos
        )
    except Exception as e:
        raise Exception(f"Error loading configuration file: {str(e)}") 