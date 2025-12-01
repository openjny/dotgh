#!/bin/bash
# Record demo using asciinema
# Usage: ./record-demo.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/.."

# Ensure mock dotgh is available
export PATH="$SCRIPT_DIR:$PATH"

# Record the demo
asciinema rec --overwrite -c "$SCRIPT_DIR/demo-script.sh" "$SCRIPT_DIR/demo.cast"

# Convert to SVG
svg-term --in "$SCRIPT_DIR/demo.cast" --out assets/demo.svg \
  --window \
  --width 80 \
  --height 20 \
  --padding 10 \
  --term iterm2

echo "Demo recorded: assets/demo.svg"
