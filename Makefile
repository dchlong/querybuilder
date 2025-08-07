# Makefile for querybuilder

# Build variables
BINARY_NAME=querybuilder
BUILD_DIR=bin
CMD_DIR=cmd/querybuilder
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go variables
GOMOD=$(shell head -1 go.mod | cut -d' ' -f2)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

.PHONY: all build install clean test lint fmt help examples

# Default target
all: build

## Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./$(CMD_DIR)
	@echo "$(BINARY_NAME) installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

## Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

## Lint the code
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## Generate examples
examples:
	@echo "Generating examples..."
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) examples/usage.go; \
		$(BUILD_DIR)/$(BINARY_NAME) examples/product_example.go; \
		echo "Examples generated successfully"; \
	else \
		echo "Binary not found. Run 'make build' first"; \
		exit 1; \
	fi

## Run examples
run-examples: build examples
	@echo "Running example generation..."
	cd examples && go run usage.go

## Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@echo "Development environment ready!"

## Quick development workflow
dev: fmt lint test build

## Release preparation
pre-release: clean fmt lint test build-all
	@echo "Release preparation complete!"

## Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Module:  $(GOMOD)"
	@echo "GOOS:    $(GOOS)"
	@echo "GOARCH:  $(GOARCH)"

## Show help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Usage examples:"
	@echo "  make build                    # Build the binary"
	@echo "  make install                  # Install to GOPATH/bin"
	@echo "  make examples                 # Generate example query builders"
	@echo "  make dev                      # Full development workflow"
	@echo "  make build-all                # Build for all platforms"