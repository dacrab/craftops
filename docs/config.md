# Configuration

Run `craftops init` to generate a default config, then edit it. Below is the full set of options with their defaults.

```toml
[minecraft]
version   = "1.20.1"
modloader = "fabric"   # fabric | forge | quilt | neoforge

[server]
jar_name        = "server.jar"
stop_command    = "stop"
session_name    = "minecraft"
max_stop_wait   = 300       # seconds to wait for a graceful stop
startup_timeout = 120       # seconds to wait for the server to come up
java_flags = [
  "-Xms4G", "-Xmx4G", "-XX:+UseG1GC",
  "-XX:+ParallelRefProcEnabled", "-XX:+UnlockExperimentalVMOptions",
  "-XX:+DisableExplicitGC", "-XX:+AlwaysPreTouch",
]

[paths]
server  = "/home/minecraft/server"
mods    = "/home/minecraft/server/mods"
backups = "/home/minecraft/backups"
logs    = "/home/minecraft/logs"

[mods]
modrinth_sources     = [
  "https://modrinth.com/mod/fabric-api",
  "https://modrinth.com/mod/sodium",
]
concurrent_downloads = 5
max_retries          = 3
retry_delay          = 2.0   # seconds between retries
timeout              = 30    # seconds per request

[backup]
enabled            = true
max_backups        = 5
compression_level  = 6
exclude_patterns   = ["*.log", "*.log.*", "cache/", "temp/", ".DS_Store", "Thumbs.db"]

[notifications]
discord_webhook       = ""          # optional — paste your webhook URL here
warning_message       = "Server will restart in {minutes} minute(s) for mod updates"
warning_intervals     = [15, 10, 5, 1]  # minutes before restart to send warnings
timeout               = 30         # seconds per webhook request
success_notifications = true
error_notifications   = true

[logging]
level          = "INFO"   # DEBUG | INFO | WARNING | ERROR | CRITICAL
format         = "json"   # json | text
file_enabled   = true
console_enabled = true
```
