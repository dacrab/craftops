#!/bin/bash

# Clean previous builds
rm -rf build/ dist/ *.egg-info/

# Create virtual environment
python -m venv venv
source venv/bin/activate

# Install build dependencies
pip install --upgrade pip
pip install build wheel

# Build package
python -m build

# Install locally for testing
pip install -e .

# Create config directory and copy example config
mkdir -p ~/.config/minecraft-mod-manager
cp minecraft_mod_manager/config/config.jsonc.example ~/.config/minecraft-mod-manager/config.jsonc

echo "Build complete! You can now test the package."
echo "The virtual environment is activated. Use 'deactivate' when done."
echo ""
echo "The example config has been copied to ~/.config/minecraft-mod-manager/config.jsonc"
echo "Edit ~/.config/minecraft-mod-manager/config.jsonc with your server paths before testing."
echo ""
echo "Try these commands:"
echo "minecraft-mod-manager --help"
echo "minecraft-mod-manager --status" 