#!/bin/bash

# Minecraft Mod Manager - Executable Build Script
# Creates a standalone executable using PyInstaller

set -e  # Exit on error

echo "🏗️  Minecraft Mod Manager - Executable Build"
echo "============================================="

# Check if we're in the project root
if [ ! -f "pyproject.toml" ]; then
    echo "❌ Error: This script must be run from the project root directory"
    exit 1
fi

# Check if we're in a virtual environment
if [ -z "$VIRTUAL_ENV" ]; then
    echo "⚠️  Warning: Not in a virtual environment. Activating venv..."
    if [ -d "venv" ]; then
        source venv/bin/activate
        echo "✅ Activated virtual environment"
    else
        echo "❌ Error: No virtual environment found. Run scripts/development-setup.sh first."
        exit 1
    fi
fi

# Install PyInstaller if not already installed
echo "📦 Installing PyInstaller..."
pip install pyinstaller

# Clean previous builds
echo "🧹 Cleaning previous build artifacts..."
rm -rf dist/ build/ *.spec

# Create PyInstaller spec file
echo "📝 Creating PyInstaller specification..."
cat > minecraft-mod-manager.spec << 'EOF'
# -*- mode: python ; coding: utf-8 -*-

a = Analysis(
    ['minecraft_mod_manager/app.py'],
    pathex=[],
    binaries=[],
    datas=[
        ('minecraft_mod_manager/settings/config.toml', 'minecraft_mod_manager/settings'),
    ],
    hiddenimports=[
        'aiohttp',
        'tqdm',
        'toml',
        'asyncio',
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)

pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name='minecraft-mod-manager',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)
EOF

# Build executable
echo "🔨 Building executable..."
pyinstaller minecraft-mod-manager.spec

# Verify the build
if [ -f "dist/minecraft-mod-manager" ]; then
    echo "✅ Executable built successfully!"
    
    # Get file size
    size=$(du -h dist/minecraft-mod-manager | cut -f1)
    echo "📊 Executable size: $size"
    
    # Test the executable
    echo "🧪 Testing executable..."
    if ./dist/minecraft-mod-manager --help >/dev/null 2>&1; then
        echo "✅ Executable test passed"
    else
        echo "⚠️  Executable test failed"
    fi
else
    echo "❌ Build failed - executable not found"
    exit 1
fi

# Clean up spec file
rm -f minecraft-mod-manager.spec

echo ""
echo "🎉 Build complete!"
echo ""
echo "📁 Executable location: dist/minecraft-mod-manager"
echo ""
echo "🚀 Installation instructions:"
echo "  1. Copy executable: sudo cp dist/minecraft-mod-manager /usr/local/bin/"
echo "  2. Make executable: sudo chmod +x /usr/local/bin/minecraft-mod-manager"
echo "  3. Create config: mkdir -p ~/.config/minecraft-mod-manager"
echo "  4. Copy config: cp minecraft_mod_manager/settings/config.toml ~/.config/minecraft-mod-manager/"
echo "  5. Edit config: nano ~/.config/minecraft-mod-manager/config.toml"
echo "  6. Test: minecraft-mod-manager --help"
echo ""