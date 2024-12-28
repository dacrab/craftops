#!/bin/bash

# Clean previous builds
rm -rf dist/ build/ *.egg-info/

# Create source distribution
python setup.py sdist bdist_wheel

echo "Build complete! Distribution files are in the dist/ directory."
echo "You can install the package directly using:"
echo "pip install dist/*.whl" 