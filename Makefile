# Makefile for OpenIPC ezconfig

BINARY_NAME := ezconfig
ENTRY_POINT := cmd/api/main.go

# Default target
all: build

# Build for host OS
build:
	@echo "Building for host OS..."
	go build -o $(BINARY_NAME) $(ENTRY_POINT)

# Cross-compile for ARMv7 Linux (OpenIPC Air Unit)
build-arm:
	@echo "Building for ARMv7 Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(BINARY_NAME)-arm $(ENTRY_POINT)
	upx --best --lzma $(BINARY_NAME)-arm

# Run locally
run:
	go run $(ENTRY_POINT)

# Clean up binaries
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME) $(BINARY_NAME)-arm

.PHONY: all build build-arm run clean
