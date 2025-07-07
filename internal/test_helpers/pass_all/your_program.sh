#!/bin/sh

# Check if git from PATH is working first
if command -v git >/dev/null 2>&1; then
    if [ "$1" = "write-tree" ]; then
        git add .
    fi
    exec git "$@"
fi

# Find git binary in /tmp locations
for tmpdir in /tmp/git-*/git; do
    if [ -x "$tmpdir" ]; then
        # If defaultBranch config is not set, we set it to main (doesn't work without global config)
        if ! "$tmpdir" config --global --get init.defaultBranch >/dev/null 2>&1; then
            "$tmpdir" config --global init.defaultBranch main
        fi

        # commit-tree stage doesn't use this script for init
        # So we need to run this setup again
        if [ "$1" = "commit-tree" ]; then
            "$tmpdir" config --local user.email "hello@codecrafters.io"
            "$tmpdir" config --local user.name "CodeCrafters-Bot"
        fi

        if [ "$1" = "write-tree" ]; then
            "$tmpdir" add .
        fi

        exec "$tmpdir" "$@"
    fi
done

echo "git binary not found in PATH or /tmp directories"
exit 1
