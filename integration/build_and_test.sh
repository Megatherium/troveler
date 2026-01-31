#!/bin/bash
# Build and run integration tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building troveler binary..."
cd "$PROJECT_DIR"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o troveler .

echo "Copying database..."
cp ~/.local/share/troveler/troveler.db "$SCRIPT_DIR/troveler.db"
cp troveler "$SCRIPT_DIR/troveler"

echo "Building Docker image..."
cd "$SCRIPT_DIR"
podman build -t troveler-integration -f Dockerfile ..

echo "Running integration tests..."
podman run --rm -v "$SCRIPT_DIR/results:/app/results:z" troveler-integration

echo ""
echo "Test results saved to $SCRIPT_DIR/results/"
