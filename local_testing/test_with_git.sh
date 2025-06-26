#!/bin/bash

set -e

# Ensure we're in the correct directory
cd "$(dirname "$0")/.."

# Build and run
docker build -t local-git-tester -f local_testing/Dockerfile .
docker run --rm -it -v $(pwd):/app local-git-tester make test_with_git
