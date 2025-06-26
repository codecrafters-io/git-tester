#!/bin/sh

# Look for git binary in temp directories first (since it might be moved)
GIT_PATH=""

# Check for git in /tmp/git-* directories
for tmpdir in /tmp/git-*/git; do
    if [ -x "$tmpdir" ]; then
        GIT_PATH="$tmpdir"
        break
    fi
done

# Check for git in /tmp/copied-git/
if [ -z "$GIT_PATH" ] && [ -x "/tmp/copied-git/git" ]; then
    GIT_PATH="/tmp/copied-git/git"
fi

# Check if git is still in PATH
if [ -z "$GIT_PATH" ] && command -v git >/dev/null 2>&1; then
    GIT_PATH=$(which git)
fi

# Exit if no git found
if [ -z "$GIT_PATH" ]; then
    echo "git binary not found in PATH or temp directories"
    exit 1
fi

# Configure git if it's in a temp directory
if echo "$GIT_PATH" | grep -q "/tmp/"; then
    "$GIT_PATH" config --global init.defaultBranch main
    "$GIT_PATH" config --global user.email "hello@codecrafters.io"
    "$GIT_PATH" config --global user.name "CodeCrafters-Bot"
fi

# Handle write-tree special case
if [ "$1" = "write-tree" ]; then
    "$GIT_PATH" add .
fi

exec "$GIT_PATH" "$@"
