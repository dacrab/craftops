# Technology Stack

## Language & Runtime
- **Python 3.9+** - Core language with modern type hints
- **Async/Await** - Asynchronous programming for HTTP requests and I/O operations

## Key Dependencies
- **aiohttp** - Async HTTP client for API requests to Modrinth/CurseForge
- **tqdm** - Progress bars for user feedback during operations
- **toml** - Configuration file parsing

## Development Tools
- **pytest** - Testing framework with async support
- **mypy** - Static type checking with strict settings
- **ruff** - Fast Python linter and formatter (100 char line limit)
- **pytest-cov** - Code coverage reporting

## Build System
- **setuptools** - Package building via pyproject.toml
- **PyInstaller** - Single-file executable generation
- **pip** - Package management and installation

## Common Commands

### Development Setup
```bash
# Create and activate virtual environment
python -m venv venv
source venv/bin/activate

# Install in development mode
pip install -e .

# Install development dependencies
pip install -r requirements-dev.txt
```

### Testing
```bash
# Run tests with coverage
pytest --cov=minecraft_mod_manager

# Type checking
mypy minecraft_mod_manager/

# Linting
ruff check minecraft_mod_manager/
```

### Building
```bash
# Standard package build
./package-build.sh

# Single-file executable
./executable-build.sh

# PyPI package
python -m build
```

## Configuration
- **TOML format** for all configuration files
- **Dataclasses** for type-safe configuration handling
- **Default config** bundled in package at `minecraft_mod_manager/config/config.toml`