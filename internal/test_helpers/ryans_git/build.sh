#!/bin/bash

# Build script for cross-platform compilation

set -e

echo "Building mygit for multiple platforms..."

echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o mygit-darwin-arm64 ./...

echo "Building for Linux (Intel)..."
GOOS=linux GOARCH=amd64 go build -o mygit-linux-amd64 ./...

echo "Building for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -o mygit-linux-arm64 ./...

echo ""
echo "Usage:"
echo "  macOS Apple Silicon: ./mygit-darwin-arm64"
echo "  Linux Intel:        ./mygit-linux-amd64"
echo "  Linux ARM64:        ./mygit-linux-arm64"