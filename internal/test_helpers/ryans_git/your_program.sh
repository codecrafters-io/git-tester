#!/bin/sh

# Detect the current platform and architecture
SCRIPT_DIR=$(dirname $0)
OS=$(uname -s)
ARCH=$(uname -m)

# Normalize architecture and build binary name
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] && ARCH="arm64"

BINARY="mygit-$(echo $OS | tr '[:upper:]' '[:lower:]')-$ARCH"

# Use the binary if it exists, otherwise fall back to go run
BINARY_PATH="$SCRIPT_DIR/$BINARY"
if [ -f "$BINARY_PATH" ]; then
    exec "$BINARY_PATH" "$@"
else
    echo "Binary not found: $BINARY_PATH"
    exit 1
fi
