# 🎮 Minecraft Mod Manager

<div align="center">

![Python Version](https://img.shields.io/badge/python-3.9%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Linux-lightgrey)

A Python-based tool for managing Minecraft server mods and maintenance.

</div>

## ✨ Features

- 🔄 Automated mod updates from Modrinth
- 💾 Smart backup system with retention policies
- 🔔 Discord webhook notifications for server events
- ⚙️ Configurable Java memory and optimization flags
- 📊 Server status monitoring and player count tracking
- ⏰ Customizable maintenance warnings with countdown
- 🔄 Intelligent mod version compatibility checking
- 📝 Comprehensive logging system
- 🛠️ Command-line interface for server management

## 📋 Quick Start

There are two ways to install the Minecraft Mod Manager:

### Method 1: Python Package (for Python users)

1. **Install the Package**:
   ```bash
   pip install minecraft-mod-manager
   ```

2. **Create Configuration**:
   ```bash
   # Create config directory
   mkdir -p ~/.config/minecraft-mod-manager
   
   # Copy example config
   cp $(pip show minecraft-mod-manager | grep Location | cut -d' ' -f2)/minecraft_mod_manager/config.jsonc.example ~/.config/minecraft-mod-manager/config.jsonc
   ```

### Method 2: Standalone Executable (for non-Python users)

1. **Download the Executable**:
   - Go to [Releases](https://github.com/dacrab/minecraft-mod-manager/releases/latest)
   - Download the `minecraft-mod-manager` file

2. **Make it Executable**:
   ```bash
   chmod +x minecraft-mod-manager
   ```

3. **Create Configuration**:
   ```bash
   # Create config directory
   mkdir -p ~/.config/minecraft-mod-manager
   
   # Initialize config (this will create an example config)
   ./minecraft-mod-manager --init
   ```

### Common Steps (for both methods)

1. **Edit Configuration**:
   ```bash
   # Edit the config file with your settings
   nano ~/.config/minecraft-mod-manager/config.jsonc
   ```
   Key things to update:
   - Set correct Minecraft paths
   - Configure server memory and Java flags
   - Add your Discord webhook URL
   - Add your Modrinth mod URLs

2. **Test the Setup**:
   ```bash
   # Check server status
   minecraft-mod-manager --status  # or ./minecraft-mod-manager --status for executable
   
   # Run a manual update
   minecraft-mod-manager --auto-update  # or ./minecraft-mod-manager --auto-update for executable
   ```

3. **Set up Auto-updates** (optional):
   ```bash
   # Add to crontab (runs daily at 4 AM)
   (crontab -l 2>/dev/null; echo "0 4 * * * $(pwd)/minecraft-mod-manager --auto-update") | crontab -
   ```

## 📋 Requirements

- Python 3.9 or newer
- Linux environment
- Java (for Minecraft server)
- Internet connection for mod updates

## 🔧 Configuration

The configuration file (`config.jsonc`) supports comments and includes these sections:

```jsonc
{
    // Minecraft server configuration
    "minecraft": {
        "version": "1.20.1",        // Minecraft version
        "modloader": "fabric"       // Mod loader (fabric/forge)
    },

    // File paths
    "paths": {
        "minecraft": "/path/to/minecraft",
        "server_jar": "/path/to/minecraft/server.jar",
        "local_mods": "/path/to/minecraft/mods",
        "backups": "/path/to/minecraft/backups",
        "logs": "/path/to/minecraft/logs/mod_manager.log"
    },

    // Server process settings
    "server": {
        "flags_source": "default",  // Use "default" or "custom"
        "custom_flags": "java -Xms8G -Xmx8G -XX:+UseG1GC -jar",  // Used if flags_source is "custom"
        "memory": {
            "min": "6G",           // Used if flags_source is "default"
            "max": "8G"
        },
        "startup": {
            "max_retries": 3,
            "retry_delay": 10
        },
        "rcon": {
            "enabled": true,
            "port": 25575,
            "password": "YOUR_RCON_PASSWORD"
        }
    },

    // API settings
    "api": {
        "user_agent": "MinecraftModManager/1.0",
        "max_retries": 5,
        "base_delay": 3,
        "chunk_size": 10
    },

    // Maintenance settings
    "maintenance": {
        "warning_intervals": [
            {"time": 30, "unit": "minutes"},
            {"time": 15, "unit": "minutes"},
            {"time": 5, "unit": "minutes"},
            {"time": 1, "unit": "minutes"}
        ],
        "backup_name_format": "%Y-%m-%d_%H-%M-%S",
        "backup_retention_days": 7,
        "max_backups": 10
    },

    // Discord webhook for notifications
    "notifications": {
        "discord_webhook": "YOUR_WEBHOOK_URL"
    },

    // List of mods to manage
    "modrinth_urls": [
        // Essential mods
        "https://modrinth.com/mod/fabric-api",      // Required for most mods
        "https://modrinth.com/mod/lithium",         // Performance improvements
        "https://modrinth.com/mod/starlight",       // Light engine optimization
        
        // Performance mods
        "https://modrinth.com/mod/ferrite-core",    // Memory optimization
        "https://modrinth.com/mod/lazydfu",         // Startup optimization
        "https://modrinth.com/mod/krypton",         // Network optimization
        
        // Utility mods
        "https://modrinth.com/mod/carpet",          // Server tools
        "https://modrinth.com/mod/spark"            // Performance profiler
    ]
}
```

## 🎮 Usage

### Command Line Interface

```bash
# Check server status
minecraft-mod-manager --status

# Start server
minecraft-mod-manager --start

# Stop server
minecraft-mod-manager --stop

# Restart server
minecraft-mod-manager --restart

# Run automated update
minecraft-mod-manager --auto-update

# Run manual maintenance (interactive)
minecraft-mod-manager

# Use custom config file
minecraft-mod-manager --config path/to/config.jsonc
```

### Automated Updates (Cron)
```bash
# Daily at 4 AM
0 4 * * * minecraft-mod-manager --auto-update
```

## 📦 Package Structure

```
minecraft_mod_manager/
├── config/
│   └── config.py           # Configuration management
├── controllers/
│   └── server.py           # Server process control
├── managers/
│   ├── backup.py           # Backup management
│   ├── mod.py              # Mod updates
│   └── notification.py     # Notifications
├── utils/
��   ├── constants.py        # Shared constants
│   └── jsonc.py           # JSONC file handling
├── __init__.py            # Package initialization
├── __main__.py            # Command-line interface
└── minecraft_mod_manager.py # Main orchestration
```

## 🔍 Troubleshooting

Common issues and solutions:

- **Server won't start**: 
  - Check Java version compatibility
  - Verify memory settings in config
  - Check server.jar path
  - Review logs for startup errors

- **Mod update failures**: 
  - Verify Modrinth URLs
  - Check version compatibility
  - Ensure proper permissions on mods directory
  - Check network connectivity

- **Backup issues**: 
  - Verify backup directory permissions
  - Check available disk space
  - Ensure proper path configuration

- **Discord notifications not working**: 
  - Verify webhook URL
  - Test webhook URL manually
  - Check network connectivity

## 📝 Distribution

### Building from Source

1. **Clone and Build**:
   ```bash
   git clone https://github.com/dacrab/minecraft-mod-manager.git
   cd minecraft-mod-manager
   ./build.sh
   ```

2. **Install the Built Package**:
   ```bash
   pip install dist/*.whl
   ```

### Direct Installation

You can also install directly from GitHub:

```bash
pip install https://github.com/dacrab/minecraft-mod-manager/releases/latest/download/minecraft-mod-manager.whl
```

## 📝 License

MIT License

---
<div align="center">
Made with ❤️ for Minecraft servers
</div>
