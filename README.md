# CraftOps

Modern CLI for Minecraft server operations and Modrinth mod management.

## Features

- Server lifecycle management (start/stop/restart via GNU screen)
- Automated mod updates from Modrinth
- Compressed backups with retention policies
- Discord notifications
- Health checks

## Install

```bash
# One-liner
curl -sSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash

# From source
make install

# Docker
docker pull ghcr.io/dacrab/craftops:latest
```

## Usage

```bash
craftops init-config                    # Create default config
craftops health-check                   # Validate setup
craftops update-mods                    # Update mods from Modrinth
craftops server start|stop|restart      # Manage server
craftops backup create|list             # Manage backups
```

**Flags:** `--config`, `--debug`, `--dry-run`

## Config

Generate with `craftops init-config`, then edit `~/.config/craftops/config.toml`:

```toml
[minecraft]
version = "1.20.1"
modloader = "fabric"

[paths]
server = "/home/minecraft/server"
mods = "/home/minecraft/server/mods"
backups = "/home/minecraft/backups"

[mods]
modrinth_sources = [
  "https://modrinth.com/mod/fabric-api",
  "https://modrinth.com/mod/lithium"
]

[backup]
enabled = true
max_backups = 5

[notifications]
discord_webhook = ""  # Optional
```

## Requirements

- Linux/macOS
- GNU screen
- Java 17+ (for Minecraft server)

## License

MIT
