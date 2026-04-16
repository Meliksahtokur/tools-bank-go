#!/bin/bash
# =============================================================================
# Build script for tools-bank-go MCP Server
# =============================================================================
# Usage: ./scripts/build.sh [version]
#        ./scripts/build.sh 1.0.0
#
# Options:
#   version    Optional version string to embed in binary
# =============================================================================

set -e

# Configuration
BINARY_NAME="mcp-server"
BUILD_DIR="./bin"
BUILD_FLAGS="-ldflags \"-s -w\""
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse arguments
VERSION="$1"

echo_step() {
    echo -e "${GREEN}==>${NC} $1"
}

echo_error() {
    echo -e "${RED}ERROR:${NC} $1" >&2
}

echo_warning() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

# Validate environment
validate_env() {
    if ! command -v go &> /dev/null; then
        echo_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    echo_step "Go version: $(go version)"
}

# Build function
do_build() {
    local version="$1"
    local output_dir="$BUILD_DIR"
    local output_file=""
    local ldflags="$BUILD_FLAGS"
    
    # Create output directory
    mkdir -p "$output_dir"
    
    # Determine output filename
    if [ -n "$version" ]; then
        output_file="${BINARY_NAME}-${version}"
        # Only add version ldflags if the main package has these variables
        ldflags="-ldflags \"-s -w\""
        echo_step "Building version: $version (without version embedding)"
    else
        output_file="$BINARY_NAME"
        ldflags="-ldflags \"-s -w\""
        echo_step "Building current version"
    fi
    
    local full_output="${output_dir}/${output_file}"
    
    # Build command
    local build_cmd="go build ${ldflags} -o ${full_output} ./cmd/mcp-server"
    
    echo_step "Running: ${build_cmd}"
    
    eval "$build_cmd"
    
    if [ -f "$full_output" ]; then
        local size=$(du -h "$full_output" | cut -f1)
        echo_step "Build complete: $full_output (${size})"
        
        # Make executable
        chmod +x "$full_output"
    else
        echo_error "Build failed - binary not found"
        exit 1
    fi
}

# Show usage
show_usage() {
    echo "Usage: $0 [version]"
    echo ""
    echo "Build the MCP server binary"
    echo ""
    echo "Arguments:"
    echo "  version    Optional version string (e.g., 1.0.0)"
    echo ""
    echo "Examples:"
    echo "  $0              # Build without version"
    echo "  $0 1.0.0        # Build version 1.0.0"
    echo "  $0 v1.2.3-beta  # Build v1.2.3-beta"
    exit 0
}

# Main
main() {
    echo ""
    echo "========================================"
    echo "  MCP Server Build Script"
    echo "========================================"
    echo ""
    
    # Show help if requested
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_usage
    fi
    
    validate_env
    
    do_build "$VERSION"
    
    echo ""
    echo "========================================"
    echo "  Build Complete!"
    echo "========================================"
    echo ""
}

main "$@"
