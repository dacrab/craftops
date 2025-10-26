# CraftOps

Modern CLI for Minecraft server ops and Modrinth‑based mod updates. Minimal, fast, and script‑friendly.

## Install

- One‑liner: `curl -sSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash`
- From source: `make install-system`
- Docker: `docker pull ghcr.io/dacrab/craftops:latest`

## Quick start

```bash
craftops init-config                  # generate default config.toml
craftops --config ./config.toml health-check
craftops update-mods                  # update mods from Modrinth
craftops server restart               # restart via screen with warnings
```

## Commands

- `init-config` create config
- `health-check` validate paths, Java, screen, API, notifications
- `update-mods [--force] [--no-backup]` update mods
- `server start|stop|restart|status` manage server
- `backup create|list` backups with retention

Global flags: `--config, -c`, `--debug`, `--dry-run`.

## Minimal config

See a short example and all options in docs/usage.md.

## Docs

- Usage: docs/usage.md
- Deployment: docs/deployment.md

## License

MIT. See LICENSE.