# 🎮 Minecraft Mod Manager

<div align="center">

![Python Version](https://img.shields.io/badge/python-3.7%2B-blue)
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

## 📋 Requirements

- Python 3.7+
- Linux environment
- Java (for Minecraft server)
- Required Python packages: aiohttp, requests, tqdm

## 🚀 Installation

```bash
# Clone repository
git clone https://github.com/dacrab/minecraft-mod-manager.git
cd minecraft-mod-manager

# Install dependencies
# Via pip
pip3 install aiohttp requests tqdm

# Via package manager
# Debian/Ubuntu
sudo apt-get install python3 python3-aiohttp python3-requests python3-tqdm

# Fedora
sudo dnf install python3 python3-aiohttp python3-requests python3-tqdm

# Arch Linux
sudo pacman -S python python-aiohttp python-requests python-tqdm

# Create config file
cp config.json.example config.json
```

## 🔧 Configuration

Edit `config.json` with your settings:

```json
{
    // Basic Minecraft server configuration
    "minecraft": {
        "version": "1.21.1",        // Minecraft version to use
        "modloader": "fabric"       // Mod loader type (fabric/forge)
    },

    // File system paths configuration
    "paths": {
        "minecraft": "/home/Minecraft",                      // Root Minecraft directory
        "server_jar": "/home/Minecraft/server.jar",         // Server executable path
        "local_mods": "/home/Minecraft/mods",               // Mods directory
        "backups": "/home/Minecraft/backups",               // Backup storage location
        "logs": "/home/Minecraft/logs/mod_manager.log"      // Log file location
    },

    // Server runtime configuration
    "server": {
        "flags_source": "custom",   // Use custom JVM flags
        "custom_flags": "java -Xms8192M -Xmx8192M --add-modules=jdk.incubator.vector -XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch -XX:G1HeapWastePercent=5 -XX:G1MixedGCCountTarget=4 -XX:InitiatingHeapOccupancyPercent=15 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:SurvivorRatio=32 -XX:+PerfDisableSharedMem -XX:MaxTenuringThreshold=1 -Dusing.aikars.flags=https://mcflags.emc.gs -Daikars.new.flags=true -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40 -XX:G1HeapRegionSize=8M -XX:G1ReservePercent=20",
        
        // Memory allocation settings
        "memory": {
            "min": "6G",           // Minimum heap size
            "max": "8G"            // Maximum heap size
        },

        // JVM optimization flags
        "java_flags": [
            "-XX:+UseG1GC",                    // Use G1 Garbage Collector
            "-XX:+ParallelRefProcEnabled",     // Enable parallel reference processing
            "-XX:MaxGCPauseMillis=200",        // Target max GC pause time
            "-XX:+UnlockExperimentalVMOptions",// Allow experimental options
            "-XX:+DisableExplicitGC",          // Disable explicit GC calls
            "-XX:G1NewSizePercent=30",         // New generation size
            "-XX:G1MaxNewSizePercent=40",      // Max new generation size
            "-XX:G1HeapRegionSize=8M",         // G1 region size
            "-XX:G1ReservePercent=20",         // Reserve memory percentage
            "-XX:G1HeapWastePercent=5"         // Acceptable waste percentage
        ],

        // Server startup settings
        "startup": {
            "max_retries": 3,      // Maximum startup attempts
            "retry_delay": 10      // Delay between retries (seconds)
        },

        // RCON configuration for remote control
        "rcon": {
            "enabled": true,                       // Enable RCON
            "port": 25575,                        // RCON port
            "password": "YOUR_RCON_PASSWORD_HERE"  // RCON password
        }
    },

    // API interaction settings
    "api": {
        "user_agent": "MinecraftModManager/1.0",  // User agent for API requests
        "max_retries": 5,                         // Maximum API retry attempts
        "base_delay": 3,                          // Base delay between retries
        "chunk_size": 10                          // Download chunk size
    },

    // Logging configuration
    "logging": {
        "max_lines": {
            "server_check": 100,    // Max lines for server checks
            "startup_check": 5,     // Max lines for startup logs
            "status_check": 50      // Max lines for status checks
        }
    },

    // Maintenance and backup settings
    "maintenance": {
        "backup_retention_days": 7,    // Days to keep backups
        "max_backups": 5,             // Maximum number of backups
        "backup_name_format": "minecraft-%Y.%m.%d-%H.%M",  // Backup filename format
        // Warning intervals before maintenance
        "warning_intervals": [
            {"time": 15, "unit": "minutes"},
            {"time": 10, "unit": "minutes"},
            {"time": 5, "unit": "minutes"},
            {"time": 1, "unit": "minute"},
            {"time": 30, "unit": "seconds"},
            {"time": 10, "unit": "seconds"},
            {"time": 5, "unit": "seconds"}
        ]
    },

    // Discord notification settings
    "notifications": {
        "discord_webhook": "YOUR_WEBHOOK_URL_HERE"  // Discord webhook URL
    },

    // Server validation settings
    "validation": {
        "required_files": [          // Required server files
            "server.jar",
            "eula.txt"
        ],
        "eula_accepted": true        // EULA acceptance flag
    },

    // Mod configuration
    "modrinth_urls": [              // Modrinth mod URLs to manage
        "https://modrinth.com/mod/1IjD5062",
        "https://modrinth.com/mod/1bokaNcj",
        "https://modrinth.com/mod/ZJTGwAND"
    ]
}
```

## 🎮 Usage

### Command Line Interface

```bash
# Check server status
python3 MinecraftModManager.py --status

# Start server
python3 MinecraftModManager.py --start

# Stop server
python3 MinecraftModManager.py --stop

# Restart server
python3 MinecraftModManager.py --restart

# Run automated update
python3 MinecraftModManager.py --auto-update

# Run manual maintenance
python3 MinecraftModManager.py
```

### Automated Updates (Cron)
```bash
# Daily at 4 AM
0 4 * * * /usr/bin/python3 /path/to/MinecraftModManager.py --auto-update
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

- **Backup issues**: 
  - Verify backup directory permissions
  - Check available disk space
  - Ensure proper path configuration

- **Discord notifications not working**: 
  - Verify webhook URL
  - Check notifications.enabled setting
  - Test webhook URL manually

## 📝 License

MIT License

---
<div align="center">
Made with ❤️ for Minecraft servers
</div>
