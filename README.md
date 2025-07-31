# Minecraft Mod Manager

[![Tests](https://img.shields.io/github/actions/workflow/status/dacrab/minecraft-mod-manager/test.yml?branch=main&label=tests)](https://github.com/dacrab/minecraft-mod-manager/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/dacrab/minecraft-mod-manager/release.yml?label=release)](https://github.com/dacrab/minecraft-mod-manager/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/dacrab/minecraft-mod-manager)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.9+-blue.svg)](https://python.org)

A comprehensive command-line tool for managing Minecraft server mods with automated updates, backups, and notifications.

## Core Features

- **Server Management**: Start, stop, restart servers with status monitoring
- **Automated Mod Updates**: Updates from Modrinth and CurseForge APIs
- **Smart Notifications**: Discord webhooks and in-game player warnings
- **Backup System**: Automatic backups before updates with retention policies
- **Configuration-Driven**: TOML-based configuration with sensible defaults

## Installation

### Option 1: PyPI Package
```bash
pip install minecraft-mod-manager
```

### Option 2: Pre-built Executable
Download from [GitHub Releases](https://github.com/dacrab/minecraft-mod-manager/releases):
- Linux: `minecraft-mod-manager-linux`
- Windows: `minecraft-mod-manager-windows.exe`
- macOS: `minecraft-mod-manager-macos`

### Option 3: From Source
```bash
git clone https://github.com/dacrab/minecraft-mod-manager.git
cd minecraft-mod-manager
make setup-dev
```

## Quick Start

1. **Install the tool**
   ```bash
   pip install minecraft-mod-manager
   ```

2. **Create configuration**
   ```bash
   mkdir -p ~/.config/minecraft-mod-manager
   minecraft-mod-manager --init-config
   ```

3. **Edit configuration**
   ```bash
   nano ~/.config/minecraft-mod-manager/config.toml
   ```

4. **Run health check**
   ```bash
   minecraft-mod-manager --health-check
   ```

5. **Start managing mods**
   ```bash
   minecraft-mod-manager --auto-update
   ```

## Configuration

The tool uses TOML configuration files. Default location: `~/.config/minecraft-mod-manager/config.toml`

```toml
[minecraft]
version = "1.20.1"
modloader = "fabric"

[paths]
server = "/path/to/minecraft/server"
mods = "/path/to/minecraft/server/mods"
backups = "/path/to/minecraft/backups"
logs = "/path/to/minecraft/logs/mod-manager.log"

[server]
jar = "server.jar"
java_flags = ["-Xms4G", "-Xmx4G"]
stop_command = "stop"

[notifications]
discord_webhook = "YOUR_DISCORD_WEBHOOK_URL"
warning_intervals = [15, 10, 5, 1]

[mods]
auto_update = true
backup_before_update = true
chunk_size = 5
base_delay = 2

[mods.sources]
modrinth = [
    "https://modrinth.com/mod/fabric-api",
    "https://modrinth.com/mod/lithium"
]
curseforge = [
    "https://www.curseforge.com/minecraft/mc-mods/jei"
]

[backup]
max_backups = 5
compression = true
```

## Usage

```bash
# Server management
minecraft-mod-manager --start          # Start server
minecraft-mod-manager --stop           # Stop server
minecraft-mod-manager --restart        # Restart server
minecraft-mod-manager --status         # Check status

# Mod management
minecraft-mod-manager --auto-update    # Update all mods
minecraft-mod-manager --check-updates  # Check for updates only
minecraft-mod-manager --backup         # Create backup

# System operations
minecraft-mod-manager --health-check   # Run diagnostics
minecraft-mod-manager --cleanup        # Clean old files
minecraft-mod-manager --init-config    # Create config template

# Custom configuration
minecraft-mod-manager --config /path/to/config.toml --auto-update
```

## Development

### Requirements
- Python 3.9+
- Virtual environment recommended

### Setup
```bash
# Clone repository
git clone https://github.com/dacrab/minecraft-mod-manager.git
cd minecraft-mod-manager

# Setup development environment
make setup-dev

# Run tests
make test

# Check code quality
make lint type-check

# Build package
make build

# Build executable  
make build-exe
```

### Available Commands
```bash
make help           # Show all commands
make test           # Run tests with coverage
make lint           # Code linting with ruff
make type-check     # Type checking with mypy
make format         # Format code
make build          # Build Python package
make build-exe      # Build standalone executable
make clean          # Clean build artifacts
make health-check   # Run application health check
```

## Architecture

The project follows a modular architecture with clear separation of concerns:

```
minecraft_mod_manager/
├── app.py                    # Main application entry point
├── services.py               # Core business logic
└── settings/                 # Configuration management
    ├── config.py            # Configuration dataclasses
    └── config.toml          # Default configuration
```

### Key Components
- **ModManager**: Handles mod downloading and updating
- **BackupManager**: Creates and manages backups
- **NotificationManager**: Discord webhooks and player notifications
- **ServerController**: Minecraft server process control

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make changes and add tests
4. Run quality checks: `make lint type-check test`
5. Commit changes: `git commit -am 'Add feature'`
6. Push to branch: `git push origin feature-name`
7. Submit a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.