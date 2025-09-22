#!/bin/bash

# Gophish-MJSET Integration Installation Script

echo "Installing Gophish-MJSET Report Integration..."

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is required but not installed."
    exit 1
fi

# Check if pip3 is available
if ! command -v pip3 &> /dev/null; then
    echo "Error: pip3 is required but not installed."
    exit 1
fi

# Install Python dependencies
echo "Installing Python dependencies..."
pip3 install -r requirements.txt

if [ $? -ne 0 ]; then
    echo "Error: Failed to install Python dependencies."
    exit 1
fi

# Make scripts executable
chmod +x gophish_bridge.py
chmod +x test_integration.py

# Create output directory
mkdir -p output

# Create assets directory structure
mkdir -p assets/{fonts,images,templates}

echo "Installation completed successfully!"
echo ""
echo "Next steps:"
echo "1. Copy fonts, images, and templates from your MJSET installation to assets/ (optional)"
echo "2. Set your Gophish API key: export GOPHISH_API_KEY='your_key_here'"
echo "3. Test the integration: python3 test_integration.py"
echo "4. Rebuild Gophish: go build"
echo ""
echo "API Endpoints added:"
echo "- GET /api/campaigns/{id}/export"
echo "- POST /api/campaigns/{id}/reports/generate" 
echo "- GET /api/campaigns/{id}/reports/download/{type}"
