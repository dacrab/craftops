#!/bin/bash

# Exit on error
set -e

# Install PyInstaller if not already installed
pip install pyinstaller

# Build executable
pyinstaller \
--onefile \
--name minecraft-mod-manager \
--add-data "minecraft_mod_manager/config/config.toml:minecraft_mod_manager/config" \
minecraft_mod_manager/__main__.py

# Print instructions
echo "Build complete! The executable has been created in the dist directory."
echo
echo "To use the executable:"
echo "1. Copy dist/minecraft-mod-manager to a location in your PATH"
echo "2. Create config file at ~/.config/minecraft-mod-manager/config.toml"
echo
echo "Run 'minecraft-mod-manager --help' for usage information." 