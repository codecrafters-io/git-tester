#!/bin/sh
export PYTHONPATH="$(dirname "${BASH_SOURCE[0]}")"
exec python -m app "$@"
