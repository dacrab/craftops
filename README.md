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
git clone https://github.com/yourusername/minecraft-mod-manager.git
cd minecraft-mod-manager

# Install dependencies
pip3 install aiohttp requests tqdm

# Create config file
cp config.jsonc.example config.jsonc
```

## 🔧 Configuration

Edit `config.jsonc` with your settings:

### Server Configuration
```jsonc
{
    "server": {
        "minecraft_version": "1.21.1",
        "modloader": "fabric",
        "java": {
            "min_memory": "4G",
            "max_memory": "6G",
            "flags": [
                "-XX:+UseG1GC",
                "-XX:+ParallelRefProcEnabled",
                "-XX:MaxGCPauseMillis=200",
                "-XX:+UnlockExperimentalVMOptions",
                "-XX:+DisableExplicitGC",
                "-XX:+AlwaysPreTouch"
            ]
        }
    }
}
```

### Paths Configuration
```jsonc
{
    "paths": {
        "local_mods": "/home/Minecraft/mods",
        "backups": "/home/Minecraft/backups",
        "minecraft": "/home/Minecraft",
        "server_jar": "/home/Minecraft/server.jar"
    }
}
```

### Maintenance Settings
```jsonc
{
    "maintenance": {
        "backup_retention_days": 7,
        "warning_intervals": [
            {"time": 15, "unit": "minutes"},
            {"time": 10, "unit": "minutes"},
            {"time": 5, "unit": "minutes"},
            {"time": 1, "unit": "minute"},
            {"time": 30, "unit": "seconds"},
            {"time": 10, "unit": "seconds"},
            {"time": 5, "unit": "seconds"}
        ]
    }
}
```

### Discord Notifications
```jsonc
{
    "notifications": {
        "discord_webhook": "YOUR_WEBHOOK_URL_HERE",
        "enabled": true
    }
}
```

### Mod List
```jsonc
{
    "modrinth_urls": [
        "https://modrinth.com/mod/example1",
        "https://modrinth.com/mod/example2"
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
