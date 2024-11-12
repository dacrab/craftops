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
- ⚙️ Configurable Java settings (custom & flags.sh)
- 📊 Server status monitoring
- ⏰ Configurable maintenance warnings
- 🔄 Smart mod version compatibility checking
- 📝 Detailed logging system

## 📋 Requirements

- Python 3.7+
- Linux environment
- Java (for Minecraft server)

## 🚀 Installation

### 1. Install Python & Dependencies

**Method A: Using Package Manager (Recommended for Beginners)**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install python3 python3-pip python3-aiohttp python3-requests python3-tqdm

# CentOS/RHEL
sudo dnf install python3 python3-pip python3-aiohttp python3-requests python3-tqdm

# Arch Linux
sudo pacman -S python python-pip python-aiohttp python-requests python-tqdm
```

**Method B: Using pip (Recommended for Advanced Users)**
```bash
# First install Python & pip
sudo apt update && sudo apt install python3 python3-pip  # Ubuntu/Debian
sudo dnf install python3 python3-pip                     # CentOS/RHEL
sudo pacman -S python python-pip                         # Arch Linux

# Then install dependencies
pip3 install aiohttp requests tqdm
```

### 2. Setup
```bash
git clone https://github.com/dacrab/minecraft-mod-manager.git
cd minecraft-mod-manager
cp config.json.example config.json
```

## 🔧 Configuration

### Basic Setup
Edit `config.json` with your settings. Here are the key configuration sections:

### 1. Paths Configuration
```json
{
    "paths": {
        "minecraft": "/home/Minecraft",
        "local_mods": "/home/Minecraft/mods",
        "backups": "/home/Minecraft/backups",
        "server_jar": "/home/Minecraft/server.jar",
        "logs": "/home/Minecraft/logs/mod_manager.log"
    }
}
```

### 2. Minecraft Settings
```json
{
    "minecraft": {
        "version": "1.21.1",
        "modloader": "fabric"  // fabric/forge/quilt
    }
}
```

### 3. Server Memory & Java Flags
Choose between flags.sh or custom configuration:

```json
{
    "server": {
        "memory": {
            "source": "flags_sh",  // or "custom"
            "min": "4G",           // used if source is "custom"
            "max": "6G"
        },
        "java_flags": {
            "source": "custom",    // or "flags_sh"
            "flags_sh": "",        // paste flags.sh output here
            "custom": [
                "-XX:+UseG1GC",
                "-XX:+ParallelRefProcEnabled",
                // ... other flags
            ]
        }
    }
}
```

### 4. Maintenance Configuration
```json
{
    "maintenance": {
        "backup_retention_days": 7,
        "backup_name_format": "minecraft-%Y.%m.%d-%H.%M",
        "warning_intervals": [
            {"time": 15, "unit": "minutes"},
            {"time": 5, "unit": "minutes"},
            {"time": 1, "unit": "minute"}
        ]
    }
}
```

### 5. Discord Notifications
```json
{
    "notifications": {
        "discord_webhook": "YOUR_DISCORD_WEBHOOK_URL_HERE"
    }
}
```

### 6. API Settings
```json
{
    "api": {
        "max_retries": 5,
        "base_delay": 3,
        "chunk_size": 10,
        "startup_timeout": 120
    }
}
```

### 7. Mod List
```json
{
    "modrinth_urls": [
        "https://modrinth.com/mod/MOD_ID_1",
        "https://modrinth.com/mod/MOD_ID_2"
    ]
}
```

## 🎮 Usage

### Manual Run
```bash
python3 MinecraftModManager.py
```

### Automated (Cron)
```bash
# Daily at 4 AM
0 4 * * * /usr/bin/python3 /path/to/MinecraftModManager.py --auto-update
```

## 🔍 Troubleshooting

Check logs:
```bash
tail -f /home/Minecraft/logs/mod_manager.log
```

Common fixes:
- **Server won't start**: Check Java version & memory settings
- **Permission errors**: Run `sudo chown -R $USER:$USER /home/Minecraft`
- **Java flags issues**: Verify flags.sh output format
- **Package conflicts**: Try alternative installation method
- **Dependencies missing**: Use package manager installation
- **Mod compatibility issues**: Check Minecraft version and modloader settings
- **API rate limits**: Adjust api.max_retries and api.base_delay in config

## 📝 License

MIT License

---
<div align="center">
Made with ❤️ for Minecraft servers
</div>
