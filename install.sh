#!/bin/bash

# Minecraft Mod Manager - Installation Script
# This script downloads and installs the latest release

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="dacrab/craftops"
BINARY_NAME="craftops"
INSTALL_DIR="/usr/local/bin"

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
echo "🎮 CraftOps - Installation Script"
echo "=============================================="
echo -e "${NC}"

# Check if running as root for system-wide installation
if [ "$EUID" -eq 0 ]; then
    log_warning "Running as root - installing system-wide"
    INSTALL_DIR="/usr/local/bin"
else
    log_info "Running as user - installing to ~/.local/bin"
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> ~/.bashrc
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> ~/.zshrc 2>/dev/null || true
        log_info "Added $INSTALL_DIR to PATH in shell configuration"
    fi
fi

# Detect platform and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $OS in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    *)
        log_error "Unsupported operating system: $OS"
        exit 1
        ;;
esac

case $ARCH in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        log_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

PLATFORM="${OS}-${ARCH}"
log_info "Detected platform: $PLATFORM"

# Get latest release version
log_info "Fetching latest release information..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    log_error "Failed to get latest release version"
    exit 1
fi

log_info "Latest version: $VERSION"

# Construct download URL
BINARY_FILE="${BINARY_NAME}-${PLATFORM}"
if [ "$OS" = "windows" ]; then
    BINARY_FILE="${BINARY_FILE}.exe"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_FILE"
log_info "Download URL: $DOWNLOAD_URL"

# Download binary
log_info "Downloading $BINARY_FILE..."
TEMP_FILE=$(mktemp)

if command -v curl >/dev/null 2>&1; then
    curl -L -o "$TEMP_FILE" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$TEMP_FILE" "$DOWNLOAD_URL"
else
    log_error "Neither curl nor wget is available. Please install one of them."
    exit 1
fi

# Verify download
if [ ! -s "$TEMP_FILE" ]; then
    log_error "Download failed or file is empty"
    rm -f "$TEMP_FILE"
    exit 1
fi

# Install binary
log_info "Installing to $INSTALL_DIR/$BINARY_NAME..."
chmod +x "$TEMP_FILE"

if [ "$EUID" -eq 0 ] || [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
fi

log_success "Binary installed successfully"

# Create aliases
log_info "Creating command aliases..."

create_alias() {
    local alias_name="$1"
    local target_path="$INSTALL_DIR/$alias_name"
    
    if [ -e "$target_path" ]; then
        log_warning "Alias $alias_name already exists, skipping"
        return
    fi
    
    if [ "$EUID" -eq 0 ] || [ -w "$INSTALL_DIR" ]; then
        ln -sf "$INSTALL_DIR/$BINARY_NAME" "$target_path"
    else
        sudo ln -sf "$INSTALL_DIR/$BINARY_NAME" "$target_path"
    fi
    
    log_success "Created alias: $alias_name"
}

create_alias "cops"
create_alias "mmu"

# Verify installation
log_info "Verifying installation..."
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    VERSION_OUTPUT=$("$BINARY_NAME" --version 2>/dev/null || echo "Version check failed")
    log_success "Installation verified: $VERSION_OUTPUT"
else
    log_warning "Binary not found in PATH. You may need to restart your shell or run:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi

# Test aliases
for alias_cmd in "cops" "mmu"; do
    if command -v "$alias_cmd" >/dev/null 2>&1; then
        log_success "Alias '$alias_cmd' is working"
    else
        log_warning "Alias '$alias_cmd' not found in PATH"
    fi
done

# Create default configuration
log_info "Setting up configuration..."
CONFIG_DIR="$HOME/.config/craftops"
CONFIG_FILE="$CONFIG_DIR/config.toml"

if [ ! -d "$CONFIG_DIR" ]; then
    mkdir -p "$CONFIG_DIR"
    log_success "Created config directory: $CONFIG_DIR"
fi

if [ ! -f "$CONFIG_FILE" ]; then
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        "$BINARY_NAME" init-config -o "$CONFIG_FILE" >/dev/null 2>&1 || true
        if [ -f "$CONFIG_FILE" ]; then
            log_success "Created default configuration: $CONFIG_FILE"
        else
            log_warning "Failed to create default configuration"
        fi
    else
        log_warning "Cannot create default config - binary not in PATH"
    fi
else
    log_success "Configuration file already exists: $CONFIG_FILE"
fi

# Final instructions
echo ""
echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
echo ""
echo "📋 Available commands:"
echo "  • craftops               (full name)"
echo "  • cops                   (short alias)"
echo "  • mmu                    (legacy alias)"
echo ""
echo "🚀 Quick start:"
echo "  1. Edit your configuration:"
echo "     nano $CONFIG_FILE"
echo ""
echo "  2. Run health check:"
echo "     cops health-check"
echo ""
echo "  3. Update mods:"
echo "     cops update-mods"
echo ""
echo "  4. Server management:"
echo "     cops server start"
echo "     cops server stop"
echo "     cops server restart"
echo "     cops server status"
echo ""
echo "  5. Backup management:"
echo "     cops backup create"
echo "     cops backup list"
echo ""
echo "💡 For help:"
echo "  cops --help"
echo "  cops [command] --help"
echo ""

if [ "$EUID" -ne 0 ] && [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  Note: You may need to restart your shell or run:${NC}"
    echo "  source ~/.bashrc"
    echo ""
fi

echo -e "${BLUE}📖 Documentation: https://github.com/$REPO${NC}"
echo -e "${BLUE}🐛 Issues: https://github.com/$REPO/issues${NC}"