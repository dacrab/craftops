# CraftOps

A small Go CLI for the unglamorous side of running a modded Minecraft server.

CraftOps starts and stops the server, keeps Modrinth mods current, makes backups, and pings Discord when something happens. One TOML file holds the settings.

## Install

Works on Linux and macOS. GNU screen and Java 17+ are required to run the server itself.

```bash
curl -fsSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
```

Building from source needs Go 1.27+:

```bash
make install
```

A Docker image is available at `ghcr.io/dacrab/craftops` if you prefer containers.

## First run

```bash
craftops init -o ~/.config/craftops/config.toml
craftops health
craftops server start
```

Edit the generated config first so the paths, mods, and backup folder point at your setup. `craftops --help` lists the rest.

## What it handles

- Server lifecycle through GNU screen: start, stop, restart, and status
- Modrinth mod updates in one command, with retries and a dry-run mode
- Compressed `.tar.gz` backups with retention limits and exclude patterns
- Optional Discord warnings before restarts and alerts on failures
- Health checks for paths, dependencies, and API access

Backups are plain archives, so keep a copy somewhere safe. Mods come from the Modrinth project pages you list in the config.

MIT licensed — see [LICENSE](LICENSE).
