# Makefile for Chart Markup Language project

.PHONY: help test-go test-python test-all clean build-go build-python

help:
	@echo "Available targets:"
	@echo "  test-go      - Test Go renderer with all examples"
	@echo "  test-python  - Test Python renderer with all examples"
	@echo "  test-all     - Test both renderers"
	@echo "  build-go     - Build Go renderer"
	@echo "  build-python - Install Python dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help message"

test-go:
	@echo "Testing Go renderer..."
	cd go-renderer && ./test_examples.sh

test-python:
	@echo "Testing Python renderer..."
	cd python-renderer && python test_examples.py

test-all: test-go test-python
	@echo "All tests completed!"

build-go:
	@echo "Building Go renderer..."
	cd go-renderer && go mod tidy && go build -o cml-renderer .

build-go-all:
	@echo "Building Go renderer for all platforms..."
	cd go-renderer && ./build.sh

build-python:
	@echo "Installing Python dependencies..."
	cd python-renderer && pip install -r requirements.txt

clean:
	@echo "Cleaning build artifacts..."
	rm -rf go-renderer/cml-renderer
	rm -rf go-renderer/test-output
	rm -rf python-renderer/test-output
	rm -rf python-renderer/__pycache__
	rm -rf python-renderer/*.pyc
