#!/bin/sh

# We need to build the binary and run it from the calling directory
# Because it will be called like `git init` from the test directory
SCRIPT_DIR=$(dirname $0)
CALLING_DIR=$(pwd)

cd $SCRIPT_DIR
go build -o ryan-git ./... > /dev/null 2>&1
cd $CALLING_DIR

exec "$SCRIPT_DIR/ryan-git" "$@"
