#!/bin/sh
go mod tidy
go build -o mygit ./cmd/mygit
exec ./mygit "$@"
