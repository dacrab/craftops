# Minecraft Mod Manager

A comprehensive command-line tool for managing Minecraft server mods, designed to simplify the process of maintaining and updating mods while ensuring server stability.

## Features

- 🎮 Server Management
  - Start, stop, and restart server with ease
  - Check server status and player count
  - Automated server maintenance
  - RCON support for server control

- 🔄 Mod Management
  - Automatic mod updates from Modrinth
  - Smart backup system with retention policies
  - Configurable backup rotation
  - Mod compatibility checking

- 🔔 Notifications
  - Discord webhook integration
  - Customizable player warnings
  - Detailed update and maintenance notifications
  - Progressive countdown notifications

- ⚙️ Advanced Configuration
  - Optimized Java flags for better performance
  - Flexible memory management
  - Detailed logging system
  - Comprehensive mod list management

## Installation

### Requirements

- Linux operating system
- Java (for Minecraft server)

### Installation Steps

1. Download the latest release from the [Releases page](https://github.com/dacrab/minecraft-mod-manager/releases/latest)
2. Make the file executable:

```bash
chmod +x minecraft-mod-manager
```

3. Optionally, move it to a directory in your PATH for easier access:

```bash
sudo mv minecraft-mod-manager /usr/local/bin/
```

## Usage

### Basic Commands

Check server status:

```bash
minecraft-mod-manager --status
```

Start the server:

```bash
minecraft-mod-manager --start
```

Stop the server:

```bash
minecraft-mod-manager --stop
```

Restart the server:

```bash
minecraft-mod-manager --restart
```

### Mod Management

Run automated update process with player warnings:

```bash
minecraft-mod-manager --auto-update
```

Use custom configuration file:

```bash
minecraft-mod-manager --config custom_config.toml --auto-update
```

## Configuration

The application requires a configuration file named `config.toml` in the appropriate directory. Here's how to set it up:

1. The default configuration file will be installed in the `minecraft_mod_manager/config` directory
2. Copy or rename this file to `config.toml` in your desired location
3. Edit the file to match your server setup:

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
    "-XX:+UseG1GC",
    "-XX:+ParallelRefProcEnabled",
    "-XX:+UnlockExperimentalVMOptions",
    "-XX:+DisableExplicitGC",
    "-XX:+AlwaysPreTouch"
]
stop_command = "stop"
status_check_interval = 30
max_stop_wait = 300

[backup]
max_backups = 5
name_format = "%Y%m%d_%H%M%S"

[notifications]
discord_webhook = ""
warning_template = "Server will restart in {minutes} minutes for updates"
warning_intervals = [15, 10, 5, 1]

[mods]
auto_update = true
backup_before_update = true
notify_before_update = true
update_check_interval = 24
chunk_size = 5
max_retries = 3
base_delay = 2

# Add your mods here
[[mods.sources]]
type = "modrinth"
url = "https://modrinth.com/mod/fabric-api"

[[mods.sources]]
type = "modrinth"
url = "https://modrinth.com/mod/lithium"

[[mods.sources]]
type = "curseforge"
url = "https://www.curseforge.com/minecraft/mc-mods/jei"
```

### Key Configuration Options

- **Minecraft**: Version and mod loader settings
- **Paths**: Server, mods, backups, and logs locations
- **Server**: Java settings and server control options
- **Backup**: Backup retention and naming
- **Notifications**: Discord alerts and player warnings
- **Mods**: Add mod URLs from Modrinth or CurseForge

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.