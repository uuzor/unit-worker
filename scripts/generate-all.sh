#!/bin/bash
# ============================================================================
# Generate code for all languages from Protocol Buffers
# ============================================================================
# This script generates code for Go, Python, and Node.js
# Usage: ./generate-all.sh

set -e

echo "🚀 Generating code for all languages..."
echo ""

# Generate Go
if [ -f "scripts/generate-go.sh" ]; then
    ./scripts/generate-go.sh
fi

# Generate Python
if [ -f "scripts/generate-python.sh" ]; then
    ./scripts/generate-python.sh
fi

# Generate Node.js (if script exists)
if [ -f "scripts/generate-nodejs.sh" ]; then
    ./scripts/generate-nodejs.sh
fi

echo "✅ All code generated successfully!"
