#!/bin/bash

# Exit on error
set -e

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install --upgrade pip
pip install -e .

# Create config directory if it doesn't exist
if [ ! -d ~/.config/minecraft-mod-manager ]; then
    echo "Creating config directory..."
    mkdir -p ~/.config/minecraft-mod-manager
fi

# Copy example config if it doesn't exist
if [ ! -f ~/.config/minecraft-mod-manager/config.toml ]; then
    echo "Copying example config..."
    cp minecraft_mod_manager/config/config.toml ~/.config/minecraft-mod-manager/config.toml
fi

echo "Build complete!"
echo "The example config has been copied to ~/.config/minecraft-mod-manager/config.toml"
echo "Edit ~/.config/minecraft-mod-manager/config.toml with your server paths before testing." 