#!/bin/bash

# Minecraft Mod Manager - Development Environment Setup
# This script provides a comprehensive development environment setup

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Header
echo -e "${BLUE}"
echo "🚀 Minecraft Mod Manager - Development Setup"
echo "============================================="
echo -e "${NC}"

# Check if we're in the project root
if [ ! -f "pyproject.toml" ]; then
    log_error "This script must be run from the project root directory"
    exit 1
fi

# Check Python version
log_info "Checking Python version..."
if command -v python3 >/dev/null 2>&1; then
    python_version=$(python3 --version 2>&1 | cut -d' ' -f2 | cut -d'.' -f1,2)
    required_version="3.9"
    
    if [ "$(printf '%s\n' "$required_version" "$python_version" | sort -V | head -n1)" != "$required_version" ]; then
        log_error "Python $required_version or higher is required. Found: $python_version"
        exit 1
    fi
    
    log_success "Python version: $python_version"
else
    log_error "Python 3 is not installed or not in PATH"
    exit 1
fi

# Create virtual environment
log_info "Setting up virtual environment..."
if [ ! -d "venv" ]; then
    python3 -m venv venv
    log_success "Created virtual environment"
else
    log_success "Virtual environment already exists"
fi

# Activate virtual environment
log_info "Activating virtual environment..."
source venv/bin/activate

# Upgrade pip and install build tools
log_info "Upgrading pip and installing build tools..."
pip install --upgrade pip setuptools wheel build

# Install package in development mode
log_info "Installing package in development mode..."
pip install -e .

# Install development dependencies
if [ -f "requirements-dev.txt" ]; then
    log_info "Installing development dependencies..."
    pip install -r requirements-dev.txt
    log_success "Development dependencies installed"
else
    log_warning "No requirements-dev.txt found, skipping dev dependencies"
fi

# Create config directory
config_dir="$HOME/.config/minecraft-mod-manager"
log_info "Setting up configuration..."
if [ ! -d "$config_dir" ]; then
    mkdir -p "$config_dir"
    log_success "Created config directory: $config_dir"
else
    log_success "Config directory already exists: $config_dir"
fi

# Copy example config
config_file="$config_dir/config.toml"
if [ ! -f "$config_file" ]; then
    cp minecraft_mod_manager/settings/config.toml "$config_file"
    log_success "Copied example config to: $config_file"
    log_warning "Please edit $config_file with your server paths before using"
else
    log_success "Config file already exists: $config_file"
fi

# Run quality checks
log_info "Running quality checks..."

# Check if command is available
if command -v minecraft-mod-manager >/dev/null 2>&1; then
    log_success "minecraft-mod-manager command is available"
    
    # Test help command
    if minecraft-mod-manager --help >/dev/null 2>&1; then
        log_success "Help command works"
    else
        log_warning "Help command failed"
    fi
else
    log_warning "minecraft-mod-manager command not found in PATH"
fi

# Run linting if available
if command -v ruff >/dev/null 2>&1; then
    log_info "Running code linting..."
    if python -m ruff check minecraft_mod_manager/ >/dev/null 2>&1; then
        log_success "Code linting passed"
    else
        log_warning "Code linting found issues"
    fi
fi

# Run type checking if available
if command -v mypy >/dev/null 2>&1; then
    log_info "Running type checking..."
    if python -m mypy minecraft_mod_manager/ >/dev/null 2>&1; then
        log_success "Type checking passed"
    else
        log_warning "Type checking found issues"
    fi
fi

# Run tests if available
if command -v pytest >/dev/null 2>&1; then
    log_info "Running tests..."
    if python -m pytest tests/ -q >/dev/null 2>&1; then
        log_success "All tests passed"
    else
        log_warning "Some tests failed"
    fi
fi

# Final summary
echo ""
echo -e "${GREEN}🎉 Development setup complete!${NC}"
echo ""
echo "📋 Summary:"
echo "  • Virtual environment: venv/"
echo "  • Config file: $config_file"
echo "  • Package installed in development mode"
echo "  • Development dependencies installed"
echo ""
echo "🚀 Next steps:"
echo "  1. Activate environment: source venv/bin/activate"
echo "  2. Edit your config: $config_file"
echo "  3. Run health check: minecraft-mod-manager --health-check"
echo "  4. Run tests: make test"
echo "  5. Check code quality: make check-all"
echo ""
echo "💡 Available make commands:"
echo "  • make help          - Show all available commands"
echo "  • make test          - Run tests"
echo "  • make lint          - Run linting"
echo "  • make type-check    - Run type checking"
echo "  • make build         - Build package"
echo "  • make clean         - Clean build artifacts"
echo ""