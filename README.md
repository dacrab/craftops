# CraftOps

Modern CLI for Minecraft server operations and Modrinth mod management — built with Go 1.25.7.

## Features

- **Lifecycle** — Start, stop, and restart your server via GNU screen sessions
- **Mods** — Automated updates from Modrinth with concurrent downloads, retries, and dry-run support
- **Backups** — Compressed `.tar.gz` archives with configurable retention and glob-based exclusion patterns
- **Alerts** — Discord webhook notifications for restarts and warnings
- **Health** — Integrated diagnostic suite for paths, dependencies, and API connectivity

## Requirements

- Linux or macOS (amd64 or arm64)
- GNU screen
- Java 17+ (host installs; not required inside Docker)

## Install

### One-liner (recommended)

Downloads the binary for your platform, verifies the SHA256 checksum, and installs to `~/.local/bin` (or `/usr/local/bin` when run as root):

```bash
curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
```

Pin a specific version:

```bash
VERSION=v2.3.0 bash <(curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh)
```

### From source

Requires Go 1.25.7+:

```bash
make install
```

### Docker

```bash
docker pull ghcr.io/dacrab/craftops:latest
```

```bash
docker run --rm \
  -v /path/to/server:/minecraft/server \
  -v /path/to/config:/config \
  ghcr.io/dacrab/craftops:latest health-check
```

## Quick Start

```bash
# 1. Generate a default config
craftops init-config

# 2. Edit the config
$EDITOR ~/.config/craftops/config.toml

# 3. Verify everything looks good
craftops health-check

# 4. Start your server
craftops server start
```

## Usage

```
craftops [command] [flags]

Commands:
  init-config          Generate a default config file
  health-check         Run system diagnostics
  server start         Start the Minecraft server (via screen)
  server stop          Stop the server gracefully
  server restart       Restart the server
  update-mods          Check and download mod updates from Modrinth
  backup create        Create a compressed server backup
  backup list          List existing backups

Global Flags:
  -c, --config string   Config file path (default: ~/.config/craftops/config.toml)
      --debug           Enable debug logging
      --dry-run         Show what would be done without making changes
      --version         Print version and exit
```

## Configuration

Run `craftops init-config` to generate a default config, then edit it:

```toml
[minecraft]
version    = "1.20.1"
modloader  = "fabric"   # fabric | forge | quilt | neoforge

[server]
jar_name     = "server.jar"
java_flags   = ["-Xmx4G", "-Xms1G"]
stop_command = "stop"

[paths]
server  = "/home/minecraft/server"
mods    = "/home/minecraft/server/mods"
backups = "/home/minecraft/backups"

[mods]
modrinth_sources      = [
  "https://modrinth.com/mod/fabric-api",
  "https://modrinth.com/mod/sodium",
]
concurrent_downloads  = 4
max_retries           = 3
retry_delay           = 2.0   # seconds between retries

[backup]
enabled          = true
max_backups      = 5
include_logs     = false
exclude_patterns = ["*.tmp", "cache/**"]

[notifications]
discord_webhook    = ""          # optional — paste your webhook URL here
warning_intervals  = [10, 5, 1]  # minutes before restart to send warnings

[logging]
level  = "info"    # info | debug
format = "json"    # json | text
```

## Releasing

Use the helper script to bump the semver tag and trigger the GitHub Actions release pipeline:

```bash
./scripts/release.sh patch                      # 2.3.0 → 2.3.1
./scripts/release.sh minor "New mod features"   # 2.3.1 → 2.4.0
./scripts/release.sh major                      # 2.4.0 → 3.0.0
```

The script generates a changelog from commits since the last tag, prompts for confirmation, then creates an annotated git tag and pushes it. GitHub Actions handles the rest.

## Development

```bash
make build        # build to build/craftops
make test         # run tests with race detector
make lint         # run golangci-lint
make fmt          # gofmt all packages
make package      # cross-compile for linux/darwin × amd64/arm64
```

## License

[MIT](LICENSE)
