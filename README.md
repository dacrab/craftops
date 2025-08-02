<div align="center">

# 🎮 CraftOps

**A powerful, modern CLI tool for Minecraft server operations and automated mod management**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/dacrab/craftops?style=for-the-badge&logo=github)](https://github.com/dacrab/craftops/releases)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Available-2496ED?style=for-the-badge&logo=docker)](https://github.com/dacrab/craftops/pkgs/container/craftops)

[**🚀 Quick Install**](#-quick-installation) • [**📖 Documentation**](#-documentation) • [**🎯 Features**](#-features) • [**💡 Examples**](#-usage-examples)

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔄 **Automated Mod Management**
- **Modrinth Integration**: Full API support with version compatibility
- **Concurrent Downloads**: Parallel processing for faster updates
- **Smart Retry Logic**: Handles network issues gracefully
- **Dry Run Mode**: Preview changes before applying

### 🎮 **Server Lifecycle Management**
- **Start/Stop/Restart**: Full server control via screen sessions
- **Status Monitoring**: Real-time server status checking
- **Graceful Shutdown**: Configurable stop timeouts
- **Player Warnings**: Discord notifications before restarts

</td>
<td width="50%">

### 💾 **Intelligent Backup System**
- **Automatic Backups**: Before every mod update
- **Compression**: Efficient tar.gz with configurable levels
- **Retention Policies**: Automatic cleanup of old backups
- **Selective Exclusion**: Skip logs, cache, and temp files

### 🔔 **Smart Notifications**
- **Discord Integration**: Rich webhook notifications
- **Restart Warnings**: Configurable warning intervals
- **Success/Error Alerts**: Comprehensive status updates
- **Customizable Messages**: Template-based notifications

</td>
</tr>
</table>

### 🏥 **Health Monitoring & Validation**
- **System Checks**: Validate paths, permissions, and dependencies
- **API Connectivity**: Test Modrinth API access
- **Configuration Validation**: Comprehensive config verification
- **Detailed Reporting**: Color-coded status with actionable feedback

---

## 🚀 Quick Installation

### One-Line Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/dacrab/craftops/main/install.sh | bash
```

**✅ What this does:**
- 🔍 Auto-detects your platform (Linux/macOS, x64/ARM64)
- 📥 Downloads the latest release binary
- 🔗 Installs the `craftops` command globally
- ⚙️ Sets up default configuration
- 🛣️ Adds to PATH automatically

### Alternative Installation Methods

<details>
<summary><b>📦 Manual Installation</b></summary>

```bash
# Download for your platform
curl -L https://github.com/dacrab/craftops/releases/latest/download/craftops-linux-amd64 -o craftops

# Install system-wide
chmod +x craftops
sudo mv craftops /usr/local/bin/

# No aliases needed - craftops is short and memorable
```

</details>

<details>
<summary><b>🐳 Docker Installation</b></summary>

```bash
# Pull the latest image
docker pull ghcr.io/dacrab/craftops:latest

# Run with volume mounts
docker run --rm \
  -v /path/to/server:/minecraft/server \
  -v /path/to/config:/config \
  ghcr.io/dacrab/craftops:latest \
  health-check
```

</details>

<details>
<summary><b>🔨 Build from Source</b></summary>

```bash
git clone https://github.com/dacrab/craftops.git
cd craftops
make install-system  # Requires sudo for system-wide install
```

</details>

---

## 🎯 Quick Start

### 1️⃣ **Initialize Configuration**
```bash
craftops init-config
```

### 2️⃣ **Configure Your Setup**
```bash
nano config.toml  # Edit with your server details
```

<details>
<summary><b>📝 Example Configuration</b></summary>

```toml
[minecraft]
version = "1.20.1"
modloader = "fabric"

[paths]
server = "/home/minecraft/server"
mods = "/home/minecraft/server/mods"
backups = "/home/minecraft/backups"

[mods.sources]
modrinth = [
    "https://modrinth.com/mod/fabric-api",
    "https://modrinth.com/mod/sodium",
    "https://modrinth.com/mod/lithium"
]

[notifications]
discord_webhook = "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"
warning_intervals = [15, 10, 5, 1]
```

</details>

### 3️⃣ **Verify Setup**
```bash
craftops health-check
```

### 4️⃣ **Start Managing Your Server**
```bash
craftops update-mods     # Update all mods
craftops server restart  # Restart with player warnings
```

---

## 💡 Usage Examples

### 🔄 **Mod Management**
```bash
# Update all mods to latest compatible versions
craftops update-mods

# Force update even if versions appear current
craftops update-mods --force

# Preview what would be updated (no changes made)
craftops update-mods --dry-run

# Update without creating backup
craftops update-mods --no-backup
```

### 🎮 **Server Control**
```bash
# Server lifecycle management
craftops server start    # Start the server
craftops server stop     # Graceful shutdown
craftops server restart  # Stop, then start with player warnings
craftops server status   # Check current status

# Advanced server management
craftops --debug server start     # Debug mode
craftops --dry-run server restart # Preview restart process
```

### 💾 **Backup Operations**
```bash
# Backup management
craftops backup create   # Create manual backup
craftops backup list     # Show all available backups

# Automated backups happen before mod updates
craftops update-mods     # Automatically creates backup first
```

### 🏥 **System Monitoring**
```bash
# Health and diagnostics
craftops health-check              # Full system validation
craftops --debug health-check      # Detailed diagnostic output
craftops --config /custom/path.toml health-check  # Custom config
```

---

## 📖 Documentation

| Document | Description |
|----------|-------------|
| **[📚 Usage Guide](USAGE_GUIDE.md)** | Comprehensive user manual with examples and troubleshooting |
| **[🚀 Deployment Guide](DEPLOYMENT_GUIDE.md)** | Release process, CI/CD, and distribution strategy |
| **[🏗️ Project Structure](PROJECT_STRUCTURE.md)** | Codebase organization and architecture details |

---

## 🔧 Configuration Reference

<details>
<summary><b>📋 Complete Configuration Options</b></summary>

```toml
# Global settings
debug = false
dry_run = false

[minecraft]
version = "1.20.1"          # Target Minecraft version
modloader = "fabric"        # Mod loader (fabric, forge, quilt, neoforge)

[paths]
server = "/path/to/server"   # Minecraft server directory
mods = "/path/to/mods"       # Mods directory
backups = "/path/to/backups" # Backup storage
logs = "/path/to/logs"       # Log directory

[server]
jar_name = "server.jar"      # Server JAR filename
java_flags = ["-Xms4G", "-Xmx4G", "-XX:+UseG1GC"]  # JVM arguments
stop_command = "stop"        # Server stop command
max_stop_wait = 300         # Max seconds to wait for stop
startup_timeout = 120       # Max seconds to wait for start

[mods]
auto_update = true          # Enable automatic updates
backup_before_update = true # Create backup before updating
concurrent_downloads = 5    # Parallel download limit
max_retries = 3            # Retry attempts for failed downloads
retry_delay = 2.0          # Delay between retries (seconds)
timeout = 30               # HTTP request timeout

[mods.sources]
modrinth = [               # Modrinth mod URLs
    "https://modrinth.com/mod/fabric-api",
    "https://modrinth.com/mod/sodium"
]

[backup]
enabled = true             # Enable backup system
max_backups = 5           # Number of backups to keep
compression_level = 6     # Compression level (1-9)
include_logs = false      # Include server logs in backup
exclude_patterns = [      # Files/patterns to exclude
    "*.log", "cache/", "temp/"
]

[notifications]
discord_webhook = ""       # Discord webhook URL
warning_intervals = [15, 10, 5, 1]  # Warning times (minutes)
warning_message = "Server will restart in {minutes} minute(s)"
success_notifications = true
error_notifications = true

[logging]
level = "INFO"            # Log level (DEBUG, INFO, WARNING, ERROR)
format = "json"           # Log format (json, text)
file_enabled = true       # Enable file logging
console_enabled = true    # Enable console logging
max_file_size = "10MB"    # Max log file size
backup_count = 5          # Number of log files to keep
```

</details>

---

## 🌟 Platform Support

| Platform | Architecture | Status | Notes |
|----------|-------------|--------|-------|
| **Linux** | x64 | ✅ Full Support | Primary platform |
| **Linux** | ARM64 | ✅ Full Support | Raspberry Pi, ARM servers |
| **macOS** | x64 | ✅ Full Support | Intel Macs |
| **macOS** | ARM64 | ✅ Full Support | Apple Silicon (M1/M2) |

> **Note**: Server management relies on Unix-specific tools like `screen`. For Windows users, consider using WSL2 or Docker.

---

## 🔮 Roadmap

### 🎯 **Version 2.1.0** (Planned)
- **CurseForge Integration**: Full API support for CurseForge mods
- **GitHub Releases**: Support for GitHub-hosted mod releases
- **Web Interface**: Optional web UI for server management
- **Plugin System**: Extensible architecture for custom integrations

### 🚀 **Version 2.2.0** (Future)
- **Multi-Server Support**: Manage multiple Minecraft servers
- **Scheduled Updates**: Cron-like scheduling for automated updates
- **Metrics & Monitoring**: Prometheus metrics and health endpoints
- **Configuration Profiles**: Multiple configuration sets

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### 🐛 **Report Issues**
- [Create an issue](https://github.com/dacrab/craftops/issues) for bugs or feature requests
- Use the issue templates for better organization
- Include system information and logs when reporting bugs

### 💻 **Development**
```bash
# Set up development environment
git clone https://github.com/dacrab/craftops.git
cd craftops
make dev

# Run tests
make test

# Build and test
make build
./build/craftops --help
```

### 📝 **Documentation**
- Improve documentation and examples
- Add translations for international users
- Create video tutorials and guides

---

## 📊 Project Stats

<div align="center">

![GitHub stars](https://img.shields.io/github/stars/dacrab/craftops?style=social)
![GitHub forks](https://img.shields.io/github/forks/dacrab/craftops?style=social)
![GitHub issues](https://img.shields.io/github/issues/dacrab/craftops)
![GitHub pull requests](https://img.shields.io/github/issues-pr/dacrab/craftops)

</div>

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for the Minecraft cocraftopsnity**

[⭐ Star this project](https://github.com/dacrab/craftops) • [🐛 Report Issues](https://github.com/dacrab/craftops/issues) • [💬 Discussions](https://github.com/dacrab/craftops/discussions)

</div>