#!/bin/sh

if command -v git >/dev/null 2>&1; then
    # In stage5, we don't move the git binary, it's required for the actual test
    # And, the test is logically wrong, before running write-tree, we need to run
    # git add .
    if [ "$1" = "write-tree" ]
        then
           git add .
        fi

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
