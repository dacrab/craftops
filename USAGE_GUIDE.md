<div align="center">

# � Minecraft tMod Manager - Complete Usage Guide

**Master your Minecraft server management with comprehensive examples and best practices**

[🏠 Back to README](README.md) • [🚀 Deployment Guide](DEPLOYMENT_GUIDE.md) • [🏗️ Project Structure](PROJECT_STRUCTURE.md)

</div>

---

## 📋 Table of Contents

- [� Insstallation & Setup](#-installation--setup)
- [⚙️ Configuration Deep Dive](#️-configuration-deep-dive)
- [🎮 Server Management](#-server-management)
- [🔄 Mod Management](#-mod-management)
- [💾 Backup Operations](#-backup-operations)
- [🔔 Notifications Setup](#-notifications-setup)
- [🏥 Health Monitoring](#-health-monitoring)
- [🤖 Automation & Scheduling](#-automation--scheduling)
- [🐛 Troubleshooting](#-troubleshooting)
- [⚡ Performance Optimization](#-performance-optimization)
- [🔐 Security Best Practices](#-security-best-practices)

---

## 🚀 Installation & Setup

### Quick Installation Methods

<table>
<tr>
<td width="50%">

#### 🎯 **One-Line Install (Recommended)**
```bash
curl -sSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
```

**✅ What it does:**
- Auto-detects your platform
- Downloads latest release
- Creates aliases (`cops`, `craftops`)
- Sets up configuration directory
- Adds to PATH

</td>
<td width="50%">

#### 🔧 **Manual Installation**
```bash
# Download binary
curl -L https://github.com/dacrab/craftops/releases/latest/download/craftops-linux-amd64 -o craftops

# Install system-wide
chmod +x craftops
sudo mv craftops /usr/local/bin/

# Create aliases
sudo ln -sf /usr/local/bin/craftops /usr/local/bin/mmu
```

</td>
</tr>
</table>

### Post-Installation Verification

```bash
# Verify installation
mmu --version
mmu --help

# Check available commands
mmu
```

**Expected Output:**
```
Minecraft Mod Manager v2.0.0

Available Commands:
  backup       💾 Backup management commands
  health-check 🏥 Run comprehensive system health checks
  init-config  📝 Initialize a new configuration file with defaults
  server       🎮 Minecraft server management commands
  update-mods  🔄 Update all configured mods to their latest versions
```

---

## ⚙️ Configuration Deep Dive

### Initial Configuration Setup

```bash
# Create default configuration
mmu init-config

# Create configuration in custom location
mmu init-config -o /custom/path/config.toml

# Force overwrite existing configuration
mmu init-config --force
```

### Configuration File Locations

The tool searches for configuration files in this order:
1. `--config` flag path
2. `./conf.toml` (current directory)
3. `~/.config/craftops/config.toml`
4. `/etc/craftops/config.toml`

### Complete Configuration Reference

<details>
<summary><b>🔧 Server Configuration</b></summary>

```toml
[server]
jar_name = "server.jar"              # Server JAR filename
stop_command = "stop"                # Command to stop server
max_stop_wait = 300                  # Max seconds to wait for stop
startup_timeout = 120                # Max seconds to wait for start

# Java flags for different server sizes
java_flags = [
    "-Xms4G",                        # Initial heap size
    "-Xmx8G",                        # Maximum heap size
    "-XX:+UseG1GC",                  # Use G1 garbage collector
    "-XX:+ParallelRefProcEnabled",   # Parallel reference processing
    "-XX:+UnlockExperimentalVMOptions",
    "-XX:+DisableExplicitGC",        # Disable explicit GC calls
    "-XX:+AlwaysPreTouch",           # Pre-touch memory pages
    "-XX:G1NewSizePercent=30",       # G1 new generation size
    "-XX:G1MaxNewSizePercent=40",    # G1 max new generation size
    "-XX:G1HeapRegionSize=8M",       # G1 heap region size
    "-XX:G1ReservePercent=20",       # G1 reserve percent
    "-XX:G1HeapWastePercent=5",      # G1 heap waste percent
    "-XX:G1MixedGCCountTarget=4",    # G1 mixed GC count target
    "-XX:InitiatingHeapOccupancyPercent=15",  # G1 GC trigger
    "-XX:G1MixedGCLiveThresholdPercent=90",   # G1 mixed GC threshold
    "-XX:G1RSetUpdatingPauseTimePercent=5",  # G1 RSet update pause
    "-XX:SurvivorRatio=32",          # Survivor space ratio
    "-XX:+PerfDisableSharedMem",     # Disable shared memory
    "-XX:MaxTenuringThreshold=1",    # Max tenuring threshold
    "-Dusing.aikars.flags=https://mcflags.emc.gs",  # Aikar's flags
    "-Daikars.new.flags=true"
]
```

**Java Flags for Different Server Sizes:**

| Server Size | Players | RAM | Recommended Flags |
|-------------|---------|-----|-------------------|
| **Small** | 2-10 | 4GB | `-Xms2G -Xmx4G -XX:+UseG1GC` |
| **Medium** | 10-50 | 8GB | Above + `-XX:+ParallelRefProcEnabled` |
| **Large** | 50+ | 16GB+ | Full Aikar's flags (shown above) |

</details>

<details>
<summary><b>🔄 Mod Management Configuration</b></summary>

```toml
[mods]
auto_update = true                   # Enable automatic updates
backup_before_update = true          # Create backup before updating
concurrent_downloads = 5             # Parallel download limit
max_retries = 3                     # Retry attempts for failed downloads
retry_delay = 2.0                   # Delay between retries (seconds)
timeout = 30                        # HTTP request timeout

[mods.sources]
modrinth = [
    "https://modrinth.com/mod/fabric-api",      # Essential Fabric API
    "https://modrinth.com/mod/sodium",          # Performance: Rendering
    "https://modrinth.com/mod/lithium",         # Performance: General
    "https://modrinth.com/mod/phosphor",        # Performance: Lighting
    "https://modrinth.com/mod/ferritecore",     # Performance: Memory
    "https://modrinth.com/mod/starlight",       # Performance: Lighting engine
    "https://modrinth.com/mod/krypton",         # Performance: Networking
    "https://modrinth.com/mod/lazydfu",         # Performance: DataFixerUpper
    "https://modrinth.com/mod/c2me-fabric",     # Performance: Chunk generation
]
```

**Popular Mod Categories:**

| Category | Example Mods | Purpose |
|----------|--------------|---------|
| **Performance** | Sodium, Lithium, Phosphor | Optimize server performance |
| **Utility** | JEI, WTHIT, REI | Quality of life improvements |
| **Content** | Create, Tech Reborn, AE2 | Add new gameplay mechanics |
| **World Gen** | Biomes O' Plenty, Terralith | Enhanced world generation |

</details>

<details>
<summary><b>💾 Backup Configuration</b></summary>

```toml
[backup]
enabled = true                       # Enable backup system
max_backups = 5                     # Number of backups to keep
compression_level = 6               # Compression level (1-9, higher = smaller)
include_logs = false                # Include server logs in backup

# Files and directories to exclude from backups
exclude_patterns = [
    "*.log",                        # Log files
    "*.log.*",                      # Rotated log files
    "cache/",                       # Cache directories
    "temp/",                        # Temporary files
    ".DS_Store",                    # macOS system files
    "Thumbs.db",                    # Windows system files
    "world/session.lock",           # Minecraft session lock
    "usercache.json",               # User cache (regenerated)
    "banned-*.json",                # Ban lists (optional)
]
```

**Backup Strategy Recommendations:**

| Server Type | Max Backups | Compression | Include Logs |
|-------------|-------------|-------------|--------------|
| **Development** | 10 | 3 (fast) | true |
| **Small Production** | 5 | 6 (balanced) | false |
| **Large Production** | 3 | 9 (maximum) | false |

</details>

<details>
<summary><b>🔔 Notification Configuration</b></summary>

```toml
[notifications]
discord_webhook = "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"
warning_intervals = [15, 10, 5, 1]  # Warning times in minutes
warning_message = "🚨 Server will restart in {minutes} minute(s) for mod updates! Please save your progress."
success_notifications = true         # Send success notifications
error_notifications = true          # Send error notifications
```

**Discord Webhook Setup:**
1. Go to your Discord server settings
2. Navigate to Integrations → Webhooks
3. Create a new webhook
4. Copy the webhook URL
5. Paste it into your configuration

**Message Templates:**
```toml
# Custom warning messages
warning_message = "⚠️ Maintenance in {minutes} min - Save your work!"

# You can use these placeholders:
# {minutes} - Minutes until restart
# {server} - Server name (if configured)
# {version} - Minecraft version
```

</details>

---

## 🎮 Server Management

### Basic Server Operations

```bash
# Start the server
mmu server start

# Stop the server gracefully
mmu server stop

# Restart with player warnings
mmu server restart

# Check server status
mmu server status
```

### Advanced Server Management

<details>
<summary><b>🔧 Server Status Monitoring</b></summary>

```bash
# Basic status check
mmu server status

# Detailed status with debug info
mmu --debug server status

# Status check with custom config
mmu --config /path/to/config.toml server status
```

**Status Output Example:**
```
🎮 Server Status
─────────────────
✅ Server is running
PID: 12345
Uptime: 2h 30m
Memory: 4.2GB / 8GB
```

</details>

<details>
<summary><b>🚀 Server Startup Optimization</b></summary>

**Screen Session Management:**
```bash
# View active screen sessions
screen -ls

# Attach to server console
screen -r minecraft

# Detach from console (Ctrl+A, then D)
# Kill stuck screen session
screen -S minecraft -X quit
```

**Startup Troubleshooting:**
```bash
# Start with debug logging
mmu --debug server start

# Check server logs
tail -f /path/to/server/logs/latest.log

# Verify Java installation
java -version

# Check server JAR
ls -la /path/to/server/server.jar
```

</details>

<details>
<summary><b>🛑 Graceful Shutdown Process</b></summary>

The server stop process follows these steps:
1. **Warning Phase**: Send stop command to server
2. **Grace Period**: Wait for configured timeout
3. **Force Kill**: If server doesn't stop, force termination

```bash
# Standard stop (uses configured timeout)
mmu server stop

# Stop with custom timeout (dry run to see what would happen)
mmu --dry-run server stop

# Debug stop process
mmu --debug server stop
```

**Stop Command Customization:**
```toml
[server]
stop_command = "stop"        # Standard Minecraft stop
# stop_command = "shutdown"  # Alternative for some servers
# stop_command = "/stop"     # With slash prefix
max_stop_wait = 300         # 5 minutes timeout
```

</details>

---

## 🔄 Mod Management

### Basic Mod Operations

```bash
# Update all mods to latest compatible versions
mmu update-mods

# Force update even if versions appear current
mmu update-mods --force

# Update without creating backup
mmu update-mods --no-backup

# Preview what would be updated (no changes made)
mmu update-mods --dry-run
```

### Advanced Mod Management

<details>
<summary><b>🔍 Mod Update Process</b></summary>

The mod update process follows these steps:

1. **Backup Creation** (if enabled)
2. **Version Checking** via Modrinth API
3. **Compatibility Validation** (Minecraft version + mod loader)
4. **Concurrent Downloads** (respecting rate limits)
5. **File Replacement** with backup of old versions
6. **Verification** of successful downloads

```bash
# Detailed update process with debug info
mmu --debug update-mods

# Update with custom concurrency
# (Edit config file to change concurrent_downloads)

# Update specific configuration
mmu --config /path/to/config.toml update-mods
```

</details>

<details>
<summary><b>📦 Modrinth Integration</b></summary>

**Supported URL Formats:**
```toml
[mods.sources]
modrinth = [
    # Full mod page URLs
    "https://modrinth.com/mod/fabric-api",
    "https://modrinth.com/mod/sodium",
    
    # Project IDs (shorter format)
    "P7dR8mSH",  # Fabric API project ID
    "AANobbMI",  # Sodium project ID
]
```

**Finding Mod Information:**
1. Visit the mod page on Modrinth
2. Copy the URL or find the project ID in the URL
3. Add to your configuration file
4. Run `mmu update-mods` to download

**Version Compatibility:**
- Automatically filters by Minecraft version
- Respects mod loader (Fabric, Forge, Quilt, NeoForge)
- Downloads latest compatible version only

</details>

<details>
<summary><b>⚡ Performance Tuning</b></summary>

**Download Performance:**
```toml
[mods]
concurrent_downloads = 10    # Increase for faster downloads
max_retries = 5             # Increase for unreliable connections
retry_delay = 1.0           # Decrease for faster retries
timeout = 60                # Increase for slow connections
```

**Network Optimization:**
- **Concurrent Downloads**: Balance between speed and API rate limits
- **Retry Logic**: Exponential backoff prevents API abuse
- **Timeout Settings**: Adjust based on connection quality
- **Rate Limiting**: Built-in delays respect API guidelines

**Recommended Settings by Connection:**

| Connection Type | Concurrent | Timeout | Retries |
|----------------|------------|---------|---------|
| **Fiber/Fast** | 10 | 30s | 3 |
| **Standard** | 5 | 30s | 3 |
| **Slow/Mobile** | 2 | 60s | 5 |

</details>

---

## 💾 Backup Operations

### Manual Backup Management

```bash
# Create immediate backup
mmu backup create

# List all available backups
mmu backup list

# Create backup with debug info
mmu --debug backup create
```

### Automated Backup System

<details>
<summary><b>🔄 Automatic Backup Triggers</b></summary>

Backups are automatically created:
- **Before mod updates** (if `backup_before_update = true`)
- **Before server restarts** (when triggered by mod updates)
- **On manual request** via `mmu backup create`

**Backup Naming Convention:**
```
minecraft_backup_20240101_143022.tar.gz
                 YYYYMMDD_HHMMSS
```

</details>

<details>
<summary><b>📊 Backup Management</b></summary>

**Backup List Output:**
```bash
mmu backup list
```

```
💾 Available Backups
────────────────────
Name                                    Date                 Size
minecraft_backup_20240101_143022.tar.gz 2024-01-01 14:30:22  245.3 MB
minecraft_backup_20240101_120015.tar.gz 2024-01-01 12:00:15  244.8 MB
minecraft_backup_20231231_235959.tar.gz 2023-12-31 23:59:59  243.1 MB

Total: 3 backups
```

**Backup Retention:**
- Automatically removes old backups beyond `max_backups` limit
- Keeps most recent backups based on creation time
- Logs cleanup operations for audit trail

</details>

<details>
<summary><b>🗜️ Compression & Exclusion</b></summary>

**Compression Levels:**
- **1-3**: Fast compression, larger files
- **4-6**: Balanced compression (recommended)
- **7-9**: Maximum compression, slower

**Smart Exclusion Patterns:**
```toml
exclude_patterns = [
    "*.log",                    # All log files
    "*.log.*",                  # Rotated logs (log.1, log.2, etc.)
    "cache/",                   # Cache directories
    "temp/",                    # Temporary files
    "world/session.lock",       # Minecraft session lock
    "world/*/poi/",            # Points of interest cache
    "world/*/region/*.mca.tmp", # Temporary region files
]
```

**Backup Size Optimization:**
- Exclude unnecessary files (logs, cache, temp)
- Use appropriate compression level
- Regular cleanup of old backups
- Monitor disk space usage

</details>

---

## 🔔 Notifications Setup

### Discord Integration

<details>
<summary><b>🔧 Discord Webhook Setup</b></summary>

**Step-by-Step Setup:**

1. **Create Webhook in Discord:**
   - Go to Server Settings → Integrations
   - Click "Create Webhook"
   - Choose channel for notifications
   - Copy webhook URL

2. **Configure in MMM:**
   ```toml
   [notifications]
   discord_webhook = "https://discord.com/api/webhooks/123456789/abcdefghijklmnop"
   ```

3. **Test Notifications:**
   ```bash
   # Test with a dry run (shows what would be sent)
   mmu --dry-run update-mods
   
   # Actual test
   mmu update-mods
   ```

</details>

<details>
<summary><b>📱 Notification Types</b></summary>

**Success Notifications:**
- ✅ Mod updates completed
- ✅ Server started successfully
- ✅ Backup created
- ✅ Health check passed

**Error Notifications:**
- ❌ Mod update failures
- ❌ Server start/stop failures
- ❌ Backup creation errors
- ❌ Configuration errors

**Warning Notifications:**
- ⚠️ Server restart warnings
- ⚠️ Health check issues
- ⚠️ Disk space warnings

**Example Discord Message:**
```
🔄 Mod Update Started
Starting automated mod update process...

✅ Mod Update Complete
Updated 3 mods successfully:
• Fabric API v0.92.0
• Sodium v0.5.3
• Lithium v0.11.2

Server restarted successfully!
```

</details>

<details>
<summary><b>⏰ Warning System</b></summary>

**Restart Warning Configuration:**
```toml
[notifications]
warning_intervals = [15, 10, 5, 1]  # Minutes before restart
warning_message = "🚨 Server restart in {minutes} minute(s) for updates!"
```

**Warning Timeline:**
- **15 minutes**: First warning sent
- **10 minutes**: Second warning sent
- **5 minutes**: Third warning sent
- **1 minute**: Final warning sent
- **0 minutes**: Server restart begins

**Custom Warning Messages:**
```toml
# Professional message
warning_message = "⚠️ Scheduled maintenance in {minutes} minutes. Please save your progress."

# Casual message
warning_message = "🔧 Quick restart in {minutes} min for mod updates!"

# Detailed message
warning_message = "🚨 Server restart in {minutes} minute(s) for mod updates. Current players will be disconnected. Please save your work!"
```

</details>

---

## 🏥 Health Monitoring

### Comprehensive Health Checks

```bash
# Full system health check
mmu health-check

# Health check with debug information
mmu --debug health-check

# Health check with custom configuration
mmu --config /path/to/config.toml health-check
```

### Health Check Categories

<details>
<summary><b>📁 Path Validation</b></summary>

**Checks Performed:**
- ✅ Directory existence
- ✅ Read/write permissions
- ✅ Disk space availability
- ✅ Path accessibility

**Example Output:**
```
Component                      Status   Details
─────────                      ──────   ───────
Server directory               ✅        OK
Mods directory                 ✅        OK (15 mods found)
Backups directory              ✅        OK (3 backups found)
Logs directory                 ✅        OK
```

**Common Issues & Solutions:**
- **❌ Directory not found**: Create missing directories
- **❌ No write permission**: Fix ownership/permissions
- **⚠️ Low disk space**: Clean up old files/backups

</details>

<details>
<summary><b>🌐 API Connectivity</b></summary>

**Modrinth API Checks:**
- ✅ API endpoint accessibility
- ✅ Response time measurement
- ✅ Rate limit status
- ✅ Authentication (if required)

**Example Output:**
```
Component                      Status   Details
─────────                      ──────   ───────
Modrinth API                   ✅        API accessible (120ms)
```

**Troubleshooting API Issues:**
```bash
# Test API connectivity manually
curl -I https://api.modrinth.com/v2/

# Check DNS resolution
nslookup api.modrinth.com

# Test with debug logging
mmu --debug health-check
```

</details>

<details>
<summary><b>⚙️ System Requirements</b></summary>

**Java Runtime Validation:**
- ✅ Java installation
- ✅ Version compatibility
- ✅ Memory allocation capability

**Server Environment:**
- ✅ Screen command availability
- ✅ Server JAR file existence
- ✅ Configuration file validity

**Example Output:**
```
Component                      Status   Details
─────────                      ──────   ───────
Java Runtime                   ✅        OpenJDK 17.0.2
Server JAR                     ✅        Found (45.2 MB)
Screen Command                 ✅        Available
```

</details>

---

## 🤖 Automation & Scheduling

### Cron Job Setup

<details>
<summary><b>⏰ Scheduled Mod Updates</b></summary>

**Daily Updates at 3 AM:**
```bash
# Edit crontab
crontab -e

# Add this line for daily updates
0 3 * * * /usr/local/bin/mmu update-mods >> /var/log/mmu-cron.log 2>&1
```

**Weekly Updates on Sunday:**
```bash
# Weekly updates at 2 AM on Sunday
0 2 * * 0 /usr/local/bin/mmu update-mods && /usr/local/bin/mmu server restart
```

**Advanced Scheduling Examples:**
```bash
# Every 6 hours
0 */6 * * * /usr/local/bin/mmu update-mods

# Weekdays only at 4 AM
0 4 * * 1-5 /usr/local/bin/mmu update-mods

# With custom config
0 3 * * * /usr/local/bin/mmu --config /etc/mmu/config.toml update-mods
```

</details>

<details>
<summary><b>🔄 Systemd Integration</b></summary>

**Create Systemd Service:**
```ini
# /etc/systemd/system/craftops.service
[Unit]
Description=Minecraft Mod Manager
After=network.target

[Service]
Type=oneshot
User=minecraft
WorkingDirectory=/home/minecraft
ExecStart=/usr/local/bin/mmu update-mods
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Create Systemd Timer:**
```ini
# /etc/systemd/system/craftops.timer
[Unit]
Description=Run Minecraft Mod Manager daily
Requires=craftops.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

**Enable and Start:**
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable timer
sudo systemctl enable craftops.timer

# Start timer
sudo systemctl start craftops.timer

# Check status
sudo systemctl status craftops.timer
```

</details>

<details>
<summary><b>📊 Monitoring & Logging</b></summary>

**Log Rotation Setup:**
```bash
# /etc/logrotate.d/craftops
/var/log/mmu-*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 minecraft minecraft
}
```

**Monitoring Script:**
```bash
#!/bin/bash
# /usr/local/bin/mmu-monitor.sh

LOG_FILE="/var/log/mmu-monitor.log"
CONFIG_FILE="/etc/mmu/config.toml"

echo "$(date): Starting MMU health check" >> "$LOG_FILE"

if /usr/local/bin/mmu --config "$CONFIG_FILE" health-check >> "$LOG_FILE" 2>&1; then
    echo "$(date): Health check passed" >> "$LOG_FILE"
else
    echo "$(date): Health check failed" >> "$LOG_FILE"
    # Send alert (email, Discord, etc.)
fi
```

</details>

---

## 🐛 Troubleshooting

### Common Issues & Solutions

<details>
<summary><b>❌ Server Management Issues</b></summary>

**"Screen command not found"**
```bash
# Install screen on Ubuntu/Debian
sudo apt update && sudo apt install screen

# Install screen on CentOS/RHEL
sudo yum install screen

# Install screen on macOS
brew install screen
```

**"Server JAR not found"**
```bash
# Check if file exists
ls -la /path/to/server/server.jar

# Update jar_name in config if different
nano conf.toml
# Change: jar_name = "your-actual-jar-name.jar"

# Download server JAR if missing
cd /path/to/server
wget https://papermc.io/api/v2/projects/paper/versions/1.20.1/builds/196/downloads/paper-1.20.1-196.jar -O server.jar
```

**"Java not found"**
```bash
# Install Java 17+ on Ubuntu/Debian
sudo apt update && sudo apt install openjdk-17-jre-headless

# Install Java 17+ on CentOS/RHEL
sudo yum install java-17-openjdk-headless

# Install Java on macOS
brew install openjdk@17

# Verify installation
java -version
```

</details>

<details>
<summary><b>🔄 Mod Update Issues</b></summary>

**"No compatible versions found"**
- Check Minecraft version in config matches your server
- Verify mod loader (Fabric/Forge) is correct
- Ensure mod supports your Minecraft version

**"API rate limit exceeded"**
```bash
# Reduce concurrent downloads
nano conf.toml
# Change: concurrent_downloads = 2

# Increase retry delay
# Change: retry_delay = 5.0
```

**"Download failed"**
```bash
# Test internet connectivity
curl -I https://api.modrinth.com/v2/

# Check DNS resolution
nslookup api.modrinth.com

# Run with debug logging
mmu --debug update-mods
```

**"Permission denied writing to mods directory"**
```bash
# Fix ownership
sudo chown -R minecraft:minecraft /path/to/server/mods

# Fix permissions
chmod -R 755 /path/to/server/mods
```

</details>

<details>
<summary><b>💾 Backup Issues</b></summary>

**"Backup directory not writable"**
```bash
# Create backup directory
sudo mkdir -p /path/to/backups

# Fix ownership
sudo chown -R minecraft:minecraft /path/to/backups

# Fix permissions
chmod -R 755 /path/to/backups
```

**"Backup creation failed"**
```bash
# Check disk space
df -h /path/to/backups

# Check for file locks
lsof +D /path/to/server

# Run with debug logging
mmu --debug backup create
```

**"Backup too large"**
```toml
# Increase compression level
compression_level = 9

# Add more exclusion patterns
exclude_patterns = [
    "*.log", "cache/", "temp/",
    "world/*/poi/",           # Points of interest cache
    "world/*/region/*.mca.tmp" # Temporary region files
]
```

</details>

<details>
<summary><b>🔔 Notification Issues</b></summary>

**"Discord notifications not working"**
```bash
# Test webhook URL manually
curl -X POST "YOUR_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{"content": "Test message from MMU"}'

# Check webhook URL format
# Should be: https://discord.com/api/webhooks/ID/TOKEN

# Run with debug logging
mmu --debug update-mods
```

**"Webhook URL invalid"**
- Ensure webhook URL is complete and correct
- Check Discord server permissions
- Verify webhook hasn't been deleted
- Test with a simple curl command

</details>

### Debug Mode & Logging

<details>
<summary><b>🔍 Debug Information</b></summary>

**Enable Debug Mode:**
```bash
# Global debug flag
mmu --debug health-check
mmu --debug update-mods
mmu --debug server start

# Debug with custom config
mmu --debug --config /path/to/config.toml health-check
```

**Log File Locations:**
```bash
# Default log location
~/.local/share/craftops/logs/craftops.log

# View recent logs
tail -f ~/.local/share/craftops/logs/craftops.log

# Search for errors
grep ERROR ~/.local/share/craftops/logs/craftops.log

# Search for specific operations
grep "update-mods" ~/.local/share/craftops/logs/craftops.log
```

**Log Configuration:**
```toml
[logging]
level = "DEBUG"              # Enable debug logging
format = "text"              # Human-readable format
file_enabled = true          # Enable file logging
console_enabled = true       # Enable console logging
```

</details>

---

## ⚡ Performance Optimization

### System Performance

<details>
<summary><b>🚀 Java Optimization</b></summary>

**Memory Allocation Guidelines:**

| Server Size | Players | RAM Available | Recommended Allocation |
|-------------|---------|---------------|----------------------|
| **Tiny** | 1-5 | 4GB | `-Xms2G -Xmx3G` |
| **Small** | 5-15 | 8GB | `-Xms4G -Xmx6G` |
| **Medium** | 15-30 | 16GB | `-Xms8G -Xmx12G` |
| **Large** | 30+ | 32GB+ | `-Xms16G -Xmx24G` |

**Optimized Java Flags:**
```toml
# High-performance configuration (Aikar's flags)
java_flags = [
    "-Xms8G", "-Xmx8G",
    "-XX:+UseG1GC",
    "-XX:+ParallelRefProcEnabled",
    "-XX:MaxGCPauseMillis=200",
    "-XX:+UnlockExperimentalVMOptions",
    "-XX:+DisableExplicitGC",
    "-XX:+AlwaysPreTouch",
    "-XX:G1NewSizePercent=30",
    "-XX:G1MaxNewSizePercent=40",
    "-XX:G1HeapRegionSize=8M",
    "-XX:G1ReservePercent=20",
    "-XX:G1HeapWastePercent=5",
    "-XX:G1MixedGCCountTarget=4",
    "-XX:InitiatingHeapOccupancyPercent=15",
    "-XX:G1MixedGCLiveThresholdPercent=90",
    "-XX:G1RSetUpdatingPauseTimePercent=5",
    "-XX:SurvivorRatio=32",
    "-XX:+PerfDisableSharedMem",
    "-XX:MaxTenuringThreshold=1",
    "-Dusing.aikars.flags=https://mcflags.emc.gs",
    "-Daikars.new.flags=true"
]
```

</details>

<details>
<summary><b>📡 Network Optimization</b></summary>

**Download Performance Tuning:**
```toml
[mods]
concurrent_downloads = 8     # Optimal for most connections
max_retries = 3             # Balance between reliability and speed
retry_delay = 1.5           # Faster retries for good connections
timeout = 45                # Longer timeout for large mods
```

**Connection-Specific Settings:**

| Connection | Concurrent | Timeout | Retry Delay |
|------------|------------|---------|-------------|
| **Gigabit** | 10-15 | 30s | 1.0s |
| **Fast Broadband** | 5-8 | 45s | 1.5s |
| **Standard** | 3-5 | 60s | 2.0s |
| **Slow/Mobile** | 1-2 | 90s | 3.0s |

</details>

<details>
<summary><b>💾 Storage Optimization</b></summary>

**Backup Optimization:**
```toml
[backup]
compression_level = 6        # Balanced compression
max_backups = 3             # Reduce for space-constrained systems

# Aggressive exclusion for large servers
exclude_patterns = [
    "*.log", "*.log.*",
    "cache/", "temp/",
    "world/*/poi/",           # POI cache (regenerated)
    "world/*/region/*.mca.tmp", # Temporary region files
    "world/*/data/raids.dat", # Raid data (regenerated)
    "usercache.json",         # User cache (regenerated)
    "world/stats/",           # Player statistics (optional)
]
```

**Disk Space Monitoring:**
```bash
# Check disk usage
df -h /path/to/server
df -h /path/to/backups

# Find large files
find /path/to/server -type f -size +100M -ls

# Clean up old logs
find /path/to/server/logs -name "*.log.gz" -mtime +7 -delete
```

</details>

---

## 🔐 Security Best Practices

### File System Security

<details>
<summary><b>👤 User & Permission Management</b></summary>

**Create Dedicated User:**
```bash
# Create minecraft user
sudo useradd -r -m -d /home/minecraft -s /bin/bash minecraft

# Set up directory structure
sudo mkdir -p /home/minecraft/{server,backups}
sudo chown -R minecraft:minecraft /home/minecraft

# Switch to minecraft user for operations
sudo -u minecraft mmu health-check
```

**Secure File Permissions:**
```bash
# Server directory permissions
chmod 750 /home/minecraft/server
chmod 640 /home/minecraft/server/*.jar
chmod 644 /home/minecraft/server/*.properties

# Configuration file permissions
chmod 600 /home/minecraft/.config/craftops/config.toml

# Backup directory permissions
chmod 750 /home/minecraft/backups
chmod 640 /home/minecraft/backups/*.tar.gz
```

**SELinux Configuration (if applicable):**
```bash
# Set SELinux contexts
sudo setsebool -P httpd_can_network_connect 1
sudo semanage fcontext -a -t bin_t "/usr/local/bin/mmu"
sudo restorecon -v /usr/local/bin/mmu
```

</details>

<details>
<summary><b>🔒 Configuration Security</b></summary>

**Secure Configuration Storage:**
```bash
# Create secure config directory
sudo mkdir -p /etc/craftops
sudo chown root:minecraft /etc/craftops
sudo chmod 750 /etc/craftops

# Secure config file
sudo chmod 640 /etc/craftops/config.toml
```

**Environment Variable Security:**
```bash
# Use environment variables for sensitive data
export MMM_DISCORD_WEBHOOK="https://discord.com/api/webhooks/..."

# Reference in config
# discord_webhook = "${MMM_DISCORD_WEBHOOK}"
```

**Configuration Validation:**
```bash
# Regular security audits
mmu --config /etc/craftops/config.toml health-check

# Check for sensitive data in logs
grep -i "webhook\|token\|password" /var/log/mmu-*.log
```

</details>

<details>
<summary><b>🌐 Network Security</b></summary>

**Firewall Configuration:**
```bash
# Allow only necessary ports
sudo ufw allow 25565/tcp  # Minecraft server
sudo ufw deny 80/tcp      # Block HTTP if not needed
sudo ufw enable

# Monitor network connections
sudo netstat -tulpn | grep :25565
```

**API Security:**
- Use HTTPS for all API communications
- Implement rate limiting in configuration
- Monitor API usage and errors
- Keep webhook URLs private and secure

**Backup Security:**
```bash
# Encrypt sensitive backups
gpg --symmetric --cipher-algo AES256 backup.tar.gz

# Secure backup transfer
rsync -avz --delete /home/minecraft/backups/ user@backup-server:/secure/backups/

# Remote backup with SSH keys
ssh-keygen -t ed25519 -f ~/.ssh/backup_key
ssh-copy-id -i ~/.ssh/backup_key user@backup-server
```

</details>

### Monitoring & Auditing

<details>
<summary><b>📊 Security Monitoring</b></summary>

**Log Monitoring:**
```bash
# Monitor authentication attempts
sudo tail -f /var/log/auth.log | grep minecraft

# Monitor file access
sudo auditctl -w /home/minecraft/server -p wa -k minecraft-access

# Check audit logs
sudo ausearch -k minecraft-access
```

**Automated Security Checks:**
```bash
#!/bin/bash
# /usr/local/bin/mmu-security-check.sh

# Check file permissions
find /home/minecraft -type f -perm /o+w -ls

# Check for suspicious processes
ps aux | grep -E "(java|screen)" | grep minecraft

# Verify configuration integrity
mmu --config /etc/craftops/config.toml health-check

# Check for unauthorized modifications
find /home/minecraft -type f -newer /var/log/mmu-last-check.log -ls
touch /var/log/mmu-last-check.log
```

</details>

---

<div align="center">

**🎮 Happy Server Managing!**

[🏠 Back to README](README.md) • [🚀 Deployment Guide](DEPLOYMENT_GUIDE.md) • [🏗️ Project Structure](PROJECT_STRUCTURE.md)

---

*This guide is continuously updated. For the latest information, visit the [GitHub repository](https://github.com/dacrab/craftops).*

</div>