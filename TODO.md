## craftOPS plan

CLI to operate manually-installed Minecraft servers (cloud or self-hosted) with safe mod updates, backups, notifications, and simple server control.

### High-level architecture
- **Entry point**: `cmd/craftops/main.go`
- **CLI layer**: `internal/cli/` (Cobra commands)
  - Root command configures global flags and wires a `ServiceFactory`
  - Subcommands: `backup`, `health-check`, `init-config`, `update-mods`, `server`
- **Services layer**: `internal/services/`
  - `BackupService`, `ModService`, `ServerService`, `NotificationService` implementing interfaces in `internal/services/interfaces.go`
- **Configuration**: `internal/config/`
  - `config.go` (structs, defaults, load/save), `validator.go` (strong validation, auto-create parent dirs for logs/backups)
- **Logging**: `internal/logger/` (Zap with console/file, json/text)
- **Metrics**: `internal/metrics/` (in-memory counters, summaries, periodic logging)
- **View (UX)**: `internal/view/` (colors, banners, tables)
- **Graceful shutdown**: `internal/shutdown/` (signal handling, cleanup)

### Project structure (key paths)
```text
cmd/craftops/main.go              # entrypoint
internal/cli/                     # cobra commands and factory
internal/config/                  # config types, defaults, validation
internal/logger/                  # zap logger setup
internal/metrics/                 # metrics collector
internal/services/                # backup/mods/server/notification services
internal/shutdown/                # graceful shutdown helper
internal/view/                    # pretty terminal output
tests/                            # CLI tests + service mocks
```

### CLI commands
- **Global flags**: `--config|-c`, `--debug|-d`, `--dry-run|-n`, `-v|--version`
- **init-config | init | configure**: create `config.toml` with sensible defaults
- **update-mods | update | mods | upgrade [--force|-f] [--no-backup]**: update mods
- **backup|bk create|now|run | list|ls**: create or list backups
- **backup|bk restore <archive> [-f|--force]**: restore archive into server dir
- **server|srv start|stop|restart|status**: manage server via `screen`
- Top-level shortcuts: `start`, `stop`, `restart`, `status`, `update`
- **health-check | health | check**: validate paths, mods, server JAR, Java, backup, notifications
 - **mods|m list**: list installed mods

### Configuration (TOML)
Validated in `internal/config/validator.go`.
UX simplifications:
- Paths accept `~` and relative values; they are auto-normalized to absolute
- Config path can be set via env `CRAFTOPS_CONFIG`
```toml
debug = false
dry_run = false

[minecraft]
version = "1.20.1"
modloader = "fabric" # fabric|forge|quilt|neoforge

[paths]
server = "/home/USER/minecraft/server"
mods = "/home/USER/minecraft/server/mods"
backups = "/home/USER/minecraft/backups"
logs = "/home/USER/.local/share/craftops/logs"

[server]
jar_name = "server.jar"
java_flags = ["-Xms4G", "-Xmx4G", "-XX:+UseG1GC"]
stop_command = "stop"
max_stop_wait = 300
startup_timeout = 120

[mods]
auto_update = true
backup_before_update = true
concurrent_downloads = 5
max_retries = 3
retry_delay = 2.0
timeout = 30
modrinth_sources = [ ] # e.g. "https://modrinth.com/mod/<project>"

[backup]
enabled = true
max_backups = 5
compression_level = 6
include_logs = false
exclude_patterns = ["*.log", "*.log.*", "cache/", "temp/"]

[notifications]
discord_webhook = ""
warning_intervals = [15,10,5,1]
warning_message = "Server will restart in {minutes} minute(s) for mod updates"
success_notifications = true
error_notifications = true

[logging]
level = "INFO"          # DEBUG|INFO|WARNING|ERROR|CRITICAL
format = "json"          # json|text
file_enabled = true
console_enabled = true
max_file_size = "10MB"
backup_count = 5
```

### Services overview
- **BackupService**
  - Creates `.tar.gz` backups of `paths.server`, applies `exclude_patterns` and `include_logs`
  - Ensures `paths.backups` exists; enforces `max_backups` retention
  - Dry-run prints intent; health checks directory and retention settings
- **ModService**
  - Parses Modrinth URLs, queries latest compatible versions for `[minecraft]` loader/version
  - Concurrent downloads controlled by `concurrent_downloads`; http `timeout`
  - Dry-run yields placeholder update; `ListInstalledMods()` helper available
  - Health checks: mods dir, sources configured, Modrinth API reachability
- **ServerService**
  - Uses `screen` to `start/stop/restart/status`; validates JAR exists; waits up to `max_stop_wait`
  - Health checks: server dir, JAR presence/size, Java runtime
- **NotificationService**
  - Discord webhook embeds; restart warnings at `warning_intervals`; success/error toggles
  - Dry-run logs instead of sending; health checks webhook format/settings

### UX & Logging
- Colored output, sections, banners, and tables (`internal/view`)
- Progress spinners for longer tasks (backups, mod updates)
- Zap logging to console and/or file, json or text (`internal/logger`)

### Roadmap (@TODO)
- Add "mods list" command using `ModService.ListInstalledMods()`
- Add "backup restore <file>" with safety checks and confirmation
- Implement network retries/backoff in `ModService` using `max_retries` and `retry_delay`
- Rollback on failed restart after updates (use latest backup automatically)
- Improve `server status` to surface PID/uptime/memory on Linux
- Support `tmux` and non-`screen` environments; detect availability
- Add CurseForge provider and generic URL adapters (beyond Modrinth)
- Export metrics (Prometheus or `metrics show` command)
- Extend tests for services (reintroduce service-level tests; currently CLI heavy)
- Windows/macOS support and service manager integration

### Quick start
1) Generate config: `craftops init-config -o ./config.toml`
2) Add Modrinth URLs to `[mods.modrinth_sources]`
3) Validate: `craftops health-check`
4) Update mods: `craftops update-mods` (uses backup by default)