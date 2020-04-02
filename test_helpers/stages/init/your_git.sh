#!/bin/sh
export PYTHONPATH="$(dirname "$0")"
exec python -m app "$@"
