# Project Structure

## Package Organization
The project follows a modular architecture with clear separation of concerns:

```
minecraft_mod_manager/
├── __init__.py                 # Package initialization with version info
├── minecraft_mod_manager.py    # Main application entry point and orchestration
├── config/                     # Configuration handling
│   ├── config.py              # Configuration dataclasses and loading logic
│   └── config.toml            # Default configuration template
├── managers/                   # Core business logic modules
│   ├── __init__.py
│   ├── mod.py                 # Mod downloading and updating logic
│   ├── notification.py        # Discord webhooks and player notifications
│   └── backup.py              # Backup creation and management
├── controllers/                # External system interfaces
│   └── server.py              # Minecraft server process control
└── utils/                      # Shared utilities
    ├── constants.py           # Global constants and enums
    └── toml_utils.py          # TOML file handling utilities
```

## Architecture Patterns

### Dependency Injection
- Main `MinecraftModManager` class orchestrates all components
- Each manager receives config and logger via constructor injection
- Protocol interfaces define contracts (e.g., `ServerControllerProtocol`)

### Configuration-Driven Design
- All behavior controlled via TOML configuration files
- Dataclasses provide type-safe configuration access
- Default config bundled with package, user config overrides

### Async-First Design
- All I/O operations use async/await patterns
- HTTP requests handled via aiohttp for concurrent mod downloads
- Server operations are async to avoid blocking

## File Naming Conventions
- **Snake_case** for all Python files and directories
- **Lowercase** package and module names
- **PascalCase** for class names
- **UPPER_CASE** for constants

## Import Patterns
- Use relative imports within the package (`from .config import load_config`)
- Absolute imports for external dependencies
- Type imports in TYPE_CHECKING blocks when needed for forward references

## Testing Structure
```
tests/
├── __init__.py
├── conftest.py                # Pytest configuration and fixtures
└── test_*.py                  # Test modules mirroring source structure
```

## Build Artifacts
- `dist/` - Built packages (wheel/sdist)
- `build/` - Build intermediates
- `*.egg-info/` - Package metadata
- `venv/` - Virtual environment (not committed)