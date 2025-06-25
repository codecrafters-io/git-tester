#!/bin/sh

# Detect the current platform and architecture
SCRIPT_DIR=$(dirname $0)
OS=$(uname -s)
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

# Map OS names and select binary
case "$OS" in
    Darwin)
        BINARY="mygit-darwin-$ARCH"
        ;;
    Linux)
        BINARY="mygit-linux-$ARCH"
        ;;
    CYGWIN*|MINGW*|MSYS*)
        BINARY="mygit-windows-amd64.exe"
        ;;
    *)
        echo "Unsupported operating system: $OS" >&2
        exit 1
        ;;
esac

# Check if the binary exists
BINARY_PATH="$SCRIPT_DIR/$BINARY"
if [ ! -f "$BINARY_PATH" ]; then
    echo "Binary not found: $BINARY_PATH" >&2
    echo "Available binaries:" >&2
    ls -1 "$SCRIPT_DIR"/mygit-* 2>/dev/null >&2 || echo "No mygit binaries found" >&2
    exit 1
fi

# Make sure the binary is executable
chmod +x "$BINARY_PATH"

# Execute the appropriate binary
exec "$BINARY_PATH" "$@"
