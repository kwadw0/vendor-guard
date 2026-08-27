#!/bin/sh
set -e
# Standalone graphify updater - no opencode needed
# Usage: ./scripts/graphify-update.sh [--force]
if ! command -v graphify >/dev/null 2>&1; then
  echo "graphify not found. Install with: pip install graphifyy  or  uv tool install graphifyy"
  exit 1
fi
echo "→ graphify update ."
if [ "$1" = "--force" ]; then
  GRAPHIFY_FORCE=1 graphify update . --force
else
  graphify update .
fi
echo "→ graphify cluster-only ."
graphify cluster-only .
echo "✓ graphify-out/ updated"
