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
minecraft-mod-manager --config custom_config.jsonc --auto-update
```

## Configuration

Create a configuration file (`config.jsonc`) with the following structure:

```jsonc
{
    "minecraft": {
        "version": "1.21.1",
        "modloader": "fabric"
    },

    "paths": {
        "minecraft": "/path/to/minecraft",
        "server_jar": "/path/to/minecraft/server.jar",
        "local_mods": "/path/to/minecraft/mods",
        "backups": "/path/to/minecraft/backups",
        "logs": "/path/to/minecraft/logs/mod_manager.log"
    },

    "server": {
        "flags_source": "custom",
        "custom_flags": "java -Xms8G -Xmx8G [your-java-flags]",
        "memory": {
            "min": "6G",
            "max": "8G"
        },
        "rcon": {
            "enabled": true,
            "port": 25575,
            "password": "YOUR_RCON_PASSWORD"
        }
    },

    "maintenance": {
        "backup_retention_days": 7,
        "max_backups": 5,
        "warning_intervals": [
            {"time": 15, "unit": "minutes"},
            {"time": 5, "unit": "minutes"},
            {"time": 1, "unit": "minute"}
        ]
    },

    "notifications": {
        "discord_webhook": "YOUR_WEBHOOK_URL"
    },

    "modrinth_urls": [
        "https://modrinth.com/mod/fabric-api",
        "https://modrinth.com/mod/lithium",
        "https://modrinth.com/mod/starlight",
        // Add your desired mods here
    ]
}
```

### Key Configuration Options

- **Minecraft Settings**: Specify your Minecraft version and mod loader
- **Paths**: Configure all necessary file paths for your server
- **Server Settings**: 
  - Customize Java flags for optimal performance
  - Configure memory allocation
  - Set up RCON for remote server control
- **Maintenance**:
  - Set backup retention policies
  - Configure warning intervals for server restarts
- **Notifications**: Set up Discord webhooks for alerts
- **Mod Management**: List all your Modrinth mod URLs for automatic updates

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.