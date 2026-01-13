#!/usr/bin/env bash
set -euo pipefail

OUTPUT_PATH=${1:-tcgcli.xcframework}

if ! command -v gomobile >/dev/null 2>&1; then
  echo "gomobile not found. Install with: go install golang.org/x/mobile/cmd/gomobile@latest" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "go not found. Please install Go 1.22+." >&2
  exit 1
fi

gomobile init

if [[ "${OUTPUT_PATH}" != *.xcframework ]]; then
  echo "Output must end with .xcframework (gomobile requires this suffix for iOS)." >&2
  exit 1
fi

echo "Building iOS xcframework to ${OUTPUT_PATH}..."
gomobile bind -target=ios -o "${OUTPUT_PATH}" ./mobile
