#!/bin/bash

# Minecraft Mod Manager - Package Build Script
# Builds the Python package for distribution

set -e  # Exit on error

echo "🏗️  Minecraft Mod Manager - Package Build"
echo "========================================="

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

# Clean previous builds
echo "🧹 Cleaning previous build artifacts..."
rm -rf build/ dist/ *.egg-info/

# Install build dependencies
echo "📦 Installing build dependencies..."
pip install --upgrade pip build twine

# Build the package
echo "🔨 Building package..."
python -m build

# Verify the build
echo "🧪 Verifying build..."
if [ -f "dist/"*.whl ] && [ -f "dist/"*.tar.gz ]; then
    echo "✅ Package built successfully!"
    
    # List built files
    echo "📦 Built packages:"
    ls -la dist/
    
    # Test installation in a temporary environment
    echo "🧪 Testing package installation..."
    python -m pip install --force-reinstall dist/*.whl
    
    if minecraft-mod-manager --help >/dev/null 2>&1; then
        echo "✅ Package installation test passed"
    else
        echo "⚠️  Package installation test failed"
    fi
else
    echo "❌ Build failed - packages not found"
    exit 1
fi

echo ""
echo "🎉 Build complete!"
echo ""
echo "📁 Package location: dist/"
echo ""
echo "🚀 Next steps:"
echo "  • Test locally: pip install dist/*.whl"
echo "  • Upload to PyPI: twine upload dist/*"
echo "  • Create GitHub release with assets"
echo ""