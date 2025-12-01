#!/bin/bash
# Record demo using asciinema
# Usage: ./assets/record-demo.sh (from project root)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Ensure mock dotgh is available
export PATH="$SCRIPT_DIR:$PATH"

# Record the demo
asciinema rec --overwrite -c "$SCRIPT_DIR/demo-script.sh" "$SCRIPT_DIR/demo.cast"

# Convert to SVG
svg-term --in "$SCRIPT_DIR/demo.cast" --out "$SCRIPT_DIR/demo.svg" \
  --window \
  --width 80 \
  --height 24 \
  --padding 10

echo "Demo recorded: $SCRIPT_DIR/demo.svg"
