#!/bin/bash

# Test script for Go CML renderer
# This script tests the renderer with all example files

set -e

echo "Building Go CML renderer..."
go mod tidy
go build -o cml-renderer .

echo "Creating test output directory..."
mkdir -p test-output

echo "Testing renderer with all example files..."

# Count examples
example_count=0
success_count=0

for example in ../examples/*.cml; do
    if [ -f "$example" ]; then
        example_count=$((example_count + 1))
        example_name=$(basename "$example" .cml)
        output_file="test-output/${example_name}.png"
        
        echo "Testing: $example_name"
        
        if ./cml-renderer "$example" "$output_file"; then
            echo "  [OK] Successfully rendered $example_name"
            success_count=$((success_count + 1))
        else
            echo "  [FAIL] Failed to render $example_name"
            exit 1
        fi
    fi
done

echo ""
echo "Test Results:"
echo "  Total examples: $example_count"
echo "  Successful renders: $success_count"
echo "  Failed renders: $((example_count - success_count))"

if [ $success_count -eq $example_count ]; then
    echo "  [OK] All tests passed!"
    echo ""
    echo "Generated files:"
    ls -la test-output/
    exit 0
else
    echo "  [FAIL] Some tests failed!"
    exit 1
fi
