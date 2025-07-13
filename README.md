# Minecraft Mod Manager: Effortless Server Modding

[![Build Status](https://img.shields.io/github/actions/workflow/status/dacrab/minecraft-mod-manager/test.yml?branch=main)](https://github.com/dacrab/minecraft-mod-manager/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/dacrab/minecraft-mod-manager)](https://github.com/dacrab/minecraft-mod-manager/blob/main/LICENSE)

**A comprehensive command-line tool for managing Minecraft server mods, designed to simplify the process of maintaining and updating mods while ensuring server stability.**

---

## ✨ Key Features

- **🎮 Server Management**: Start, stop, and restart your server with ease. Check server status and player count.
- **🔄 Automated Mod Updates**: Automatically update mods from Modrinth and CurseForge.
- **🔔 Smart Notifications**: Get Discord notifications for updates and send warnings to players before restarts.
- **💾 Automatic Backups**: Create backups of your mods directory before updating, with configurable retention.
- **⚙️ Highly Configurable**: Customize everything from Java flags to notification messages.

## 🚀 Getting Started

### Requirements
- Python 3.9+
- Java (for running the Minecraft server)

### Installation

1.  **Install from PyPI (Recommended)**

    ```bash
    pip install minecraft-mod-manager
    ```

2.  **Install from source**

    ```bash
    git clone https://github.com/dacrab/minecraft-mod-manager.git
    cd minecraft-mod-manager
    pip install -e .
    ```

## Usage

Once installed, you can use the `minecraft-mod-manager` command.

```bash
# Run the automated update process
minecraft-mod-manager --auto-update

# Start the server
minecraft-mod-manager --start

# Stop the server
minecraft-mod-manager --stop

# Check server status
minecraft-mod-manager --status

# Use a custom config file
minecraft-mod-manager --config /path/to/your/config.toml --auto-update
```

## 🔧 Configuration

Create a `config.toml` file. You can place it in `~/.config/minecraft-mod-manager/config.toml` or use the `--config` flag to specify a path.

Here is an example configuration:

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
java_flags = [
    "-Xms4G",
    "-Xmx4G",
]
stop_command = "stop"

[notifications]
discord_webhook = "YOUR_DISCORD_WEBHOOK_URL"
warning_intervals = [15, 10, 5, 1]

[mods]
auto_update = true
backup_before_update = true

# Add your mods here
[mods.sources]
modrinth = [
    "https://modrinth.com/mod/fabric-api",
    "https://modrinth.com/mod/lithium",
    "https://modrinth.com/mod/starlight"
]
curseforge = [
    "https://www.curseforge.com/minecraft/mc-mods/jei",
    "https://www.curseforge.com/minecraft/mc-mods/jade"
]
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. If you find any issues or have a feature request, please open an issue on the [GitHub repository](https://github.com/dacrab/minecraft-mod-manager/issues).

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.