#!/bin/bash

# Release script for Zerops MCP Server

set -e

echo "🚀 Building Zerops MCP Server for release..."

# Get version from user or use default
VERSION=${1:-"3.0.0"}
echo "Version: $VERSION"

# Create dist directory
mkdir -p dist

# Build for multiple platforms
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64" "windows/amd64")

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT="dist/mcp-server-${VERSION}-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" cmd/mcp-server/main.go
done

echo "✅ Build complete! Binaries in dist/"
echo ""
echo "📦 Creating release archives..."

cd dist
for FILE in mcp-server-*; do
    if [[ -f "$FILE" ]]; then
        ARCHIVE="${FILE}.tar.gz"
        if [[ "$FILE" == *.exe ]]; then
            ARCHIVE="${FILE%.exe}.zip"
            zip "$ARCHIVE" "$FILE"
        else
            tar -czf "$ARCHIVE" "$FILE"
        fi
        echo "Created $ARCHIVE"
    fi
done

cd ..

echo ""
echo "🎉 Release preparation complete!"
echo "Next steps:"
echo "1. Create a GitHub release with tag v$VERSION"
echo "2. Upload the archives from dist/"
echo "3. Update the installation instructions in README.md"