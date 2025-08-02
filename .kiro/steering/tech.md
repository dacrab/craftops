# Technology Stack

## Language & Runtime
- **Go 1.21+** - Modern, compiled language with excellent concurrency support
- **Goroutines & Channels** - Concurrent processing for mod downloads and operations
- **Context Package** - Proper cancellation and timeout handling

## Key Dependencies
- **github.com/spf13/cobra** - Modern CLI framework with subcommands and flags
- **github.com/BurntSushi/toml** - TOML configuration file parsing
- **github.com/fatih/color** - Colored terminal output for better UX
- **github.com/schollz/progressbar/v3** - Progress bars for user feedback
- **go.uber.org/zap** - Structured, high-performance logging

## Development Tools
- **golangci-lint** - Comprehensive Go linter with multiple analyzers
- **go test** - Built-in testing framework with coverage support
- **go fmt/goimports** - Code formatting and import organization
- **go vet** - Static analysis for common Go mistakes

## Build System
- **Go modules** - Native dependency management (go.mod/go.sum)
- **Makefile** - Build automation and development workflows
- **Multi-platform builds** - Cross-compilation for Linux/macOS (x64/ARM64)
- **Docker** - Containerized deployment with multi-stage builds

## Common Commands

### Development Setup
```bash
# Clone and setup
git clone <repo>
cd craftops
make dev  # Install development dependencies

# Build locally
make build

# Run tests
make test

# Format and lint
make format
make lint
```

### Building & Installation
```bash
# Build for current platform
make build

# Install locally
make install

# Install system-wide (requires sudo)
make install-system

# Create distribution packages
make package
```

### Running
```bash
# Initialize configuration
craftops init-config

# Health check
craftops health-check

# Update mods
craftops update-mods

# Server management
craftops server start
craftops server stop
craftops server restart
```

## Configuration
- **TOML format** for all configuration files
- **Go structs with tags** for type-safe configuration handling
- **Default config** generated via `craftops init-config`
- **Multiple config locations** supported (./config.toml, ~/.config/craftops/, /etc/craftops/)