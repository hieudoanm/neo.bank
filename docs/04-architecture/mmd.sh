#!/bin/bash
set -e

# Set cache directory for puppeteer
CACHE_DIR="${PUPPETEER_CACHE_DIR:-$HOME/.cache/puppeteer}"

# Check if chrome-headless-shell is installed
shopt -s nullglob
browser_dirs=($CACHE_DIR/chrome-headless-shell-*)
if [ ${#browser_dirs[@]} -eq 0 ]; then
    echo "Chrome headless shell not found in $CACHE_DIR. Installing..."
    pnpx puppeteer browsers install chrome-headless-shell
fi

# Create output directories
mkdir -p svg images

# Convert all mmd files to svg and png
for f in mmd/*.mmd; do
    filename=$(basename "$f" .mmd)
    echo "Converting $filename.mmd to SVG and PNG..."

    # Read the file content
    content=$(cat "$f")
    # Check if the content starts with ```mermaid and ends with ```
    case "$content" in
        \`\`\`mermaid*\`\`\`)
            # Remove first and last line
            content=$(echo "$content" | sed '1d;$d')
            ;;
    esac
    
    # Use a pipe to pass the content to mermaid-cli for SVG
    echo "$content" | pnpx @mermaid-js/mermaid-cli -i /dev/stdin -o "./svg/${filename}.svg"
    
    # Use a pipe to pass the content to mermaid-cli for PNG
    echo "$content" | pnpx @mermaid-js/mermaid-cli -i /dev/stdin -o "./images/${filename}.png"
done
