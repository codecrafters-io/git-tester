#!/bin/sh

# Check if git is accessible in PATH first
if command -v git >/dev/null 2>&1; then
    exec git "$@"
elif [ -x "/tmp/copied-git/git" ]; then
    "/tmp/copied-git/git" config --global init.defaultBranch main
    "/tmp/copied-git/git" config --global user.email "hello@codecrafters.io"
    "/tmp/copied-git/git" config --global user.name "CodeCrafters-Bot"
    exec "/tmp/copied-git/git" "$@"
else
    echo "git binary not found in PATH or /tmp/copied-git/"
    exit 1
fi
