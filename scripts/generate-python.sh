#!/bin/bash
# ============================================================================
# Generate Python code from Protocol Buffers
# ============================================================================
# This script generates Python code from .proto files
# Usage: ./generate-python.sh
# Output: generated/python/

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Generating Python code from protos${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${YELLOW}Error: protoc not found${NC}"
    echo "Install with: brew install protobuf (macOS) or apt-get install protobuf-compiler (Linux)"
    exit 1
fi

# Check if grpcio-tools is installed
if ! python3 -c "import grpc_tools" 2>/dev/null; then
    echo -e "${YELLOW}Installing grpcio-tools...${NC}"
    pip3 install grpcio-tools
fi

# Create output directory
OUTPUT_DIR="generated/python"
mkdir -p "$OUTPUT_DIR"

# Create __init__.py files
touch "$OUTPUT_DIR/__init__.py"

# Proto source directory
PROTO_DIR="protos"

echo -e "${YELLOW}Generating Python code...${NC}"

# Generate for each proto file
find "$PROTO_DIR" -name "*.proto" -print0 | while IFS= read -r -d '' proto_file; do
    echo "  → Processing $proto_file"
    python3 -m grpc_tools.protoc \
        --proto_path="$PROTO_DIR" \
        --python_out="$OUTPUT_DIR" \
        --grpc_python_out="$OUTPUT_DIR" \
        "$proto_file"
done

# Create package structure
for dir in "$OUTPUT_DIR"/*; do
    if [ -d "$dir" ]; then
        touch "$dir/__init__.py"
    fi
done

echo ""
echo -e "${GREEN}✓ Python code generated successfully${NC}"
echo -e "Output directory: ${YELLOW}$OUTPUT_DIR${NC}"
echo ""

# Count generated files
py_files=$(find "$OUTPUT_DIR" -name "*.py" | wc -l)
echo -e "Generated ${GREEN}$py_files${NC} Python files"
echo ""
