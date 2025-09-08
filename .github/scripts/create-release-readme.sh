#!/bin/bash

# Check if RELEASE_DIR is set
if [ -z "$RELEASE_DIR" ]; then
    echo "Error: RELEASE_DIR environment variable is not set"
    exit 1
fi

# Create the directory if it doesn't exist
mkdir -p "$RELEASE_DIR"

# Create README for the release
cat > "${RELEASE_DIR}/README.md" << 'EOF'
# CML Renderer Binaries

This archive contains pre-built CML renderer binaries for multiple platforms.

## Usage

### Linux/macOS/Unix
```bash
./cml-renderer-<platform> input.cml output.png
```

### Windows
```cmd
cml-renderer-<platform>.exe input.cml output.png
```

## Available Platforms

- linux-amd64 - Linux x86_64
- linux-arm64 - Linux ARM64
- windows-amd64 - Windows x86_64
- windows-arm64 - Windows ARM64
- darwin-amd64 - macOS Intel
- darwin-arm64 - macOS Apple Silicon
- freebsd-amd64 - FreeBSD x86_64
- openbsd-amd64 - OpenBSD x86_64

## Examples

See the examples/ directory in the repository for sample CML files.

## Build Info

- Commit: ${COMMIT_SHA}
- Build Date: $(date -u +'%Y-%m-%d %H:%M:%S UTC')
- Go Version: $(go version | cut -d' ' -f3)
EOF
