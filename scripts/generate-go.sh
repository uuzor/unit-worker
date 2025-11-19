#!/bin/bash
# ============================================================================
# Generate Go code from Protocol Buffers
# ============================================================================
# This script generates Go code from .proto files
# Usage: ./generate-go.sh
# Output: generated/go/

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Generating Go code from protos${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${YELLOW}Error: protoc not found${NC}"
    echo "Install with: brew install protobuf (macOS) or apt-get install protobuf-compiler (Linux)"
    exit 1
fi

# Check if Go plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create output directory
OUTPUT_DIR="generated/go"
mkdir -p "$OUTPUT_DIR"

# Proto source directory
PROTO_DIR="protos"

echo -e "${YELLOW}Generating Go code...${NC}"

# Generate for each proto file
find "$PROTO_DIR" -name "*.proto" -print0 | while IFS= read -r -d '' proto_file; do
    echo "  → Processing $proto_file"
    protoc \
        --proto_path="$PROTO_DIR" \
        --go_out="$OUTPUT_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$OUTPUT_DIR" \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"
done

echo ""
echo -e "${GREEN}✓ Go code generated successfully${NC}"
echo -e "Output directory: ${YELLOW}$OUTPUT_DIR${NC}"
echo ""

# Count generated files
go_files=$(find "$OUTPUT_DIR" -name "*.go" | wc -l)
echo -e "Generated ${GREEN}$go_files${NC} Go files"
echo ""
