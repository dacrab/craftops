# CraftOps

Modern CLI for Minecraft server operations and Modrinth mod management — built with Go.

## Features

- **Lifecycle** — Start, stop, and restart your server via GNU screen sessions
- **Mods** — Automated updates from Modrinth with concurrent downloads, retries, and dry-run support
- **Backups** — Compressed `.tar.gz` archives with configurable retention and glob-based exclusion patterns
- **Alerts** — Discord webhook notifications for restarts and warnings
- **Health** — Integrated diagnostic suite for paths, dependencies, and API connectivity

## Requirements

- Linux or macOS (amd64 or arm64)
- GNU screen
- Java 17+ (bundled in the Docker image via `openjdk17-jre-headless`)

## Install

### One-liner (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
```

Pin a specific version:

```bash
VERSION=v2.8.0 bash <(curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh)
```

### From source

Requires Go 1.26+:

```bash
make install
```

### Docker

No prebuilt image is published yet; build one locally from the repo root:

```bash
docker build -t craftops .
docker run -it --rm -v "$PWD/data:/minecraft" craftops server status
```

Mount your config at `/config/config.toml` (or `/minecraft/config.toml`) and it
is picked up automatically.

## Quick Start

```bash
# 1. Generate a default config
craftops init

# 2. Edit the config
$EDITOR ~/.config/craftops/config.toml

# 3. Verify everything looks good
craftops health

# 4. Start your server
craftops server start
```

See `craftops --help` for all commands and flags.

## Documentation

- [Configuration reference](docs/config.md)
- [Contributing & development](CONTRIBUTING.md)

## License

[MIT](LICENSE)
