#!/bin/sh

# Check if git is working in PATH first
if command -v git >/dev/null 2>&1; then
    if [ "$1" = "write-tree" ]; then
        git add .
    fi
    exec git "$@"
fi

# Find git binary in /tmp locations
for tmpdir in /tmp/git-*/git; do
    if [ -x "$tmpdir" ]; then
        # commit-tree stage doesn't use call this script for init
        # So we need to run this setup again
        if [ "$1" = "commit-tree" ]; then
            "$tmpdir" config --local user.email "hello@codecrafters.io"
            "$tmpdir" config --local user.name "CodeCrafters-Bot"
        fi

        if [ "$1" = "write-tree" ]; then
            "$tmpdir" add .
        fi

        if [ "$1" = "init" ]; then
            "$tmpdir" "$@"
            # If init.defaultBranch is not set, set it to main
            # If not set globally, git always shows a warning
            if ! "$tmpdir" config --global --get init.defaultBranch >/dev/null 2>&1; then
                "$tmpdir" config --global init.defaultBranch main
            fi
            # Setup is run locally so it's only set for the current temp repo
            "$tmpdir" config --local user.email "hello@codecrafters.io"
            "$tmpdir" config --local user.name "CodeCrafters-Bot"
        else
            exec "$tmpdir" "$@"
        fi
    fi
done

echo "git binary not found in PATH or /tmp directories"
exit 1
