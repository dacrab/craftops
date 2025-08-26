<div align="center">

# 🏗️ CraftOps Project Structure

**Overview of folders, key files, and their responsibilities**

[🏠 Back to README](../README.md) • [📚 Usage Guide](USAGE_GUIDE.md) • [🚀 Deployment Guide](DEPLOYMENT_GUIDE.md)

</div>

---

## 📁 Repository Layout

```text
craftops/
  cmd/
    craftops/
      main.go                 # CLI entrypoint
  docs/
    DEPLOYMENT_GUIDE.md
    USAGE_GUIDE.md
    PROJECT_STRUCTURE.md
  internal/
    cli/                      # Cobra command implementations (user-facing CLI)
      backup.go
      health.go
      init.go
      mods.go
      root.go
      server.go
    config/
      config.go               # Configuration loading and validation
    services/                 # Core domain logic used by CLI commands
      backup_service.go
      mod_service.go
      notification_service.go
      server_service.go
  tests/
    config/
      config_test.go
    services/
      backup_service_test.go
      mod_service_test.go
      notification_service_test.go
      server_service_test.go
  Dockerfile
  install.sh
  Makefile
  go.mod
  go.sum
  LICENSE
  README.md
```

---

## 🔌 Entrypoints

- `cmd/craftops/main.go`
  - Boots the CLI application and registers commands.

---

## 🧩 CLI Commands (`internal/cli`)

- `root.go`: Base CLI initialization (flags, version, help).
- `init.go`: Generates a starter `config.toml`.
- `health.go`: Runs environment and configuration checks.
- `mods.go`: Mod update workflow (dry-run, retries, concurrency).
- `backup.go`: Manual backup actions (create/list) and retention.
- `server.go`: Server lifecycle (start/stop/restart/status).

Each command delegates to the corresponding service for business logic.

---

## 🧠 Services (`internal/services`)

- `backup_service.go`: Create compressed backups, enforce retention, exclusions.
- `mod_service.go`: Resolve and download compatible mod versions (Modrinth).
- `notification_service.go`: Discord webhook messaging and templates.
- `server_service.go`: Server process control and status.

Services are designed to be testable and independent from CLI concerns.

---

## ⚙️ Configuration (`internal/config`)

- `config.go`: Loads, parses, and validates configuration; exposes typed accessors.

---

## 🧪 Tests (`tests`) 

- Unit tests cover configuration parsing and each service module.
- Run all tests: `make test`.

---

## 🧰 Tooling & Ops

- `install.sh`: One-line installer for published releases.
- `Dockerfile`: Multi-stage image for running the CLI in a container.
- `Makefile`: Common developer tasks (build, test, dev, install).

---

## 🔗 Related Documentation

- [Usage Guide](USAGE_GUIDE.md): Commands, flags, and examples.
- [Deployment Guide](DEPLOYMENT_GUIDE.md): Release, CI/CD, and container usage.


