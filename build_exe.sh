#!/bin/bash

# Exit on error
set -e

# Clean up previous builds
rm -rf build dist

# Install build dependencies
pip install pyinstaller

# Create executable
pyinstaller --onefile \
  --name minecraft-mod-manager \
  --add-data "minecraft_mod_manager/config/config.jsonc:minecraft_mod_manager/config" \
  minecraft_mod_manager/__main__.py

# Make the executable executable
chmod +x dist/minecraft-mod-manager

# Print success message
echo "Build complete! The executable is located at dist/minecraft-mod-manager"
echo
echo "To use the executable:"
echo "1. Copy dist/minecraft-mod-manager to your desired location"
echo "2. Create config file at ~/.config/minecraft-mod-manager/config.jsonc"
echo "3. Run ./minecraft-mod-manager --help for usage information" 