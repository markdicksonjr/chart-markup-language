#!/bin/bash

# Build script for CML renderer
# This script builds the renderer for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building CML Renderer for multiple platforms...${NC}"

# Get version info
VERSION=${VERSION:-"dev-$(git rev-parse --short HEAD 2>/dev/null || echo 'local')"}
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_REF=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'local')

echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Git Ref: ${GIT_REF}${NC}"

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitRef=${GIT_REF}"

# Create output directory
OUTPUT_DIR="dist"
mkdir -p "${OUTPUT_DIR}"

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "freebsd/amd64"
    "openbsd/amd64"
)

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r goos goarch <<< "$platform"
    
    echo -e "${YELLOW}Building for ${goos}/${goarch}...${NC}"
    
    # Set environment variables
    export GOOS=${goos}
    export GOARCH=${goarch}
    export CGO_ENABLED=0
    
    # Determine output filename
    if [ "${goos}" = "windows" ]; then
        OUTPUT_NAME="cml-renderer-${goos}-${goarch}.exe"
    else
        OUTPUT_NAME="cml-renderer-${goos}-${goarch}"
    fi
    
    # Build the binary
    go build -ldflags "${LDFLAGS}" -o "${OUTPUT_DIR}/${OUTPUT_NAME}" .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}[OK] Built ${OUTPUT_NAME}${NC}"
        
        # Test the binary if it's for the current platform
        if [ "${goos}" = "$(go env GOOS)" ] && [ "${goarch}" = "$(go env GOARCH)" ]; then
            echo -e "${YELLOW}Testing ${OUTPUT_NAME}...${NC}"
            if [ -f "../examples/minimal.cml" ]; then
                mkdir -p test-output
                "./${OUTPUT_DIR}/${OUTPUT_NAME}" "../examples/minimal.cml" "test-output/test-${goos}-${goarch}.png"
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}[OK] Test passed for ${OUTPUT_NAME}${NC}"
                else
                    echo -e "${RED}[FAIL] Test failed for ${OUTPUT_NAME}${NC}"
                fi
            fi
        fi
    else
        echo -e "${RED}[FAIL] Failed to build ${OUTPUT_NAME}${NC}"
        exit 1
    fi
done

echo -e "${GREEN}All builds completed successfully!${NC}"
echo -e "${YELLOW}Binaries are available in the ${OUTPUT_DIR}/ directory:${NC}"
ls -la "${OUTPUT_DIR}/"

# Create a simple README for the distribution
cat > "${OUTPUT_DIR}/README.md" << EOF
# CML Renderer Binaries

This directory contains pre-built CML renderer binaries for multiple platforms.

## Usage

### Linux/macOS/Unix
\`\`\`bash
./cml-renderer-<platform> input.cml output.png
\`\`\`

### Windows
\`\`\`cmd
cml-renderer-<platform>.exe input.cml output.png
\`\`\`

## Available Platforms

- \`linux-amd64\` - Linux x86_64
- \`linux-arm64\` - Linux ARM64
- \`windows-amd64\` - Windows x86_64
- \`windows-arm64\` - Windows ARM64
- \`darwin-amd64\` - macOS Intel
- \`darwin-arm64\` - macOS Apple Silicon
- \`freebsd-amd64\` - FreeBSD x86_64
- \`openbsd-amd64\` - OpenBSD x86_64

## Examples

See the \`examples/\` directory in the repository for sample CML files.

## Build Info

- Version: ${VERSION}
- Build Date: ${BUILD_TIME}
- Git Ref: ${GIT_REF}
EOF

echo -e "${GREEN}Created README.md in ${OUTPUT_DIR}/ directory${NC}"
