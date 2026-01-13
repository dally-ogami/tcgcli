#!/usr/bin/env bash
set -euo pipefail

OUTPUT_PATH=${1:-tcgcli.framework}

if ! command -v gomobile >/dev/null 2>&1; then
  echo "gomobile not found. Install with: go install golang.org/x/mobile/cmd/gomobile@latest" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "go not found. Please install Go 1.22+." >&2
  exit 1
fi

gomobile init

echo "Building iOS framework to ${OUTPUT_PATH}..."
gomobile bind -target=ios -o "${OUTPUT_PATH}" ./mobile
