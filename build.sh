#!/bin/bash

# Exit on error
set -e

# Clean up previous builds
rm -rf build dist *.egg-info

# Create virtual environment
python -m venv venv
source venv/bin/activate

# Install dependencies
pip install --upgrade pip
pip install -r requirements.txt
pip install -e .

# Create config directory if it doesn't exist
mkdir -p ~/.config/minecraft-mod-manager

# Copy config file if it doesn't exist
if [ ! -f ~/.config/minecraft-mod-manager/config.jsonc ]; then
    cp minecraft_mod_manager/config/config.jsonc ~/.config/minecraft-mod-manager/config.jsonc
fi

# Print success message
echo "Build complete! The package has been installed in development mode."
echo
echo "The example config has been copied to ~/.config/minecraft-mod-manager/config.jsonc"
echo "Edit ~/.config/minecraft-mod-manager/config.jsonc with your server paths before testing."
echo
echo "Run 'minecraft-mod-manager --help' for usage information." 