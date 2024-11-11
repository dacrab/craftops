# 🎮 Minecraft Mod Manager

<div align="center">

![Python Version](https://img.shields.io/badge/python-3.7%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Linux-lightgrey)

A Python-based tool for managing Minecraft server mods and maintenance.

</div>

## ✨ Features

- 🔄 Automated mod updates from Modrinth
- 💾 Automated backups with retention policy
- 🔔 Discord webhook notifications
- ⚠️ Player warnings before maintenance
- ⚙️ Configurable Java settings
- 📊 Server status monitoring

## 📋 Requirements

- Python 3.7+
- Required packages: `aiohttp`, `requests`, `tqdm`
- Linux environment
- Java (for Minecraft server)

## 🚀 Installation

### 1. Install Python and pip
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install python3 python3-pip

# CentOS/RHEL
sudo dnf install python3 python3-pip

# Arch Linux
sudo pacman -S python python-pip
```

### 2. Install Required Packages

Choose one of these methods:

```bash
# Using pip (Recommended)
pip3 install aiohttp requests tqdm

# OR using package manager:

# Ubuntu/Debian
sudo apt install python3-aiohttp python3-requests python3-tqdm

# CentOS/RHEL
sudo dnf install python3-aiohttp python3-requests python3-tqdm

# Arch Linux
sudo pacman -S python-aiohttp python-requests python-tqdm
```

### 3. Setup

```bash
# Clone and setup
git clone https://github.com/dacrab/minecraft-mod-manager.git
cd minecraft-mod-manager
cp config.json.example config.json
```

## 🔧 Configuration

Edit `config.json` with your settings:

```json
{
    "paths": {
        "local_mods": "/home/Minecraft/mods",
        "backups": "/home/Minecraft/backups",
        "minecraft": "/home/Minecraft",
        "server_jar": "/home/Minecraft/server.jar",
        "logs": "/home/Minecraft/logs/mod_manager.log"
    },
    "minecraft": {
        "version": "1.21.1",
        "modloader": "fabric"
    },
    "server": {
        "memory": {
            "min": "4G",
            "max": "6G"
        },
        "java_flags": [
            "-XX:+UseG1GC",
            "-XX:+ParallelRefProcEnabled",
            "-XX:MaxGCPauseMillis=200"
            // ... additional optimization flags available
        ]
    },
    "maintenance": {
        "backup_retention_days": 7,
        "backup_name_format": "minecraft-%Y.%m.%d-%H.%M",
        "warning_intervals": [
            {"time": 15, "unit": "minutes"},
            {"time": 5, "unit": "minutes"},
            {"time": 1, "unit": "minute"}
        ]
    },
    "notifications": {
        "discord_webhook": "YOUR_DISCORD_WEBHOOK_URL_HERE"
    },
    "api": {
        "user_agent": "MinecraftModManager/1.0",
        "max_retries": 5,
        "base_delay": 3,
        "chunk_size": 10,
        "startup_timeout": 120
    },
    "logging": {
        "max_lines": {
            "status_check": 50,
            "startup_check": 5,
            "server_check": 100
        }
    },
    "modrinth_urls": [
        "https://modrinth.com/mod/example1",
        "https://modrinth.com/mod/example2"
    ]
}
```

## 🎮 Usage

### Manual Update
```bash
python3 MinecraftModManager.py
```

### Automated Update
```bash
python3 MinecraftModManager.py --auto-update
```

### Automated Maintenance (Cron)
```bash
# Add to crontab for daily maintenance at 4 AM
0 4 * * * /usr/bin/python3 /path/to/MinecraftModManager.py --auto-update
```

## 🔍 Troubleshooting

Check logs at:
```bash
tail -f /home/Minecraft/logs/mod_manager.log
tail -f /home/Minecraft/logs/latest.log
```

## 📝 License

This project is licensed under the MIT License.

---
<div align="center">
Made with ❤️ for Minecraft servers
</div>
