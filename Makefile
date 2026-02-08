# Makefile for OpenIPC ezconfig

BINARY_NAME := ezconfig
ENTRY_POINT := cmd/ezconfig/main.go
OUTPUT_DIR := output/

# Default target
all: build-air-unit

# Cross-compile for ARMv7 Linux (OpenIPC Air Unit)
build-air-unit:
	@echo "Building for ARMv7 Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(OUTPUT_DIR)$(BINARY_NAME) $(ENTRY_POINT)
	upx --best --lzma $(OUTPUT_DIR)$(BINARY_NAME)

# Run locally
run:
	go run $(ENTRY_POINT)

# Clean up binaries
clean:
	@echo "Cleaning up..."
	rm -f $(OUTPUT_DIR)$(BINARY_NAME)

.PHONY: all build-air-unit run clean
