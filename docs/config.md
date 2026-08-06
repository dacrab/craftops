# Configuration

Run `craftops init` to generate a default config, then edit it:

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
