# GoDataCleaner Makefile

# Binary name
BINARY_NAME=godatacleaner

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Build flags
CGO_ENABLED=1
LDFLAGS=-ldflags "-s -w"

# Version info (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: all build run test clean deps vet fmt help \
        build-linux-amd64 build-linux-arm64 \
        build-darwin-amd64 build-darwin-arm64 \
        build-windows-amd64 build-all

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/godatacleaner

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Run tests
test:
	@echo "Running tests..."
	CGO_ENABLED=$(CGO_ENABLED) $(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	CGO_ENABLED=$(CGO_ENABLED) $(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.db

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Cross-compilation targets
# Note: CGO is required for SQLite. Cross-compiling with CGO requires appropriate C cross-compilers.
# For native builds on each platform, use the standard 'make build' command.

# Linux AMD64
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		CC=x86_64-linux-musl-gcc \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/godatacleaner

# Linux ARM64
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) \
		CC=aarch64-linux-musl-gcc \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/godatacleaner

# macOS AMD64 (Intel)
build-darwin-amd64:
	@echo "Building for macOS AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/godatacleaner

# macOS ARM64 (Apple Silicon)
build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/godatacleaner

# Windows AMD64
build-windows-amd64:
	@echo "Building for Windows AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) \
		CC=x86_64-w64-mingw32-gcc \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/godatacleaner

# Build all platforms (requires cross-compilers installed)
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64
	@echo "All builds complete!"

# Build for current platform only
build-native:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/godatacleaner

# Help
help:
	@echo "GoDataCleaner Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build                Build the application (current platform)"
	@echo "  make build-native         Alias for build"
	@echo "  make run                  Build and run the application"
	@echo "  make run ARGS=sync        Run with specific command"
	@echo "  make test                 Run tests with race detector"
	@echo "  make test-coverage        Run tests with coverage report"
	@echo "  make clean                Remove build artifacts"
	@echo "  make deps                 Download and tidy dependencies"
	@echo "  make vet                  Run go vet"
	@echo "  make fmt                  Format code"
	@echo ""
	@echo "Cross-compilation (requires CGO cross-compilers):"
	@echo "  make build-linux-amd64    Build for Linux AMD64"
	@echo "  make build-linux-arm64    Build for Linux ARM64"
	@echo "  make build-darwin-amd64   Build for macOS Intel"
	@echo "  make build-darwin-arm64   Build for macOS Apple Silicon"
	@echo "  make build-windows-amd64  Build for Windows AMD64"
	@echo "  make build-all            Build for all platforms"
	@echo ""
	@echo "  make help                 Show this help"
