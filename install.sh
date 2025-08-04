#!/bin/bash

# Installation script for Zerops MCP Server

set -e

REPO="zeropsio/zerops-mcp-v3"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest release
echo "🔍 Finding latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=${LATEST_RELEASE#v}

if [ -z "$VERSION" ]; then
    echo "❌ Could not determine latest version"
    exit 1
fi

echo "📦 Installing Zerops MCP Server v$VERSION for $OS/$ARCH..."

# Download URL
FILENAME="mcp-server-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$FILENAME"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

echo "⬇️  Downloading..."
curl -sL "$URL" -o "$TEMP_DIR/$FILENAME"

echo "📂 Extracting..."
tar -xzf "$TEMP_DIR/$FILENAME" -C "$TEMP_DIR"

echo "🚀 Installing..."
BINARY_NAME="mcp-server-${VERSION}-${OS}-${ARCH}"
mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/mcp-server"
chmod +x "$INSTALL_DIR/mcp-server"

echo "✅ Installation complete!"
echo ""
echo "The mcp-server binary has been installed to: $INSTALL_DIR/mcp-server"
echo ""

# Check if install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠️  Warning: $INSTALL_DIR is not in your PATH"
    echo ""
    echo "Add it to your PATH by running:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
    echo "To make it permanent, add the above line to your shell profile:"
    echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
fi

echo "Next steps:"
echo "1. Set your Zerops API key:"
echo "   export ZEROPS_API_KEY='your-api-key'"
echo ""
echo "2. Run the server:"
echo "   mcp-server"
echo ""
echo "For more information, see: https://github.com/$REPO"