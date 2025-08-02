# Product Overview

**CraftOps** is a modern, comprehensive CLI tool for Minecraft server operations and automated mod management. Built in Go for performance and reliability, it automates mod updates from Modrinth while ensuring server stability through automated backups and Discord notifications.

## Core Features
- **Server Lifecycle Management**: Start, stop, restart servers with status monitoring via screen sessions
- **Automated Mod Updates**: Full Modrinth API integration with concurrent downloads
- **Intelligent Backup System**: Compressed backups with retention policies and selective exclusion
- **Smart Notifications**: Discord webhook integration with configurable restart warnings
- **Health Monitoring**: Comprehensive system validation and diagnostics
- **Configuration-Driven**: TOML-based configuration with sensible defaults and validation

## Current Status (v2.0.0)
- ✅ **Modrinth Integration**: Fully implemented and tested
- ✅ **Linux/macOS Support**: Complete server management via screen sessions

## Target Users
System administrators and server operators who need reliable, automated mod management for Minecraft servers on Linux/macOS platforms.

## Key Value Propositions
- **Performance**: Go-based with concurrent downloads and minimal resource usage
- **Reliability**: Comprehensive error handling and automatic backups
- **User Experience**: Beautiful CLI with colored output, progress bars, and clear feedback
- **Automation**: Reduces manual mod management overhead significantly
- **Safety**: Prevents server downtime through automated backups and player warnings

## Codebase Guidelines

### Code Quality Standards
- **Clean Architecture**: Follow service-oriented design with clear separation of concerns
- **Go Best Practices**: Adhere to effective Go patterns and idioms
- **Error Handling**: Comprehensive error handling with context and proper logging
- **Documentation**: Clear code comments and comprehensive README/guides
- **Testing**: Write focused tests for critical business logic (avoid over-testing)

### Development Workflow
- **Plan First**: Always understand the codebase and create a plan before implementing
- **Search & Understand**: Use grep/find to understand existing patterns before adding new code
- **Keep It Simple**: Avoid overcomplicating and overengineering (KISS principle)
- **User Experience**: Think about the end-user experience and CLI usability
- **Clean Up**: Remove dead code, unused imports, and temporary files
- **Read Carefully**: Take time to thoroughly read terminal output, files, and prompts - don't rush to results or be sloppy

### Code Style
- **Go Formatting**: Use `go fmt` and `goimports` for consistent formatting
- **Naming**: Follow Go naming conventions (PascalCase for exports, camelCase for private)
- **Structure**: Organize code logically with clear package boundaries
- **Dependencies**: Minimize external dependencies, prefer standard library when possible
- **Logging**: Use structured logging (zap) with appropriate log levels

### Performance & Reliability
- **Concurrency**: Use goroutines and channels appropriately for concurrent operations
- **Context**: Always use context.Context for cancellation and timeouts
- **Resource Management**: Properly close files, HTTP connections, and other resources
- **Error Recovery**: Implement retry logic and graceful degradation where appropriate
- **Memory Efficiency**: Avoid memory leaks and unnecessary allocations

### User Interface
- **Colored Output**: Use consistent color scheme for success/error/warning/info messages
- **Progress Feedback**: Show progress bars for long-running operations
- **Clear Messages**: Provide actionable error messages and helpful guidance
- **Dry Run Support**: Support `--dry-run` flag for preview operations
- **Verbose Logging**: Support `--debug` flag for detailed troubleshooting