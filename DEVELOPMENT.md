# Development Guide

This document provides comprehensive information for developers working on the Minecraft Mod Manager project.

## 🚀 Quick Start

### Prerequisites
- Python 3.9 or higher
- Git
- Make (optional, for convenience commands)

### Setup Development Environment

```bash
# Clone the repository
git clone <repository-url>
cd minecraft-mod-manager

# Run development setup
./scripts/development-setup.sh

# Or use make (recommended)
make setup-dev
```

## 📁 Project Structure

```
minecraft-mod-manager/
├── minecraft_mod_manager/          # Main package
│   ├── app.py                      # Main application entry point
│   ├── services.py                 # Core business logic services
│   └── settings/                   # Configuration management
├── tests/                          # Test suite
├── scripts/                        # Development scripts
├── .kiro/steering/                 # AI assistant guidance
├── pyproject.toml                  # Modern Python packaging
├── Makefile                        # Development commands
└── requirements-dev.txt            # Development dependencies
```

## 🛠️ Development Commands

### Using Make (Recommended)

```bash
make help           # Show all available commands
make install-dev    # Install with dev dependencies
make test           # Run tests
make lint           # Run code linting
make type-check     # Run type checking
make format         # Format code
make build          # Build package
make build-exe      # Build standalone executable
make clean          # Clean build artifacts
make check-all      # Run all quality checks
```

### Manual Commands

```bash
# Install in development mode
pip install -e .
pip install -r requirements-dev.txt

# Run tests
python -m pytest tests/ -v

# Code quality
python -m ruff check minecraft_mod_manager/
python -m mypy minecraft_mod_manager/

# Build package
python -m build

# Build executable
./executable-build.sh
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run specific test file
python -m pytest tests/test_basic.py -v

# Run with verbose output
python -m pytest tests/ -v -s
```

### Test Structure

```
tests/
├── conftest.py         # Pytest configuration and fixtures
├── test_basic.py       # Basic functionality tests
└── test_data/          # Test data directory
```

## 🔍 Code Quality

### Linting and Formatting

We use **Ruff** for both linting and formatting:

```bash
# Check for issues
python -m ruff check minecraft_mod_manager/

# Fix auto-fixable issues
python -m ruff check --fix minecraft_mod_manager/

# Format code
python -m ruff format minecraft_mod_manager/
```

### Type Checking

We use **MyPy** for static type checking:

```bash
# Run type checking
python -m mypy minecraft_mod_manager/
```

### Code Style Guidelines

- **Line length**: 100 characters maximum
- **Type hints**: Required for all functions and methods
- **Docstrings**: Required for all modules, classes, and public methods
- **Import organization**: Use isort-compatible import ordering
- **Async patterns**: Use proper async/await patterns throughout

## 🏗️ Building and Distribution

### Building Wheel Package

```bash
# Build wheel and source distribution
make build

# Or manually
python -m build
```

### Building Standalone Executable

```bash
# Build executable with PyInstaller
make build-exe

# Or manually
./executable-build.sh
```

The executable will be created in the `dist/` directory.

## 📦 Package Management

### Dependencies

- **Production dependencies**: Defined in `pyproject.toml`
- **Development dependencies**: Listed in `requirements-dev.txt`

### Adding Dependencies

1. Add to `pyproject.toml` for production dependencies
2. Add to `requirements-dev.txt` for development dependencies
3. Update the package: `pip install -e .`

## 🔧 Configuration

### Development Configuration

The application uses TOML configuration files:

- **Template**: `minecraft_mod_manager/settings/config.toml`
- **User config**: `~/.config/minecraft-mod-manager/config.toml`

### Environment Variables

The application supports environment variable overrides for configuration values.

## 🐛 Debugging

### Logging

The application uses Python's built-in logging module:

```python
import logging
logger = logging.getLogger(__name__)
```

### Performance Monitoring

Use the built-in performance monitoring utilities:

```python
from minecraft_mod_manager.utils.performance import time_async_operation

async def my_function():
    async with time_async_operation("my_operation"):
        # Your code here
        pass
```

## 🚀 Release Process

### Version Management

1. Update version in `pyproject.toml`
2. Update version in `minecraft_mod_manager/__init__.py`
3. Create git tag: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`

### Building Release

```bash
# Clean previous builds
make clean

# Run all quality checks
make check-all

# Build package
make build

# Build executable
make build-exe
```

## 🤝 Contributing

### Code Review Checklist

- [ ] All tests pass
- [ ] Code follows style guidelines
- [ ] Type checking passes
- [ ] Documentation updated
- [ ] Performance impact considered
- [ ] Error handling implemented

### Git Workflow

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and commit
3. Run quality checks: `make check-all`
4. Push branch and create pull request

## 📚 Architecture

### Design Patterns

- **Dependency Injection**: Clean separation of concerns
- **Strategy Pattern**: Pluggable retry and validation strategies
- **Context Managers**: Resource management and cleanup
- **Protocol Interfaces**: Clear contracts between components

### Key Components

- **BaseManager**: Common functionality for all managers
- **ModManager**: Handles mod updates and version checking
- **BackupManager**: Creates and manages server backups
- **NotificationManager**: Handles Discord notifications
- **ServerController**: Controls Minecraft server process

### Utilities

- **Retry mechanisms**: Robust error handling with backoff
- **Performance monitoring**: Track operation metrics
- **Configuration validation**: Ensure valid settings
- **Health checking**: Monitor system health
- **Cleanup utilities**: Automated maintenance

## 🔒 Security Considerations

- **Input validation**: All user inputs are validated
- **Path sanitization**: Prevent directory traversal attacks
- **Permission checks**: Validate file system permissions
- **Error handling**: Don't expose sensitive information in errors

## 📈 Performance

### Optimization Guidelines

- Use async/await for I/O operations
- Implement connection pooling for HTTP requests
- Use streaming for large file operations
- Monitor performance with built-in utilities

### Profiling

Use the performance monitoring utilities to identify bottlenecks:

```python
from minecraft_mod_manager.utils.performance import performance_tracker

# After operations
performance_tracker.log_summary()
```

## 🆘 Troubleshooting

### Common Issues

1. **Import errors**: Ensure package is installed in development mode
2. **Permission errors**: Check file system permissions
3. **Network errors**: Verify internet connectivity and API access
4. **Configuration errors**: Validate configuration file syntax

### Debug Mode

Enable debug logging by setting the log level:

```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

## 📞 Support

For development questions and issues:

1. Check this documentation
2. Review the steering rules in `.kiro/steering/`
3. Check existing issues and tests
4. Create an issue with detailed information