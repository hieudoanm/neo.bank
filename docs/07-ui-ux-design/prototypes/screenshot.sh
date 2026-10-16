#!/usr/bin/env bash
#
# Generate screenshots from HTML design files.
#
# This script installs dependencies (if needed), installs the browser (if needed),
# and generates screenshots from all HTML files in the html/ directory.
#
# Usage:
#   ./docs/07-ui-ux-design/screenshot.sh

set -euo pipefail

design_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

cd "${design_dir}"

echo "Installing dependencies..."
uv sync

echo "Installing Chromium browser (if not already installed)..."
uv run playwright install chromium

echo "Generating screenshots..."
uv run screenshot.py
