# CraftOps Usage

Minimal, task‑focused guide to install, configure, and operate CraftOps.

## Install

- One‑liner: `curl -sSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash`
- From source: `make install-system`
- Docker: `docker pull ghcr.io/dacrab/craftops:latest`

## Global flags

- `--config, -c` Path to config file (TOML)
- `--debug` Enable debug logging
- `--dry-run` Print actions without changing the system

## Commands

- `init-config` Generate a default config file
- `health-check` Validate paths, Java, screen, API, and notifications
- `update-mods` Update Modrinth mods; flags: `--force`, `--no-backup`
- `server start|stop|restart|status` Manage the server via screen
- `backup create|list` Create and list tar.gz backups with retention

Examples:

```bash
craftops init-config -o ./config.toml --force
craftops --config ./config.toml health-check
craftops update-mods --no-backup
craftops server restart
craftops backup create && craftops backup list
```

## Minimal config (TOML)

```toml
debug = false
dry_run = false

[minecraft]
version = "1.20.1"
modloader = "fabric"

[paths]
server  = "/home/minecraft/server"
mods    = "/home/minecraft/server/mods"
backups = "/home/minecraft/backups"
logs    = "/home/minecraft/.local/share/craftops/logs"

[server]
jar_name        = "server.jar"
java_flags      = ["-Xms4G", "-Xmx4G", "-XX:+UseG1GC"]
stop_command    = "stop"
max_stop_wait   = 300
startup_timeout = 120
session_name    = "minecraft"

[mods]
concurrent_downloads = 5
max_retries          = 3
retry_delay          = 2.0
timeout              = 30
modrinth_sources     = []

[backup]
enabled           = true
max_backups       = 5
compression_level = 6
include_logs      = false
exclude_patterns  = ["*.log", "cache/", "temp/"]

[notifications]
discord_webhook       = ""
warning_intervals     = [15, 10, 5, 1]
warning_message       = "Server will restart in {minutes} minute(s) for mod updates"
success_notifications = true
error_notifications   = true

[logging]
level           = "INFO"
format          = "json"
file_enabled    = true
console_enabled = true
```

Notes:
- Server management requires `screen` and Java 17+ on the host.
- `--dry-run` is safe for CI smoke checks and cron previews.
