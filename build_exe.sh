#!/bin/bash

# Clean previous builds
rm -rf build/ dist/ *.egg-info/

# Create virtual environment
python -m venv venv
source venv/bin/activate

# Install dependencies
pip install --upgrade pip
pip install -r requirements.txt
pip install pyinstaller

# Build executable
pyinstaller --onefile \
    --name minecraft-mod-manager \
    --add-data "minecraft_mod_manager/config/config.jsonc.example:minecraft_mod_manager/config" \
    minecraft_mod_manager/__main__.py

echo "Build complete! Executable is in dist/minecraft-mod-manager"
echo ""
echo "To distribute to users:"
echo "1. Copy dist/minecraft-mod-manager to the desired location"
echo "2. Create config file at ~/.config/minecraft-mod-manager/config.jsonc"
echo "3. Run the executable: ./minecraft-mod-manager --help" 