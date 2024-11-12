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
Edit `config.json` with your settings:
```json
{
    "paths": {
        "minecraft": "/home/Minecraft",
        "local_mods": "/home/Minecraft/mods",
        "server_jar": "/home/Minecraft/server.jar"
    },
    "minecraft": {
        "version": "1.21.1",
        "modloader": "fabric"
    }
}
```

### Java Flags Setup
Choose between:

1. **flags.sh** (Recommended):
   - Visit [flags.sh](https://flags.sh/)
   - Copy generated flags
   - Add to config:
   ```json
   {
       "server": {
           "java_flags": {
               "source": "flags_sh",
               "flags_sh": "YOUR_FLAGS_FROM_FLAGS_SH"
           }
       }
   }
   ```

2. **Custom Flags**:
   ```json
   {
       "server": {
           "java_flags": {
               "source": "custom",
               "custom": ["-XX:+UseG1GC", "..."]
           }
       }
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

## 📝 License

MIT License

---
<div align="center">
Made with ❤️ for Minecraft servers
</div>
