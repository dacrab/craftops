# Project Structure

## Package Organization
The project follows Go's standard project layout with clear separation of concerns:

```
craftops/
├── cmd/craftops/               # Application entry points
│   └── main.go                # Main application entry point
├── internal/                   # Private application code
│   ├── cli/                   # CLI command implementations
│   │   ├── root.go           # Root command and global setup
│   │   ├── init.go           # Configuration initialization
│   │   ├── health.go         # Health check command
│   │   ├── mods.go           # Mod management commands
│   │   ├── server.go         # Server management commands
│   │   └── backup.go         # Backup management commands
│   ├── config/                # Configuration handling
│   │   └── config.go         # Configuration structs and loading logic
│   └── services/              # Core business logic services
│       ├── mod_service.go    # Mod downloading and updating logic
│       ├── server_service.go # Minecraft server process control
│       ├── backup_service.go # Backup creation and management
│       └── notification_service.go # Discord webhooks and notifications
├── build/                     # Build output directory
├── dist/                      # Distribution packages
├── config.toml                  # Default configuration file
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── Makefile                   # Build automation
└── Dockerfile                 # Container build definition
```

## Architecture Patterns

### Service-Oriented Architecture
- Each service handles a specific domain (mods, server, backup, notifications)
- Services receive config and logger via constructor injection
- Clean separation between CLI layer and business logic

### Configuration-Driven Design
- All behavior controlled via TOML configuration files
- Go structs with TOML tags for type-safe configuration access
- Default config with validation and multiple file locations

### Concurrent Design
- HTTP requests use goroutines for concurrent mod downloads
- Context-based cancellation and timeout handling
- Semaphore pattern for controlling concurrency limits

## File Naming Conventions
- **snake_case** for all Go files and directories
- **lowercase** package names following Go conventions
- **PascalCase** for exported types and functions
- **camelCase** for unexported types and functions
- **UPPER_CASE** for constants

## Import Patterns
- Use relative imports within the module (`craftops/internal/config`)
- Group imports: standard library, third-party, local packages
- Use blank imports only when necessary (e.g., database drivers)

## Testing Structure
```
*_test.go files alongside source files (Go convention)
```

## Build Artifacts
- `build/` - Local build output
- `dist/` - Multi-platform distribution binaries
- `coverage.out` - Test coverage reports
- `coverage.html` - HTML coverage reports