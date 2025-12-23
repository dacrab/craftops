# CraftOps

Modern CLI for Minecraft server operations and Modrinth mod management.

## Features

- **Lifecycle:** Start, stop, and restart via GNU screen.
- **Mods:** Automated updates and dependency resolution from Modrinth.
- **Backups:** Compressed archives with smart retention policies.
- **Alerts:** Discord webhook integration for status and errors.
- **Health:** Integrated diagnostic suite for paths and dependencies.

## Install

### One-liner
```bash
curl -sSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
```

### From source
```bash
make install
```

### Docker
```bash
docker pull ghcr.io/dacrab/craftops:latest
```

## Usage

```bash
craftops init-config                    # Initialize default config
craftops health-check                   # Run system diagnostics
craftops update-mods                    # Check and install mod updates
craftops server start|stop|restart      # Control server process
craftops backup create|list             # Manage server archives
```

**Global Flags:** `--config`, `--debug`, `--dry-run`

## Configuration

Generate a default config with `craftops init-config`, then edit `~/.config/craftops/config.toml`:

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
discord_webhook = ""  # Optional Discord integration
```

## Requirements

- Linux or macOS
- GNU screen
- Java 17+ (installed on the host for non-Docker use)

## License

MIT
